package myredisDB

import (
	"github.com/garyburd/redigo/redis"
	"strconv"
	"testing"
)

/*
 * Tests creating and deleting a user
 */
func TestUser(t *testing.T) {
	db := NewDB("localhost:6379")
	c := db.Get()
	defer c.Close()
	// save the original global user count
	oldGlobalID, err := redis.String(c.Do("GET", "user:id"))
	// create a test user using the API
	uid, err := db.CreateUser("TestUser", "testy")

	if err != nil {
		t.Error("recieved error while creating user: ", err)
	}
	// -1 uid indicates that the user was not created
	if uid == -1 {
		t.Errorf("new user not created. uid: %v\n", uid)
	}

	// check to see if newly created username is registered
	exists, err := redis.Int(c.Do("HEXISTS", "users:", "TestUser"))
	// redis returns 1 if the hash exists
	if err != nil || exists != 1 {
		t.Error("TestUser is not successfully created.\n")
	}

	// check if the user hash structure is created
	res, err := redis.Values(c.Do("HGETALL", "user:"+strconv.Itoa(uid)))
	// scan the values into a struct so we can inspect
	var createdUser User
	err = redis.ScanStruct(res, &createdUser)

	if err != nil {
		t.Error("error scanning user struct.\n")
	}
	if createdUser.Id != uid {
		t.Errorf("createdUser.Id == %v\tuid == %v\n", createdUser.Id,
			uid)
	}

	if createdUser.Login != "TestUser" {
		t.Errorf("createdUser.Login == %v\tlogin == TestUser\n",
			createdUser.Login)
	}

	if createdUser.Name != "testy" {
		t.Errorf("createdUser.Name == %v\tname == testy\n",
			createdUser.Name)
	}

	if createdUser.Posts != 0 && createdUser.Followers != 0 &&
		createdUser.Following != 0 {
		t.Errorf("Posts, following and followers are not 0.\n")
	}
	// test delete user
	r, err := db.DeleteUser(uid)

	if err != nil {
		t.Error("recieved error while deleting user: ", err)
	}

	if r != true {
		t.Error("expected true when delete user got: ", r)
	}
	// check that username is no longer registered
	exists, err = redis.Int(c.Do("HEXISTS", "users:", "TestUser"))

	if err != nil || exists != 0 {
		t.Error("TestUser was not successfully deleted.\n")
	}
	// check if the hash structure is deleted
	getAll, err := redis.Values(c.Do("HGETALL", "user:"+strconv.Itoa(uid)))

	if len(getAll) != 0 {
		t.Errorf("Error deleting user\tgetAll == %v\n", getAll)
	}
	// reset the global user count
	c.Do("SET", "user:id", oldGlobalID)

}

// tests follow and unfollow methods
func TestFollow(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	c := db.Get()
	defer c.Close()
	// use two uids that are invalid, so as to not mess with application data
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
	if res, err := c.Do("ZSCORE", "followers:-2", "-1"); err != nil || res != nil {
		t.Error("error checking to see if unfollowed")
	}

	if res, err := c.Do("ZSCORE", "following:-1", "-2"); err != nil || res != nil {
		t.Error("error checking to see if unfollowing")
	}

	c.Do("DEL", "following:-1")
	c.Do("DEL", "followers:-2")
}

// tests to see if posts appear on the users's timeline and follower's timeline
func TestPost(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	c := db.Get()
	defer c.Close()
	// get the oldsid so we can reset value in database
	// we don't want test code to interfere with the application data
	oldSID, _ := redis.String(c.Do("GET", "status:id"))

	// user -1 follows user -2. User -2 makes a post.

	if res, err := db.Follow(-1, -2); res == false || err != nil {
		t.Error("error user: -1 following user: -2")
	}

	sid, err := db.PostStatus(-2, "this is a post")
	if sid == -1 || err != nil {
		t.Error("error user: -2 posting status ", err)
	}
	// check to see if post showed up on both timelines
	if res, err := redis.String(c.Do("ZSCORE", "timeline:-2", sid)); err != nil || res == "" {
		t.Error("error checking user: -2 timeline ", err)
	}

	if res, err := redis.String(c.Do("ZSCORE", "timeline:-1", sid)); err != nil || res == "" {
		t.Error("error checking user: -1 timeline ", err)
	}
	// unfollow the users
	if _, err := db.Unfollow(-1, -2); err != nil {
		t.Error("unable to unfollow ", err)
	}

	// remove the timelines
	c.Do("DEL", "timeline:-1")
	c.Do("DEL", "timeline:-2")

	c.Do("SET", "status:id", oldSID)
}
