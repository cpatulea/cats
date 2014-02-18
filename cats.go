package main

import (
  "net/http/fcgi"
  "log"
  "net"
  "net/http"
  "io/ioutil"
  "regexp"
  "fmt"
  "math/rand"
  "strconv"
)

var imgsRe *regexp.Regexp

func init() {
  imgsRe = regexp.MustCompile(`http://i.imgur.com/[^"]{5,20}`)
}

type server struct{}

func (s server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
  rw.Header().Set("Content-Type", "text/html")

  req, err := http.NewRequest("GET", "http://www.reddit.com/r/catpictures/hot.json", nil)
  if err != nil {
    log.Print(err)
    rw.Write([]byte(err.Error()))
    return
  }

  req.Header.Add("User-Agent", "cats.go 0.1 (contact /u/eigma or cronos586@gmail.com)")
  resp, err := (&http.Client{}).Do(req)
  if err != nil {
    log.Print(err)
    rw.Write([]byte(err.Error()))
    return
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Print(err)
    rw.Write([]byte(err.Error()))
    return
  }

  imgs := imgsRe.FindAll(body, -1)
  if len(imgs) == 0 {
    log.Print("no imgs")
    rw.Write([]byte("no imgs"))
    return
  }

  img := imgs[rand.Intn(len(imgs))]
  html := fmt.Sprintf(
      "<title>um, cats</title>" +
      "<body background=/static/bg.jpg>" +
      "<div align=center>" +
      "<img style=\"max-width: 90%%; max-height: 80%%\" src=\"%s\"/>" +
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
