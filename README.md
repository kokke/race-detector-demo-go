# Golang Race Detector Demo

This is a demo of Go's builtin data race detection mechanism and how to use it to find race conditions.

Documentation link: <https://go.dev/doc/articles/race_detector>

Blog post introducing the feature: <https://go.dev/blog/race-detector>

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

Run the `make clean all test` to build and test a HTTP web service with and without thread safety.

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
Read at 0x00c0001a2c00 by goroutine 16:
  runtime.mapaccess2_faststr()
      /usr/lib/go-1.19/src/runtime/map_faststr.go:108 +0x0
  main.handler()
      /mnt/d/tmp/golang-race-detector/ws.go:92 +0x559
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

