![CI](https://github.com/kokke/race-detector-demo-go/actions/workflows/go.yml/badge.svg)

# Automatic detection of race conditions in Go

This is a demo using Go's builtin data race detection mechanism to find race conditions automatically.

Official documentation: <https://go.dev/doc/articles/race_detector>

Blog post introducing the feature: <https://go.dev/blog/race-detector>

## Description

`ws.go` is a small web service exposing a key-value store with CRUD-semantics mapped to HTTP PUT/GET/POST/DELETE.

Data is stored in a `map[string]string` and a CLI-flag `-safety` determines whether or not to use thread synchronization.

Building with the data race detector enabled (`go build --race ...`) makes the program crash with a trace when a data race is encountered.

In this demo, we build the web service with and without thread synchronization and use [Apache HTTP server benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html) `ab` to make concurrent read and write requests.

## Assumptions

The demo assumed the following programs are available on the command-line:

```sh
ab
curl
go
make
```

They can be installed on Debian/Ubuntu with `sudo apt-get install -y curl golang make apache2-utils`.

## Instructions

Run `make clean all test` to build and test a HTTP web service with and without thread safety.

Building the code with `go build --race` enables the data race detector, which will crash the program if a race is detected.

Output looks like this on my machine:

```sh
$ make clean all test

make clean
make all
make test

testing thread-safe version
2024/10/29 15:32:07 Starting web service, listening on ':8080' - thread safe: true

testing thread-unsafe version
2024/10/29 15:32:11 Starting web service, listening on ':8080' - thread safe: false

==================
WARNING: DATA RACE
Write at 0x00c000124c00 by goroutine 11:
  runtime.mapassign_faststr()
      /usr/lib/go-1.19/src/runtime/map_faststr.go:203 +0x0
  main.handler()
      /mnt/d/tmp/golang-race-detector/ws.go:99 +0x71d
  net/http.HandlerFunc.ServeHTTP()
      /usr/lib/go-1.19/src/net/http/server.go:2109 +0x4d
  net/http.(*ServeMux).ServeHTTP()
      /usr/lib/go-1.19/src/net/http/server.go:2487 +0xc5
  net/http.serverHandler.ServeHTTP()
      /usr/lib/go-1.19/src/net/http/server.go:2947 +0x641
  net/http.(*conn).serve()
      /usr/lib/go-1.19/src/net/http/server.go:1991 +0xbe4
  net/http.(*Server).Serve.func3()
      /usr/lib/go-1.19/src/net/http/server.go:3102 +0x58
```

Here the program crashes with an error message about a `map`-write in `ws.go` line 99.

We can verify that this is indeed where the data is racing:

```sh
$ cat -n ws.go | grep -A5 -B5 99
    94                          buf, err := io.ReadAll(r.Body)
    95                          if err != nil {
    96                                  log.Fatalf("Error happened in POST-handler (io.ReadAll failed). Err: %s", err)
    97                          }
    98                          reqBody := string(buf)
    99                          kv[url] = reqBody
   100                          asJson(w, http.StatusOK, reqBody)
   101                  } else {
   102                          asJson(w, http.StatusNotFound, "not found")
   103                  }
   104                  if safety { mutex.Unlock() }
```

The trace output shows where the data race was triggered specifically and a complete stacktrace of how the execution got there.
