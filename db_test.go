package main

import "testing"

func TestPostUser(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}
	r, err := db.CreateUser("inthe6", "Henry")

	if err != nil {
		t.Error("recieved error: ", err)
	}

	if r != true {
		t.Error("expected true got: ", r)
	}

}

func TestPostStatus(t *testing.T) {
	db := NewDB("localhost:6379")
	if db == nil {
		t.Error("db is nil")
	}

	r, err := db.CreateStatus("this is a status", 9)

	if err != nil {
		t.Error("recieved error: ", err)
	}

	if r != true {
		t.Error("expected true got: ", r)
	}

}
