package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
)

type DB struct {
	// pool of redis connections
	pool *redis.Pool
}

func NewDB(server string) *DB {
	db := new(DB)
	db.pool = newPool("localhost:6379")
	return db
}

// init redis pool
func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 & time.Second,
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

/*
 * create user in redis layer
 */
func (db *DB) CreateUser(login, name string) (int, error) {
	c := db.Get()
	if c == nil {
		fmt.Printf("db is nil\n")
		return -1, nil
	}
	defer c.Close()
	// check if users:<login> hash exists
	existsr, err := c.Do("HEXISTS", "users:", login)
	if err != nil {
		return -1, err
	}
	exists, _ := redis.Int(existsr, nil)
	if exists == 1 {
		return -1, err
	}

	// increment global user count
	idr, err := c.Do("INCR", "user:id")
	if err != nil {
		return -1, err
	}
	id, _ := redis.Int(idr, nil)

	/* using pipeline mode */
	c.Send("MULTI")
	c.Send("HSET", "users:", login, id)
	// set user:<login>:<id> => user:1
	c.Send("HMSET", "user:"+strconv.Itoa(id), "login", login,
		"id", id, "name", name, "followers", "0", "following", "0",
		"posts", "0", "signup", time.Now().Unix())
	if _, err := c.Do("EXEC"); err != nil {
		return -1, err
	}

	return id, nil
}

func (db *DB) DeleteUser(uid int) (bool, error) {
	c := db.Get()
	if c == nil {
		return false, nil
	}
	defer c.Close()

	loginr, err := c.Do("HGET", "user:"+strconv.Itoa(uid), "login")
	if err != nil {
		return false, err
	}
	login, _ := redis.String(loginr, nil)
	// user exists
	if login != "" {
		c.Do("MULTI")
		c.Do("HDEL", "users:", login)
		c.Do("HDEL", "user:"+strconv.Itoa(uid), "login")
		c.Do("HDEL", "user:"+strconv.Itoa(uid), "id")
		c.Do("HDEL", "user:"+strconv.Itoa(uid), "name")
		c.Do("HDEL", "user:"+strconv.Itoa(uid), "followers")
		c.Do("HDEL", "user:"+strconv.Itoa(uid), "following")
		c.Do("HDEL", "user:"+strconv.Itoa(uid), "posts")
		c.Do("HDEL", "user:"+strconv.Itoa(uid), "signup")
		if _, err := c.Do("EXEC"); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (db *DB) GetUser(uid int) (*User, error) {
	var user User
	c := db.Get()
	if c == nil {
		return nil, nil
	}
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

/*
 * create status for user with string content
 */
func createStatus(message string, uid int, c redis.Conn) (int, error) {
	var login string
	var sid int
	defer c.Close()

	c.Do("MULTI")
	// get login name for user with id
	c.Do("HGET", "user:"+strconv.Itoa(uid), "login")
	// get the incremented global article count
	c.Do("INCR", "status:id")
	reply, err := redis.Values(c.Do("EXEC"))

	if err != nil {
		return -1, err
	}

	if _, err := redis.Scan(reply, &login, &sid); err != nil {
		return -1, err
	}
	// set status:<id>
	if _, err := c.Do("HMSET", "status:"+strconv.Itoa(sid), "message", message,
		"posted", time.Now().Unix(), "id", sid, "uid", uid, "login", login); err != nil {
		return -1, err
	}

	if _, err := c.Do("HINCRBY", "user:"+strconv.Itoa(uid), "posts", 1); err != nil {
		return -1, err
	}

	return sid, nil
}

func (db *DB) GetStatus(sid int) (*Status, error) {
	var status Status
	c := db.Get()
	if c == nil {
		fmt.Printf("db is nil\n")
		return nil, nil
	}
	defer c.Close()

	r, err := redis.Values(c.Do("HGETALL", "status:"+strconv.Itoa(sid)))

	if err != nil {
		return nil, err
	}

	if err := redis.ScanStruct(r, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (db *DB) GetUserTimeline(uid, page, count int) ([]int, error) {
	timeline := make([]int, page)
	c := db.Get()
	if c == nil {
		fmt.Printf("db is nil\n")
		return nil, nil
	}
	defer c.Close()

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

func (db *DB) Follow(uid, otherid int) (bool, error) {
	c := db.Get()
	if c == nil {
		fmt.Printf("db is nil\n")
		return false, nil
	}
	defer c.Close()

	fkey1 := "following:" + strconv.Itoa(uid)
	fkey2 := "followers:" + strconv.Itoa(otherid)

	r, err := c.Do("ZSCORE", fkey1, strconv.Itoa(otherid))
	if r != nil {
		return true, err
	}

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

func (db *DB) Unfollow(uid, otherid int) (bool, error) {
	c := db.Get()
	if c == nil {
		fmt.Printf("db is nil\n")
		return false, nil
	}
	defer c.Close()

	fkey1 := "following:" + strconv.Itoa(otherid)
	fkey2 := "followers:" + strconv.Itoa(uid)

	r, err := c.Do("ZSCORE", fkey1, strconv.Itoa(otherid))

	if err != nil {
		fmt.Printf("error checking to see if following.\n")
		return false, err
	}

	if r == nil {
		return true, err
	}

	c.Do("MULTI")
	c.Do("ZREM", fkey1, strconv.Itoa(otherid))
	c.Do("ZREM", fkey2, strconv.Itoa(uid))
	c.Do("HINCRBY", "user:"+strconv.Itoa(uid), "following", "-1")
	c.Do("HINCRBY", "user:"+strconv.Itoa(otherid), "followers", "-1")
	if _, err := c.Do("EXEC"); err != nil {
		fmt.Printf("error in MULTI.\n")
		return false, err
	}

	return true, nil
}

func (db *DB) PostStatus(uid int, message string) (int, error) {
	c := db.Get()
	if c == nil {
		fmt.Printf("db is nil\n")
		return -1, nil
	}
	defer c.Close()

	sid, err := createStatus(message, uid, db.Get())
	if err != nil {
		return -1, err
	}

	if sid == -1 {
		return -1, nil
	}

	timer, err := c.Do("HGET", "status:"+strconv.Itoa(sid), "posted")
	if err != nil {
		return -1, err
	}

	time, _ := redis.Int(timer, nil)

	if _, err := c.Do("ZADD", "timeline:"+strconv.Itoa(uid), time, sid); err != nil {
		return -1, err
	}
	/*
		arg1 := strconv.Atoi(sid)
		arg2 := strconv.Atoid(time)
	*/
	succ, err := syndicateStatus(strconv.Itoa(uid), strconv.Itoa(sid), strconv.Itoa(time),
		db.Get())

	if succ != true || err != nil {
		return -1, err
	}

	return sid, nil
}

func syndicateStatus(uid, sid, time string, c redis.Conn) (bool, error) {
	defer c.Close()
	r, err := redis.Values(c.Do("ZRANGEBYSCORE", "followers:"+uid, "-inf",
		"+inf"))

	if err != nil {
		return false, err
	}

	var followerSlice []string
	if err := redis.ScanSlice(r, &followerSlice); err != nil {
		return false, nil
	}
	c.Do("MULTI")
	for i := range followerSlice {
		c.Do("ZADD", "timeline:"+followerSlice[i], time, sid)
	}
	if _, err := c.Do("EXEC"); err != nil {
		return false, err
	}
	return true, nil
}
