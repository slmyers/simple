package main

import "./db"

type UserPayload struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type StatusPayload struct {
	Uid int    `json:"uid"`
	Msg string `json:"msg"`
}

type FollowPayload struct {
	Uid     int `json:"uid"`
	Otherid int `json:"otherid"`
}

type TimelinePayload struct {
	Uid  int `json:"uid"`
	Page int `json:"page"`
}

type UserinfoPayload struct {
	Uid int `json:"uid"`
}

type TimelineResponse struct {
	Uid   int                `json:"uid"`
	Page  int                `json:"page"`
	Posts []myredisDB.Status `json:"posts"`
}
