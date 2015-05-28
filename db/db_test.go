package myredisDB

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
	// save the original global user count
	oldGlobalID, err := redis.String(c.Do("GET", "user:id"))
	// create a test user using the API
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

	// reset the global user count
	c.Do("SET", "user:id", oldGlobalID)

}

func TestFollow(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	c := db.Get()
	defer c.Close()

	if res, err := db.Follow(-1, -2); res == false || err != nil {
		t.Error("error user -1 following user -2")
	}
	// double check it actually worked
	if res, err := redis.String(c.Do("ZSCORE", "followers:-2", "-1")); res == "" || err != nil {
		t.Error("followers not updated properly")
	}
	if res, err := redis.String(c.Do("ZSCORE", "following:-1", "-2")); res == "" || err != nil {
		t.Error("following not updated properly")
	}

	// test unfollow
	if res, err := db.Unfollow(-1, -2); res == false || err != nil {
		t.Error("error unfollowing ", err)
	}

	// check to see if worked
	if res, err := c.Do("ZSCORE", "followers:-2", "-1"); err != nil || res == nil {
		t.Error("error checking to see if unfollowed")
	}

	if res, err := c.Do("ZSCORE", "following:-1", "-2"); err != nil || res == nil {
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
	// get the old status id
	oldSID, _ := redis.String(c.Do("GET", "status:id"))

	if res, err := db.Follow(-1, -2); res == false || err != nil {
		t.Error("error user: -1 following user: -2")
	}

	sid, err := db.PostStatus(-2, "this is a post")
	if sid == -1 || err != nil {
		t.Error("error user: -2 posting status ", err)
	}

	res, err := redis.String(c.Do("ZSCORE", "timeline:-2", sid))

	if err != nil {
		t.Error("error checking user: -2 timeline ", err)
	}

	fmt.Printf("res = %v\n", res)

	if res, err := redis.String(c.Do("ZSCORE", "timeline:-1", sid)); err != nil || res == "" {
		t.Error("error checking user: -1 timeline ", err)
	}

	res2, err := db.Unfollow(-1, -2)
	/*
		{
			t.Error("unable to unfollow ", err)
		}
	*/
	fmt.Printf("res for unfollow = %v\n", res2)

	c.Do("SET", "status:id", oldSID)
}
