package main

import (
	rdb "./db"
	"github.com/slmyers/go-json-rest/rest"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	i := Impl{}
	i.InitDB()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		rest.Post("/createuser", i.CreateUser),
		rest.Post("/poststatus", i.PostStatus),
		rest.Post("/follow", i.FollowUser),
		rest.Post("/unfollow", i.UnfollowUser),
		rest.Get("/timeline", i.GetTimeline),
		rest.Get("/user", i.GetUser),
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
	var status Statuspayload
	if err := r.DecodeJsonPayload(&status); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sid, err := i.DB.PostStatus(status.Uid, status.Msg)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(map[string]int{"sid": sid})
}

func (i *Impl) FollowUser(w rest.ResponseWriter, r *rest.Request) {
	var follow Followpayload
	if err := r.DecodeJsonPayload(&follow); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := i.DB.Follow(follow.Uid, follow.Otherid)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(map[string]string{"following": folllow.Otherid,
		"follower": follow.Uid})
}

func (i *Impl) UnfollowUser(w rest.ResponseWriter, r *rest.Request) {
	var follow Followpayload
	if err := r.DecodeJsonPayload(&follow); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := i.DB.Unfollow(follow.Uid, follow.Otherid)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(map[string]string{"unfollowed": "true"})

}

func (i *Impl) GetTimeline(w rest.ResponseWriter, r *rest.Request) {
	var timeline TimelineRequest
	if err := r.DecodeJsonPayload(&timeline); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := i.DB.GetUserTimeline(timeline.Uid, timeline.Page, 30)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output := new(TimelineResponse)
	output.Posts = make([]Status, 30)
	outputIndex := 0
	// channel to send/recieve status structs
	statuses := make(chan Status)

	for pst := range res {
		// anon goroutine to get a status in timeline page
		go func(post int) {
			Status, err := i.DB.GetStatus(post)
			if err != nil {
				log.Printf("error getting post %d, %v\n", post, err)
				return
			}
			statuses <- Status
		}(pst)
	}

	for outputIndex < len(res) {
		select {
		case sts := <-statuses:
			output.Post[outputIndex] = sts
			outputIndex++
		case <-time.After(time.Second * 1):
			log.Printf("timeout getting timeline:%d page:%d\n", timeline.Uid,
				timeline.Page)
			break
		}
	}
	w.WriteJson(&output)
}

func (i *Impl) GetUser(w rest.ResponseWriter, r *rest.Request) {

}
