package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	_ "embed"
)

const message = "reload"

//go:embed dev.js.template
var devJsTemplate string

var (
	jsFile = flag.String("f", "", "js file path")
	addr   = flag.String("l", "0.0.0.0:8080", "listen address")
)

var (
	upgrader  = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	hasher    = md5.New()
	clients   = make(map[*websocket.Conn]struct{})
	changedCh = make(chan struct{})
)

func main() {
	flag.Parse()
	_, err := os.Stat(*jsFile)
	if os.IsNotExist(err) {
		fmt.Println("js file not exists")
		os.Exit(1)
	}

	fmt.Printf("open follow url to install the dev script\n\n  http://127.0.0.1:%s/dev.user.js\n\n", strings.Split(*addr, ":")[1])

	go boardercaster()
	go fileWatcher()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/dev.user.js", serveUserJs)
	http.HandleFunc("/ws", serveWs)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

func serveUserJs(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(getDevScript()))
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	clients[c] = struct{}{}
}

func boardercaster() {
	ticker := time.NewTicker(time.Minute)
	for {
		messageType := websocket.PingMessage
		data := []byte{}

		select {
		case <-changedCh:
			messageType = websocket.TextMessage
			data = []byte(message)
		case <-ticker.C:
			log.Println("send ping message")
		}

		for c := range clients {
			go func(c *websocket.Conn) {
				err := c.WriteMessage(messageType, data)
				if err != nil {
					delete(clients, c)
					c.Close()
				}
			}(c)
		}
	}
}

func fileWatcher() {
	oldHash := make([]byte, 0)

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		hash, err := calcFileHash(*jsFile)
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(oldHash, hash) {
			log.Printf("file changed: %s", hex.EncodeToString(hash))
			changedCh <- struct{}{}
			oldHash = hash
		}
	}
}

func calcFileHash(name string) ([]byte, error) {
	fd, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	hasher.Reset()
	io.Copy(hasher, fd)
	return hasher.Sum(nil), nil
}

func getDevScript() string {
	fd, err := os.Open(*jsFile)
	if err != nil {
		panic(err)
	}
	lines := make([]string, 0, 50)
	scaner := bufio.NewScanner(fd)
	for scaner.Scan() {
		l := scaner.Text()
		if strings.HasPrefix(l, "// ==/UserScript==") {
			path, _ := filepath.Abs(*jsFile)
			content := strings.Replace(devJsTemplate, "__FILE_PATH__", path, 1)
			port := strings.Split(*addr, ":")[1]
			content = strings.Replace(content, "__PORT__", port, 1)
			lines = append(lines, content)
			break
		}
		if strings.HasPrefix(l, "//") {
			lines = append(lines, l)
		}
	}

	return strings.Join(lines, "\n")
}
