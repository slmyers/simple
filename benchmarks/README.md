# benchmarks

I'm using the [Apache HTTP server benchmarking tool](http://httpd.apache.org/docs/2.0/programs/ab.html) to benchmark the server. The server was run on my local machine.

### machine specs
12GB RAM
4x2.67Ghz processor

### get user summary 
`ab -c 100 -n 100 http://127.0.0.1:8000/user?uid=7`

```
Percentage of the requests served within a certain time (ms)
  50%     18
  66%     20
  75%     25
  80%     26
  90%     29
  95%     30
  98%     31
  99%     31
 100%     31 (longest request)
```
### get timeline summary
`ab -c 100 -n 100 "http://127.0.0.1:8000/timeline?uid=7&page=1"`

```
Percentage of the requests served within a certain time (ms)
  50%     16
  66%     21
  75%     23
  80%     25
  90%     28
  95%     29
  98%     30
  99%     30
 100%     30 (longest request)
```

### follow summary
`ab -c 100 -n 100 "http://127.0.0.1:8000/follow?uid=1&otherId=7"`

```
Percentage of the requests served within a certain time (ms)
  50%     13
  66%     16
  75%     17
  80%     17
  90%     18
  95%     19
  98%     20
  99%     20
 100%     20 (longest request)
```

### unfollow summary
`ab -c 100 -n 100 "http://127.0.0.1:8000/unfollow?uid=1&otherId=7"`

```
Percentage of the requests served within a certain time (ms)
  50%     15
  66%     18
  75%     19
  80%     19
  90%     20
  95%     21
  98%     22
  99%     22
 100%     22 (longest request)
```
