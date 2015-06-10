# simple
simple social media app using redis written in go

need redigo to run:
`go get github.com/garyburd/redigo/redis`

and you'll also need `redis`:
http://redis.io/

use this command to build
`go build server.go models.go`

featured on my blog:

http://slmyers.github.io/simple/social/network/2015/05/29/Simple-Social-Network/

# examples with output

### create user
```
curl -i \
-H 'Content-Type: application/json' \
-X POST -d '{"username": "slmyers", "name": "Steven Myers"}' \
http://127.0.0.1:8000/user
```

```
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
```
### post status
```
curl -i \
-H 'Content-Type: application/json' \
-X POST -d '{"uid": 7, "msg": "This is just a test."}' \
http://127.0.0.1:8000/status
```

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Powered-By: go-json-rest
Date: Mon, 01 Jun 2015 19:50:06 GMT
Content-Length: 108

{
  "Message": "This is just a test.",
  "Posted": 1433188206,
  "Id": 9,
  "Uid": 7,
  "Login": "slmyers"
}
```

### follow user
```
curl -i -X POST "http://127.0.0.1:8000/follow?uid=1&otherId=7"
```

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Powered-By: go-json-rest
Date: Mon, 01 Jun 2015 20:06:02 GMT
Content-Length: 63

{
  "followed": "true",
  "follower": "1",
  "following": "7"
}
```

### unfollow user
```
curl -i -X POST "http://127.0.0.1:8000/unfollow?uid=1&otherId=7"
```

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Powered-By: go-json-rest
Date: Mon, 01 Jun 2015 20:54:34 GMT
Content-Length: 65

{
  "follower": "1",
  "following": "7",
  "unfollowed": "true"
}
```

### get user's timeline
```
curl -i "http://127.0.0.1:8000/timeline?uid=7&page=1"
```

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Powered-By: go-json-rest
Date: Mon, 01 Jun 2015 21:19:18 GMT
Content-Length: 458

{
  "uid": 7,
  "page": 1,
  "posts": [
    {
      "Message": "This is just a test.",
      "Posted": 1433188206,
      "Id": 9,
      "Uid": 7,
      "Login": "slmyers"
    },
    {
      "Message": "This is just a test.",
      "Posted": 1433188118,
      "Id": 8,
      "Uid": 7,
      "Login": "slmyers"
    },
    {
      "Message": "This is just a test.",
      "Posted": 1433188059,
      "Id": 7,
      "Uid": 7,
      "Login": "slmyers"
    }
  ]
}
```

### get user

```
curl -i "http://127.0.0.1:8000/user?uid=7"
```

```
HTTP/1.1 200 OK
Content-Type: application/json
X-Powered-By: go-json-rest
Date: Mon, 01 Jun 2015 21:25:27 GMT
Content-Length: 135

{
  "login": "slmyers",
  "Id": 7,
  "name": "Steven Myers",
  "Followers": 0,
  "Following": 0,
  "Posts": 3,
  "Signup": 1433183728
}
```
