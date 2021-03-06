package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
	_ "net/http/pprof"
)

var imgsRe *regexp.Regexp
var mu sync.Mutex
var nextUpdate time.Time
var imgs [][]byte

func init() {
	imgsRe = regexp.MustCompile(`(http://i.imgur.com/[^"]{5,20}|https://i.redditmedia.com/[^"]+\?s=[^"]+)|https://i.redd.it/[0-9a-z.]{10,30}`)
}

func maybeUpdateImgs() ([][]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	if time.Now().After(nextUpdate) {
		req, err := http.NewRequest("GET", "http://www.reddit.com/r/catpics/hot.json", nil)
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
			log.Printf("No cat pictures found: %s: %s", resp, body)
			return nil, errors.New("No cat pictures found")
		}

		newImgs = append(newImgs, []byte("http://i.imgur.com/iHzbXfL.jpg"))
		imgs = newImgs
		nextUpdate = time.Now().Add(2 * time.Second)
	}

	return imgs, nil
}

func serveUrl(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain")

	imgs, e := maybeUpdateImgs()
	if e != nil {
		log.Print(e)
		http.Error(rw, e.Error(), 503)
		return
	}

	img := imgs[rand.Intn(len(imgs))]
	rw.Header().Set("Content-Length", strconv.Itoa(len(img)))
	rw.Write([]byte(img))
}

func serveRoot(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")

	imgs, e := maybeUpdateImgs()
	if e != nil {
		log.Print(e)
		http.Error(rw, e.Error(), 503)
		return
	}

	img := imgs[rand.Intn(len(imgs))]
	html := fmt.Sprintf(
		"<title>um, cats</title>"+
			"<body background=http://104.131.51.57/static/bg.jpg>"+
			"<div align=center style=\"width: 100%%; height: 100%%\">"+
			"<img style=\"max-width: 90%%; max-height: 80%%\" src=\"%s\"/>"+
			"</div>", img)
	rw.Header().Set("Content-Length", strconv.Itoa(len(html)))
	rw.Write([]byte(html))
}

func main() {
	logger, e := syslog.New(syslog.LOG_NOTICE | syslog.LOG_DAEMON, "cats")
	if e != nil {
		log.Fatal(e)
	}
	log.SetOutput(logger)
	http.HandleFunc("/url", serveUrl)
	http.HandleFunc("/", serveRoot)

	l, e := net.Listen("tcp", "127.0.0.1:8000")
	if e != nil {
		log.Fatal(e)
	}

	log.Print("serving")
	e = http.Serve(l, nil)
	if e != nil {
		log.Fatal(e)
	}
}
