/*

Simple, stupid Go web service exposing a map as a key-value store over HTTP

CRUD semantics:

Function | HTTP Method |
---------|-------------|
Create   | POS         |
Read     | GET         |
Update   | PUT         |
Delete   | DELETE      |

*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"
)

var (
	kv      = map[string]string{}
	mutex   = sync.RWMutex{}
	safety  = false
	ts_save = int64(0) // unix timestamp of last save
	ts_chng = int64(0) // unix timestamp of last change
)

func asJson(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	if status == http.StatusOK {
		resp["success"] = "true"
	} else {
		resp["success"] = "false"
	}
	resp["message"] = msg
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func saveToDisk() {
	// Load file from disk or create it if necessary
	f, err := os.OpenFile("data.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(f)
	err = encoder.Encode(kv)

	f.Close()

	fmt.Println("saved to disk")
}

func loadFromDisk() {
	f, err := os.Open("data.json")
	if err == nil {
		encoder := json.NewDecoder(f)
		err = encoder.Decode(&kv)
		f.Close()
	}
}

func saveIfNeeded() {
	// save to disk if enough time has elapsed
	const secs_between_saves = 30
	if (ts_chng - ts_save) >= secs_between_saves {
		saveToDisk()
		ts_save = time.Now().Unix()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[1:]

	// key-value store service -> CRUD functionality

	switch r.Method {

	// CREATE
	case "POST":
		if safety {
			mutex.Lock()
		}
		if _, ok := kv[url]; ok {
			asJson(w, http.StatusBadRequest, "already exists")
		} else {
			defer r.Body.Close()
			buf, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatalf("Error happened in PUT-handler (io.ReadAll failed). Err: %s", err)
			} else {
				reqBody := string(buf)
				kv[url] = reqBody
				asJson(w, http.StatusOK, reqBody)
				ts_chng = time.Now().Unix()
			}
		}
		saveIfNeeded()
		if safety {
			mutex.Unlock()
		}

	// READ
	case "GET":
		if safety {
			mutex.RLock()
		}
		if val, ok := kv[url]; ok {
			asJson(w, http.StatusOK, val)
		} else {
			asJson(w, http.StatusNotFound, "not found")
		}
		if safety {
			mutex.RUnlock()
		}

	// UPDATE
	case "PUT":
		if safety {
			mutex.Lock()
		}
		if _, ok := kv[url]; ok {
			defer r.Body.Close()
			buf, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatalf("Error happened in POST-handler (io.ReadAll failed). Err: %s", err)
			}
			reqBody := string(buf)
			kv[url] = reqBody
			ts_chng = time.Now().Unix()
			asJson(w, http.StatusOK, reqBody)
		} else {
			asJson(w, http.StatusNotFound, "not found")
		}
		saveIfNeeded()
		if safety {
			mutex.Unlock()
		}

	// DELETE
	case "DELETE":
		if safety {
			mutex.Lock()
		}
		if val, ok := kv[url]; ok {
			delete(kv, url)
			ts_chng = time.Now().Unix()
			asJson(w, http.StatusOK, val)
		} else {
			asJson(w, http.StatusNotFound, "not found")
		}
		saveIfNeeded()
		if safety {
			mutex.Unlock()
		}

	default:
		asJson(w, http.StatusBadRequest, "Only PUT, GET, DELETE and POST methods are supported.")
	}

}

func newReuseAddrListener(network, address string) (net.Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	// Set the SO_REUSEADDR option
	tcpListener, ok := l.(*net.TCPListener)
	if !ok {
		return nil, fmt.Errorf("listener is not a TCPListener")
	}

	// Get the underlying file descriptor
	file, err := tcpListener.File()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Set the socket option
	if err := setReuseAddr(file.Fd()); err != nil {
		return nil, err
	}

	return l, nil
}

func setReuseAddr(fd uintptr) error {
	return syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
}

func main() {

	addr := ":8080"

	// Create a new listener with SO_REUSEADDR enabled
	listener, err := newReuseAddrListener("tcp", addr)
	if err != nil {
		fmt.Println("Error creating listener:", err)
		os.Exit(1)
	}
	defer listener.Close()

	safetyp := flag.Bool("safety", false, "thread-safety enabled")
	flag.Parse()
	safety = *safetyp

	loadFromDisk()

	// Create an HTTP server
	server := &http.Server{
		Addr: addr,
	}

	http.HandleFunc("/", handler)

	log.Printf("Starting web service, listening on '%s' - thread safe: %v\n", addr, safety)

	// Start the server
	log.Fatal(server.Serve(listener))
}
