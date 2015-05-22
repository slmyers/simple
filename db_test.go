package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"testing"
)

func TestCreateUser(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	c := db.Get()
	defer c.Close()
	uid, err := db.CreateUser("TestUser", "testy")

	if err != nil {
		t.Error("recieved error while creating user: ", err)
	}

	if uid == -1 {
		t.Error("uid for new user is: ", uid)
	}

	// check to see if user is actually created
	exists, err := redis.Int(c.Do("HEXISTS", "users:", "TestUser"))

	if err != nil || exists != 1 {
		t.Error("TestUser is not successfully created.\n")
	}

	// test delete user
	r, err := db.DeleteUser(uid)

	if err != nil {
		t.Error("recieved error while deleting user: ", err)
	}

	if r != true {
		t.Error("expected true when delete user got: ", r)
	}
	// double check
	exists, err = redis.Int(c.Do("HEXISTS", "users:", "TestUser"))

	if err != nil || exists != 0 {
		t.Error("TestUser was not successfully deleted.\n")
	}
}

func TestFollow(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	c := db.Get()
	defer c.Close()

	if res, err := db.Follow(54, 55); res == false || err != nil {
		t.Error("error user 1 following user 2: ", uids[0], uids[1])
	}
	// double check it actually worked
	if res, err := redis.Int(c.Do("ZSCORE", "followers:55", "54")); res == nil || err != nil {
		t.Error("followers not updated properly")
	}
	if res, err := redis.Int(c.Do("ZSCORE", "following:54", "55")); res == nil || err != nil {
		t.Error("following not updated properly")
	}

	// test unfollow
	if res, err := db.UnFollow(54, 55); res == false || err != nil {
		t.Error("error unfollowing")
	}

	// check to see if worked
	if _, err := c.Do("ZSCORE", "followers:55", "54"); err != nil {
		t.Error("error checking to see if unfollowed")
	}

	if _, err := c.Do("ZSCORE", "following:54", "55"); err != nil {
		t.Error("error checking to see if unfollowing")
	}
}
