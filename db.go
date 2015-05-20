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
func (db *DB) CreateUser(login, name string) (bool, error) {
	c := db.Get()
	if c == nil {
		fmt.Printf("db is nil\n")
	}
	defer c.Close()
	// check if users:<login> hash exists
	existsr, err := c.Do("HEXISTS", "users:", login)
	if err != nil {
		return false, err
	}
	exists, _ := redis.Int(existsr, nil)
	fmt.Printf("exists is: %v\n", exists)
	if exists == 1 {
		return false, err
	}

	// increment global user count
	idr, err := c.Do("INCR", "user:id")
	if err != nil {
		return false, err
	}
	id, _ := redis.Int(idr, nil)
	fmt.Printf("id is: %v\n", id)

	/* using pipeline mode */
	c.Send("MULTI")
	c.Send("HSET", "users:", login, id)
	// set user:<login>:<id> => user:slmyers:1
	c.Send("HMSET", "user:"+strconv.Itoa(id), "login", login,
		"id", id, "name", name, "followers", "0", "following", "0",
		"posts", "0", "signup", time.Now().Unix())
	if _, err := c.Do("EXEC"); err != nil {
		return false, err
	}

	return true, nil
}

/*
 * create article for user with string content
 */
func (db *DB) CreateStatus(message string, uid int) (bool, error) {
	var login string
	var sid int
	c := db.Get()
	defer c.Close()

	c.Do("MULTI")
	// get login name for user with id
	c.Do("HGET", "user:"+strconv.Itoa(uid), "login")
	// get the incremented global article count
	c.Do("INCR", "status:id")
	reply, err := redis.Values(c.Do("EXEC"))

	if err != nil {
		return false, err
	}

	if _, err := redis.Scan(reply, &login, &sid); err != nil {
		return false, err
	}

	fmt.Printf("login = %v\n", login)
	fmt.Printf("sid = %v\n", sid)

	// set status:<id>
	if _, err := c.Do("HMSET", "status:"+strconv.Itoa(sid), "message", message,
		"posted", time.Now().Unix(), "id", sid, "uid", uid, "login", login); err != nil {
		return false, err
	}

	return true, nil
}
