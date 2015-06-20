package myredisDB

type User struct {
	Login     string `redis:"login" json:"login"`
	Id        int    `redis:"id" json:"id"`
	Name      string `redis:"name" json:"name"`
	Followers int    `redis:"followers" json:"followers"`
	Following int    `redis:"following" json:"following"`
	Posts     int    `redis:"posts" json:"posts"`
	Signup    int64  `redis:"signup" json:"signup"`
}

type Status struct {
	Message string `redis:"message" json:"message"`
	Posted  int64  `redis:"posted" json:"posted"`
	Id      int    `redis:"id" json:"id"`
	Uid     int    `redis:"uid" json:"uid"`
	Login   string `redis:"login" json:"login"`
}
