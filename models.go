package main

type User struct {
	login     string `redis:"login"`
	uid       int    `redis:id`
	name      string `redis:"name"`
	followers int    `redis:followers`
	following int    `redis:following`
	posts     int    `redis:posts`
	signup    int64  `redis:signup`
}

type Status struct {
	message string `redis:"message"`
	posted  int64  `redis:"posted"`
	id      int    `redis:"id"`
	uid     int    `redis:"uid"`
	login   string `redis:"login"`
}

type Timeline struct {
	timestamp int64 `redis:"timestamp"`
	status    int   `redis:"article"`
}

type Follow struct {
	timestamp int64 `redis:"timestamp"`
	user      int   `redis:"user"`
}

type Follower struct {
	timestamp int64 `redis:"timestamp"`
	user      int   `redis:"user"`
}

type Timelines []Timeline
type Following []Following
type Followers []Follower
type Statuses []Status
type Users []User
