package drones

import "testing"

import (
	"github.com/benbearchen/antpost"

	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type Http struct {
}

func (h *Http) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request) {
	path := r.URL.Path
	if !r.Close {
		w.Header().Set("Connection", "Keep-Alive")
	}

	w.WriteHeader(200)
	if r.Method == "GET" {
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		fmt.Fprintf(w, "%s", path)
	} else {
		defer r.Body.Close()
		bytes, _ := ioutil.ReadAll(r.Body)
		w.Write(bytes)
	}
}

func (h *Http) Run(addr string) {
	http.ListenAndServe(addr, h)
}

func TestAsyncHttp(t *testing.T) {
	addr := "localhost:8848"
	go new(Http).Run(addr)

	async, err := newAsyncHttp(addr)
	if err != nil {
		t.Errorf("newAsyncHttp() failed: %v", err)
		return
	}

	async.Get("http://localhost:8848/first", true)
	response, ok := <-async.Response()
	if !ok {
		t.Errorf("first meet close")
	} else if response != nil {
		statusCode, content := response.StatusCode, response.Content
		if statusCode != 200 || string(content) != "/first" {
			t.Errorf("first result not ok, statusCode %v, content '%v'", statusCode, string(content))
		}
	}

	async.Get("http://localhost:8848/second", true)
	async.Post("http://localhost:8848/", []byte("third"), false)
	next := "/second"
	for len(next) > 0 {
		select {
		case response, ok := <-async.Response():
			if !ok {
				t.Errorf("%s meet close", next)
				next = ""
				break
			} else {
				statusCode, content := response.StatusCode, response.Content
				if statusCode != 200 || string(content) != next {
					t.Errorf("%v result not ok, statusCode %v, content '%v'", next, statusCode, string(content))
				} else {
					if next == "/second" {
						next = "third"
					} else {
						next = ""
						break
					}
				}
			}
		}
	}
}

type asyncop struct {
	url   string
	datas [][]byte
	c     chan *AsyncHttpReq
	stop  chan bool
}

func newop(url string, datas [][]byte) *asyncop {
	op := asyncop{url, datas, make(chan *AsyncHttpReq), make(chan bool)}
	go op.run()
	return &op
}

func (a *asyncop) Response(context *antpost.Context, response *AsyncHttpResposne) {
	if response.Err != nil || string(response.Content) == "hello" {
		a.datas = nil
		close(a.c)
		close(a.stop)
	}
}

func (a *asyncop) Next() <-chan *AsyncHttpReq {
	return a.c
}

func (a *asyncop) NextSession() *AsyncHttpSession {
	host, _, _ := parseUrl(a.url)
	datas := [][]byte{[]byte("hello"), []byte("world"), []byte("second")}
	shuffle := make([][]byte, len(datas))
	for i, v := range rand.Perm(len(datas)) {
		shuffle[i] = datas[v]
	}

	return &AsyncHttpSession{host, newop(a.url, shuffle)}
}

func (a *asyncop) run() {
	for {
		if len(a.datas) > 0 {
			data := a.datas[0]
			a.datas = a.datas[1:]
			keepalive := len(a.datas) > 0
			a.c <- &AsyncHttpReq{a.url, "POST", nil, data, keepalive}
		} else if a.datas != nil {
			a.datas = nil
			close(a.c)
		}

		select {
		case <-time.NewTimer(640e6).C:
		case <-a.stop:
			return
		}
	}
}

func TestAsyncHttpDrone(t *testing.T) {
	addr := "localhost:8849"
	go new(Http).Run(addr)

	h := newop("http://"+addr+"/", nil).NextSession()
	d := NewAsyncHttpDrone(h)
	for i := 1; i <= 256; i *= 2 {
		c := antpost.Run(d, i, 0, 15*time.Second)
		fmt.Println("goroutines: ", i, c.Report()[0])
		time.Sleep(time.Second * 1)
	}
}
