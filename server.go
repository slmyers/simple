package main

import (
	rdb "./db"
	"github.com/slmyers/go-json-rest/rest"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

func main() {
	i := Impl{}
	i.InitDB()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := rest.MakeRouter(
		rest.Post("/user", i.CreateUser),
		rest.Post("/status", i.PostStatus),
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

/*
 *	consumes JSON of the form:
 *	{
 *		"username":	"<username>",
 *		"name":	"<users' name>"
 *	}
 *
 *
 *
 *
 *
 */

func (i *Impl) CreateUser(w rest.ResponseWriter, r *rest.Request) {
	var user UserPayload
	err := r.DecodeJsonPayload(&user)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// returns -1 if username is unable to be registered, ie, already taken
	uid, err := i.DB.CreateUser(user.Username, user.Name)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if uid == -1 {
		w.WriteJson(map[string]string{
			"uid":  strconv.Itoa(uid),
			"user": "unable to create",
		})
		return
	}

	user, err := i.DB.GetUser(uid)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&user)
}

/*
 * consumes JSON of the form
   {
		"uid": <user id>
		"msg": <text string containing message>
   }
*/
func (i *Impl) PostStatus(w rest.ResponseWriter, r *rest.Request) {
	var status StatusPayload
	if err := r.DecodeJsonPayload(&status); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sid, err := i.DB.PostStatus(status.Uid, status.Msg)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post, err := i.DB.GetStatus(status.Uid)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(&post)
}

func (i *Impl) FollowUser(w rest.ResponseWriter, r *rest.Request) {
	v, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uid, err := strconv.Atoi(v.Get("uid"))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	otherId, err := strconv.Atoi(v.Get("otherId"))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := i.DB.Follow(uid, otherId)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if res == true {
		w.WriteJson(map[string]string{"following": v.Get("otherId"),
			"follower": v.Get("uid"), "followed": "true"})
	} else {
		w.WriteJson(map[string]string{"following": v.Get("otherId"),
			"follower": v.Get("uid"), "followed": "false"})
	}
}

func (i *Impl) UnfollowUser(w rest.ResponseWriter, r *rest.Request) {
	v, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uid, err := strconv.Atoi(v.Get("uid"))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	otherId, err := strconv.Atoi(v.Get("otherId"))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := i.DB.Unfollow(uid, otherId)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if res == true {
		w.WriteJson(map[string]string{"following": v.Get("otherId"),
			"follower": v.Get("uid"), "unfollowed": "true"})
	} else {
		w.WriteJson(map[string]string{"following": v.Get("otherId"),
			"follower": v.Get("uid"), "unfollowed": "false"})
	}
}

func (i *Impl) GetTimeline(w rest.ResponseWriter, r *rest.Request) {
	v, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uid, err := strconv.Atoi(v.Get("uid"))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page, err := strconv.Atoi(v.Get("page"))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := i.DB.GetUserTimeline(uid, page, 30)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output := new(TimelineResponse)
	output.Posts = make([]rdb.Status, len(res))
	output.Uid = uid
	output.Page = page
	outputIndex := 0
	// channel to send/recieve status structs
	statuses := make(chan rdb.Status)
	var pst int
	for j := 0; j < len(res); j++ {
		pst = res[j]
		// anon goroutine to get a status in timeline page
		go func(post int) {
			status, err := i.DB.GetStatus(post)
			if err != nil {
				log.Printf("error getting post %d, %v\n", post, err)
				return
			}
			statuses <- status
		}(pst)
	}

	for outputIndex < len(res) {
		select {
		case sts := <-statuses:
			output.Posts[outputIndex] = sts
			outputIndex++
		case <-time.After(time.Second * 1):
			log.Printf("timeout getting timeline:%d page:%d\n", uid,
				page)
			break
		}
	}
	sort.Sort(output.Posts)
	w.WriteJson(&output)
}

func (i *Impl) GetUser(w rest.ResponseWriter, r *rest.Request) {
	v, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uid, err := strconv.Atoi(v.Get("uid"))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	usr, err := i.DB.GetUser(uid)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(&usr)
}
