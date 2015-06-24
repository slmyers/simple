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

type TimelineHeader struct {
	Id      string `json:"id"`
	Login   string `json:"login"`
	Page    int    `json:"page"`
	PostIds []int  `json:"posts"`
}

type Posts struct {
	Posts Statuses `json:"posts"`
}

type Statuses []myredisDB.Status

func (s Statuses) Len() int      { return len(s) }
func (s Statuses) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// we want the "larger" time to be first in sorting order
func (s Statuses) Less(i, j int) bool { return s[i].Posted > s[j].Posted }
