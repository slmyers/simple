package main

/*
 * backend for simple social network
 */

import (
	rdb "./db"
	"flag"
	"fmt"
	"github.com/slmyers/go-json-rest/rest"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strconv"
	"time"
)

var (
	port = flag.String("port", "8000", "port bound to server")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	i := Impl{}
	i.InitDB()
	flag.Parse()
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	// declare the handlers for various requests
	router, err := rest.MakeRouter(
		rest.Post("/users", i.CreateUser),
		rest.Post("/status", i.PostStatus),
		rest.Post("/follow", i.FollowUser),
		rest.Post("/unfollow", i.UnfollowUser),
		rest.Get("/timelines", i.GetTimeline),
		rest.Get("/users", i.GetUser),
		// uncomment if you would also like to serve files
		//rest.Get("/", homeHandler),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	fmt.Printf("serving on http://localhost:%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, api.MakeHandler()))
}

/* this is included as an example of how to serve files (webpages)*/
func homeHandler(w rest.ResponseWriter, r *rest.Request) {
	http.ServeFile(w.(http.ResponseWriter), r.Request,
		r.URL.Path[1:])
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

	userOut, err := i.DB.GetUser(uid)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&userOut)
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

	post, err := i.DB.GetStatus(sid)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(&post)
}

/*
 * handles requests of form /follow?uid=2&otherId=3
 */
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

/*
 * handles requests of form /unfollow?uid=2&otherId=3
 */

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

/*
 * handles requests of the form /timeline?uid=7&page=1
 */

func (i *Impl) GetTimeline(w rest.ResponseWriter, r *rest.Request) {
	v, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var uid int
	if _, val := v["uid"]; val {
		uid, _ = strconv.Atoi(v.Get("uid"))
	} else if _, val := v["login"]; val {
		uid = i.DB.GetUserID(v.Get("login"))
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		rest.NotFound(w, r)
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
	statuses := make(chan rdb.Status, len(res))
	defer close(statuses)
	var pst int
	for j := 0; j < len(res); j++ {
		pst = res[j]
		// anon goroutine to get a status in timeline page
		// this means that all statuses are fetched concurrently
		go func(post int) {
			status, err := i.DB.GetStatus(post)
			if err != nil {
				log.Printf("error getting post %d, %v\n", post, err)
				return
			}
			// pipe the fetched status into the channel previously made
			statuses <- status
		}(pst)
	}

	// this code is blocking
Loop:
	for outputIndex < len(res) {
		select {
		case sts := <-statuses:
			output.Posts[outputIndex] = sts
			outputIndex++
		case <-time.After(time.Second * 1):
			log.Printf("timeout getting timeline:%d page:%d\n", uid,
				page)
			break Loop
		}
	}
	// because the statuses were retrieved concurrently we can't be sure
	// of what order they will appear in output.Posts, so we must sort them
	// if we want them to appear from newest to oldest
	sort.Sort(output.Posts)
	res2 := make([]TimelineResponse, 1)
	res2[0] = *output
	w.WriteJson(map[string]interface{}{"timeline": res2})
}

/*
 * handles requests of the form /user?uid=7
 */

func (i *Impl) GetUser(w rest.ResponseWriter, r *rest.Request) {
	v, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if uid, ok := v["uid"]; ok {
		id, _ := strconv.Atoi(uid[0])
		usr, err := i.DB.GetUser(id)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res := make(rdb.Users, 1)
		res[0] = *usr
		w.WriteJson(&res)
		return
	} else if login, ok := v["login"]; ok {
		usr, err := i.DB.GetUserByLogin(login[0])
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res := make(rdb.Users, 1)
		res[0] = *usr
		// http://dukex.svbtle.com/tip-golang-with-emberjs-json-response
		w.WriteJson(map[string]interface{}{"user": res})
		return
	} else {
		rest.NotFound(w, r)
		return
	}
}
