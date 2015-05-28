package main

import (
	rdb "./db"
	"github.com/slmyers/go-json-rest/rest"
	"log"
	"net/http"
	"strconv"
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

func (i *Impl) CreateUser(w rest.ResponseWriter, r *rest.Request) {
	var user Userpayload
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	uid, err := i.DB.CreateUser(user.Username, user.Name)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// server responds with uid of newly created user
	// -1 if username already exists
	w.WriteJson(map[string]string{"uid": strconv.Itoa(uid)})
}

func (i *Impl) PostStatus(w rest.ResponseWriter, r *rest.Request) {

}
