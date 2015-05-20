package main

type User struct {
	Login     string `redis:"login"`
	Id        int    `redis:"id"`
	Name      string `redis:"name"`
	Followers int    `redis:"followers"`
	Following int    `redis:"following"`
	Posts     int    `redis:"posts"`
	Signup    int64  `redis:"signup"`
}

type Status struct {
	Message string `redis:"message"`
	Posted  int64  `redis:"posted"`
	Id      int    `redis:"id"`
	Uid     int    `redis:"uid"`
	Login   string `redis:"login"`
}

type Timeline struct {
	Timestamp int64 `redis:"timestamp"`
	Status    int   `redis:"article"`
}

type Follow struct {
	Timestamp int64 `redis:"timestamp"`
	User      int   `redis:"user"`
}

type Follower struct {
	Timestamp int64 `redis:"timestamp"`
	User      int   `redis:"user"`
}

type Timelines []Timeline
type Following []Following
type Followers []Follower
type Statuses []Status
type Users []User
