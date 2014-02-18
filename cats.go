package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/fcgi"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var imgsRe *regexp.Regexp
var mu sync.Mutex
var nextUpdate time.Time
var imgs [][]byte

func init() {
	imgsRe = regexp.MustCompile(`http://i.imgur.com/[^"]{5,20}`)
}

type server struct{}

func maybeUpdateImgs() ([][]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	if time.Now().After(nextUpdate) {
		req, err := http.NewRequest("GET", "http://www.reddit.com/r/catpictures/hot.json", nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("User-Agent", "cats.go 0.1 (contact /u/eigma or cronos586@gmail.com)")
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		newImgs := imgsRe.FindAll(body, -1)
		if len(newImgs) == 0 {
			return nil, err
		}

		imgs = newImgs
		nextUpdate = time.Now().Add(1 * time.Second)
	}

	return imgs, nil
}

func (s server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")

	imgs, e := maybeUpdateImgs()
	if e != nil {
		log.Print(e)
		rw.Write([]byte(e.Error()))
		return
	}

	img := imgs[rand.Intn(len(imgs))]
	html := fmt.Sprintf(
		"<title>um, cats</title>"+
			"<body background=/static/bg.jpg>"+
			"<div align=center>"+
			"<img style=\"max-width: 90%%; max-height: 80%%\" src=\"%s\"/>"+
			"</div>", img)
	rw.Header().Set("Content-Length", strconv.Itoa(len(html)))
	rw.Write([]byte(html))
}

func main() {
	l, e := net.Listen("tcp", "127.0.0.1:8080")
	if e != nil {
		log.Fatal(e)
	}

	log.Print("serving")
	e = fcgi.Serve(l, &server{})
	if e != nil {
		log.Fatal(e)
	}
}
