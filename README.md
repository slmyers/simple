# CURRENTLY UNTESTED

# simple
api for simple social media app backed by redis written in go

need redigo to run:
`go get github.com/garyburd/redigo/redis`

and you'll also need `redis`:
http://redis.io/

use this command to build
`go build server.go models.go`

# examples with output

>curl -i \
-H 'Content-Type: application/json' \
-X POST -d '{"username": "slmyers", "name": "Steven Myers"}' \
http://127.0.0.1:8000/user
HTTP/1.1 200 OK

Content-Type: application/json
X-Powered-By: go-json-rest
Date: Mon, 01 Jun 2015 18:35:28 GMT
Content-Length: 135

{
  "login": "slmyers",
  "Id": 7,
  "name": "Steven Myers",
  "Followers": 0,
  "Following": 0,
  "Posts": 0,
  "Signup": 1433183728
}
