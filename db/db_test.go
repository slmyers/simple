package db

import (
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
		t.Error("error user 54 following user 55")
	}
	// double check it actually worked
	if res, err := redis.String(c.Do("ZSCORE", "followers:55", "54")); res == "" || err != nil {
		t.Error("followers not updated properly")
	}
	if res, err := redis.String(c.Do("ZSCORE", "following:54", "55")); res == "" || err != nil {
		t.Error("following not updated properly")
	}

	// test unfollow
	if res, err := db.Unfollow(54, 55); res == false || err != nil {
		t.Error("error unfollowing ", err)
	}

	// check to see if worked
	if res, err := c.Do("ZSCORE", "followers:55", "54"); err != nil || res == nil {
		t.Error("error checking to see if unfollowed")
	}

	if res, err := c.Do("ZSCORE", "following:54", "55"); err != nil || res == nil {
		t.Error("error checking to see if unfollowing")
	}
}

func TestPost(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	c := db.Get()
	defer c.Close()

	if res, err := db.Follow(100, 101); res == false || err != nil {
		t.Error("error user: 100 following user: 101")
	}

	sid, err := db.PostStatus(101, "this is a post")
	if sid == -1 || err != nil {
		t.Error("error user: 101 posting status ", err)
	}

	if res, err := redis.String(c.Do("ZSCORE", "timeline:101", sid)); err != nil || res == "" {
		t.Error("error checking user: 101 timeline ", err)
	}

	if res, err := redis.String(c.Do("ZSCORE", "timeline:100", sid)); err != nil || res == "" {
		t.Error("error checking user: 100 timeline ", err)
	}

	if _, err := db.Unfollow(100, 101); err != nil {
		t.Error("unable to unfollow ", err)
	}

}

func TestTimeline(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}

	if res, err := db.GetUserTimeline(100, 1, 30); err != nil {
		t.Error("error pulling timeline", res, err)
	}
}
