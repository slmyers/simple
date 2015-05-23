package db

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
