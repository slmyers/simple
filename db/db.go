package myredisDB

import (
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
)

/*******************************************
************** Fields *********************/

type DB struct {
	// pool of redis connections
	pool *redis.Pool
}

/********************************************
************* Constructor ******************/

func NewDB(server string) *DB {
	db := new(DB)
	// default port for redis server
	db.pool = newPool("localhost:6379")
	return db
}

/*********************************************
************ Connection Code ****************/

// init redis pool
func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		MaxActive:   1000, // limit to 1000 active users
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (db *DB) Get() redis.Conn {
	return db.pool.Get()
}

/********************************************
***************  User code *****************/

func (db *DB) CreateUser(login, name string) (int, error) {
	// allocate a connection
	c := db.Get()
	defer c.Close()
	// check if username is taken
	if exists, err := redis.Int(c.Do("HEXISTS", "users:", login)); err != nil || exists == 1 {
		return -1, err
	}

	// increment global user count
	id, err := redis.Int(c.Do("INCR", "user:id"))
	if err != nil {
		return -1, err
	}

	// we want to do a transaction so we use MULTI cmd1 cmd2 ... EXEC
	c.Do("MULTI")
	// register the username to the uid
	c.Do("HSET", "users:", login, id)
	// set fields for user structure
	c.Do("HMSET", "user:"+strconv.Itoa(id), "login", login,
		"id", id, "name", name, "followers", "0", "following", "0",
		"posts", "0", "signup", time.Now().Unix())
	if _, err := c.Do("EXEC"); err != nil {
		return -1, err
	}

	return id, nil
}

func (db *DB) DeleteUser(uid int) (bool, error) {
	c := db.Get()
	defer c.Close()
	// get the users login value so we can remove it from global store
	login, err := redis.String(c.Do("HGET", "user:"+strconv.Itoa(uid), "login"))
	if err != nil {
		return false, err
	}
	// user exists
	if login != "" {
		c.Do("MULTI")
		// remove username from global store
		c.Do("HDEL", "users:", login)
		// delete the key to the hash structure
		c.Do("DEL", "user:"+strconv.Itoa(uid))
		if _, err := c.Do("EXEC"); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (db *DB) GetUser(uid int) (*User, error) {
	var user User
	c := db.Get()
	defer c.Close()

	r, err := redis.Values(c.Do("HGETALL", "user:"+strconv.Itoa(uid)))
	if err != nil {
		return nil, err
	}

	if err := redis.ScanStruct(r, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

/********************************************
*************** Status code ****************/

// creates a status hash structure and returns it's status id.
func createStatus(message string, uid int, c redis.Conn) (int, error) {
	var login string
	var sid int
	defer c.Close()

	c.Do("MULTI")
	// get login name for user with id
	c.Do("HGET", "user:"+strconv.Itoa(uid), "login")
	// get the incremented global status count
	c.Do("INCR", "status:id")
	// reply contains both the login for the uid and the global status count
	reply, err := redis.Values(c.Do("EXEC"))

	if err != nil {
		return -1, err
	}
	// scan reply into the local variables
	if _, err := redis.Scan(reply, &login, &sid); err != nil {
		return -1, err
	}
	// set all the appropriate values in the hash store
	if _, err := c.Do("HMSET", "status:"+strconv.Itoa(sid), "message", message,
		"posted", time.Now().Unix(), "id", sid, "uid", uid, "login", login); err != nil {
		return -1, err
	}
	// increment the user's post count
	if _, err := c.Do("HINCRBY", "user:"+strconv.Itoa(uid), "posts", 1); err != nil {
		return -1, err
	}

	return sid, nil
}

// simple function to fetch a status hash
func (db *DB) GetStatus(sid int) (Status, error) {
	var status Status
	c := db.Get()
	defer c.Close()

	r, err := redis.Values(c.Do("HGETALL", "status:"+strconv.Itoa(sid)))

	if err != nil {
		return status, err
	}

	if err := redis.ScanStruct(r, &status); err != nil {
		return status, err
	}

	return status, nil
}

/*
	exported function for posting a user's status
	will call createStatus(...) and syndicateStatus(...)
*/
func (db *DB) PostStatus(uid int, message string) (int, error) {
	c := db.Get()
	defer c.Close()
	// create the status hash structure
	sid, err := createStatus(message, uid, db.Get())
	if err != nil {
		return -1, err
	}
	// check to see if the status is actually created before sharing to timelines
	if sid == -1 {
		return -1, nil
	}
	// the time of the post is the score for the key in the sorted set
	time, err := redis.Int(c.Do("HGET", "status:"+strconv.Itoa(sid), "posted"))

	if err != nil {
		return -1, err
	}
	// add the post to the user's timeline
	if _, err := c.Do("ZADD", "timeline:"+strconv.Itoa(uid), time, sid); err != nil {
		return -1, err
	}
	// push the status to the follower's timelines
	succ, err := syndicateStatus(strconv.Itoa(uid), strconv.Itoa(sid), strconv.Itoa(time),
		db.Get())

	if succ != true || err != nil {
		return -1, err
	}
	// return the status id of published status
	return sid, nil
}

// adds a status to a user's follower's timelines
func syndicateStatus(uid, sid, time string, c redis.Conn) (bool, error) {
	defer c.Close()
	// get all the followers user ids sorted by score
	// if this is exceptionally large we may want to do the syndication in stages
	r, err := redis.Values(c.Do("ZRANGEBYSCORE", "followers:"+uid, "-inf",
		"+inf"))

	if err != nil {
		return false, err
	}

	var followerSlice []string
	if err := redis.ScanSlice(r, &followerSlice); err != nil {
		return false, nil
	}
	// push the status in a single transaction
	c.Do("MULTI")
	for i := range followerSlice {
		c.Do("ZADD", "timeline:"+followerSlice[i], time, sid)
	}
	if _, err := c.Do("EXEC"); err != nil {
		return false, err
	}
	return true, nil
}

/*******************************************
************ Timeline code ****************/

func (db *DB) GetUserTimeline(uid, page, count int) ([]int, error) {
	timeline := make([]int, page)
	c := db.Get()
	defer c.Close()
	// we use the page and count values to grab unique chunks of the timeline
	r, err := redis.Values(c.Do("ZREVRANGE", "timeline:"+strconv.Itoa(uid),
		strconv.Itoa((page-1)*count), strconv.Itoa(page*(count-1))))

	if err != nil {
		return nil, err
	}

	if err := redis.ScanSlice(r, &timeline); err != nil {
		panic(err)
	}

	return timeline, nil
}

/********************************************
************* Follow code ******************/

// user A follows user B
func (db *DB) Follow(uid, otherid int) (bool, error) {
	c := db.Get()
	defer c.Close()

	fkey1 := "following:" + strconv.Itoa(uid)
	fkey2 := "followers:" + strconv.Itoa(otherid)

	// check to see if user A is following user B already
	r, err := c.Do("ZSCORE", fkey1, strconv.Itoa(otherid))
	if r != nil {
		return true, err
	}
	// add the user ids to the appropriate sets
	// increment the appropriate following and follower hashes
	c.Do("MULTI")
	c.Do("ZADD", fkey1, time.Now().Unix(), strconv.Itoa(otherid))
	c.Do("ZADD", fkey2, time.Now().Unix(), strconv.Itoa(uid))
	c.Do("HINCRBY", "user:"+strconv.Itoa(uid), "following", "1")
	c.Do("HINCRBY", "user:"+strconv.Itoa(otherid), "followers", 1)
	if _, err := c.Do("EXEC"); err != nil {
		return false, err
	}

	return true, nil
}

// user A unfollows user B
func (db *DB) Unfollow(uid, otherid int) (bool, error) {
	c := db.Get()
	defer c.Close()

	fkey1 := "following:" + strconv.Itoa(uid)
	fkey2 := "followers:" + strconv.Itoa(otherid)

	// checking following status
	r, err := c.Do("ZSCORE", fkey1, strconv.Itoa(otherid))
	if r == nil {
		return true, err
	}

	c.Do("MULTI")
	c.Do("ZREM", fkey1, strconv.Itoa(otherid))
	c.Do("ZREM", fkey2, strconv.Itoa(uid))
	c.Do("HINCRBY", "user:"+strconv.Itoa(uid), "following", "-1")
	c.Do("HINCRBY", "user:"+strconv.Itoa(otherid), "followers", "-1")
	if _, err := c.Do("EXEC"); err != nil {
		return false, err
	}
	return true, nil
}
