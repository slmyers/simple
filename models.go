package main

type Userpayload struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type Statuspayload struct {
	Uid string `json:"uid"`
	Msg string `json:"msg"`
}

type Followpayload struct {
	Uid     string `json:"uid"`
	Otherid string `json:"otherid"`
}

type Timelinepayload struct {
	Uid  string `json:"uid"`
	Page string `json:"page"`
}
