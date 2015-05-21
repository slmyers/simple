package main

import (
	"fmt"
	"testing"
)

func TestPostUser(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	uid, err := db.CreateUser("TestUser", "testy")

	if err != nil {
		t.Error("recieved error while creating user: ", err)
	}

	if uid == -1 {
		t.Error("uid for new user is: ", uid)
	}

	r, err := db.DeleteUser(uid)

	if err != nil {
		t.Error("recieved error while deleting user: ", err)
	}

	if r != true {
		t.Error("expected true when delete user got: ", r)
	}

}

func TestPostStatus(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}

	sid, err := db.CreateStatus("this is a status", 10)

	if err != nil {
		t.Error("recieved error: ", err)
	}

	if sid == -1 {
		t.Error("expected sid > 1 got: ", sid)
	}
}

func TestGetUser(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	user, err := db.GetUser(10)

	if err != nil {
		t.Error("recieved error: ", err)
	}

	if user == nil {
		t.Error("user is nil")
	}

	fmt.Printf("user = %v\n", user)
}

func TestGetStatus(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	status, err := db.GetStatus(9)

	if err != nil {
		t.Error("recieved error: ", err)
	}

	if status == nil {
		t.Error("status is nil")
	}

	fmt.Printf("status = %v\n", status)
}
