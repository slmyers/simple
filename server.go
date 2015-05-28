package main

import (
	rdb "./db"
	"github.com/slmyers/go-json-rest/rest"
	"log"
	"net/http"
)

func main() {
	i := Impl{}
	i.InitDB()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		rest.Post("/createuser", i.CreateUser),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8000", api.MakeHandler()))
}

type Impl struct {
	DB *rdb.DB
}

func (i *Impl) InitDB() {
	i.DB = rdb.NewDB("localhost:6379")
	if i.DB == nil {
		log.Fatal("i.DB is nil")
	}
}

type Userpayload struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

func (i *Impl) CreateUser(w rest.ResponseWriter, r *rest.Request) {
	user := new(Userpayload)
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if uid, err := i.DB.CreateUser(user.Username, user.Name); uid != -1 || err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
