package drones

import (
	"github.com/benbearchen/antpost"

	"bytes"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type AsyncHttpResposne struct {
	StatusCode int
	Content    []byte
	Err        error
}

type AsyncHttpReq struct {
	Url       string
	Method    string
	Header    http.Header
	Data      []byte
	KeepAlive bool
}

type AsyncHttpSession struct {
	Host     string
	Operator AsyncOperator
}

type AsyncOperator interface {
	Response(context *antpost.Context, response *AsyncHttpResposne)
	Next() <-chan *AsyncHttpReq
	NextSession() *AsyncHttpSession
}

func NewAsyncHttpDrone(h *AsyncHttpSession) antpost.Drone {
	return &asyncHttpDrone{h}
}

type asyncHttpDrone struct {
	http *AsyncHttpSession
}

func (h *asyncHttpDrone) Run(context *antpost.Context) antpost.DroneResult {
	async, err := newAsyncHttp(h.http.Host)
	context.Step(antpost.StepConnected)
	if err != nil {
		return antpost.ResultConnectFail
	}

	p, q := false, false
	for !p && !q {
		select {
		case req, ok := <-h.http.Operator.Next():
			if !ok {
				p = true
			} else {
				async.Do(req.Url, req.Method, req.Header, req.Data, req.KeepAlive)
			}
		case response, ok := <-async.Response():
			if !ok {
				q = true
			} else {
				h.http.Operator.Response(context, response)
			}
		}
	}

	async.Shutdown()

	if !q {
		for !q {
			select {
			case response, ok := <-async.Response():
				if !ok {
					q = true
				} else {
					h.http.Operator.Response(context, response)
				}
			}
		}
	}

	context.Step(antpost.StepResponsed)
	if err != nil {
		return antpost.ResultResponseBroken
	} else {
		return antpost.ResultOK
	}
}

func (h *asyncHttpDrone) Next() antpost.Drone {
	return NewAsyncHttpDrone(h.http.Operator.NextSession())
}

type asyncHttpRequest struct {
	url       string
	method    string
	header    http.Header
	data      []byte
	keepalive bool
}

type asyncHttp struct {
	conn  *net.TCPConn
	wlock sync.Mutex
	w     []*asyncHttpRequest
	wc    chan bool
	r     chan *AsyncHttpResposne
}

func newAsyncHttp(hostport string) (*asyncHttp, error) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		host, port, err = net.SplitHostPort(hostport + ":80")
		if err != nil {
			return nil, err
		}
	}

	if len(port) == 0 {
		port = "80"
	}

	addr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	c := new(asyncHttp)
	c.conn = conn
	c.w = make([]*asyncHttpRequest, 0)
	c.wc = make(chan bool)
	c.r = make(chan *AsyncHttpResposne)
	go c.goWrite()
	go c.goRead()
	return c, nil
}

func (h *asyncHttp) Shutdown() {
	h.newReq(nil)
	close(h.wc)
}

func (h *asyncHttp) Post(url string, data []byte, keepalive bool) {
	h.newReq(&asyncHttpRequest{url, "POST", nil, data, keepalive})
	h.wc <- true
}

func (h *asyncHttp) Get(url string, keepalive bool) {
	h.newReq(&asyncHttpRequest{url, "GET", nil, nil, keepalive})
	h.wc <- true
}

func (h *asyncHttp) Do(url, method string, header http.Header, data []byte, keepalive bool) {
	h.newReq(&asyncHttpRequest{url, method, header, data, keepalive})
	h.wc <- true
}

func (h *asyncHttp) newReq(req *asyncHttpRequest) {
	h.wlock.Lock()
	defer h.wlock.Unlock()
	if req != nil {
		h.w = append(h.w, req)
	} else {
		h.w = nil
	}
}

func (h *asyncHttp) doWrite(req *asyncHttpRequest) error {
	host, path, err := parseUrl(req.url)
	if err != nil {
		return err
	}

	headerBytes := h.createHeader(req.method, path, host, req.header, len(req.data), req.keepalive)
	err = h.write(headerBytes)
	if err != nil {
		return err
	}

	if len(req.data) > 0 {
		return h.write(req.data)
	}

	return nil
}

func (h *asyncHttp) Response() <-chan *AsyncHttpResposne {
	return h.r
}

func (h *asyncHttp) createHeader(method, path, host string, header http.Header, contentLength int, keepalive bool) []byte {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%s %s HTTP/1.1\r\n", method, path)
	fmt.Fprintf(buf, "Host: %s\r\n", host)
	fmt.Fprintf(buf, "Content-Length: %d\r\n", contentLength)
	if keepalive {
		fmt.Fprintf(buf, "Connection: Keep-Alive\r\n")
	} else {
		fmt.Fprintf(buf, "Connection: close\r\n")
	}

	if header != nil {
		for k, v := range header {
			fmt.Fprintf(buf, "%s: %s\r\n", k, v)
		}
	}

	fmt.Fprintf(buf, "\r\n")
	return buf.Bytes()
}

func (h *asyncHttp) getReq() *asyncHttpRequest {
	h.wlock.Lock()
	defer h.wlock.Unlock()

	if len(h.w) > 0 {
		req := h.w[0]
		h.w = h.w[1:]
		return req
	} else {
		return nil
	}
}

func (h *asyncHttp) goWrite() {
	defer h.conn.CloseWrite()
	for {
		select {
		case _, ok := <-h.wc:
			if !ok {
				return
			}

			req := h.getReq()
			if req == nil {
				continue
			}

			h.doWrite(req)
		}
	}
}

func (h *asyncHttp) write(data []byte) error {
	sent := 0
	for sent < len(data) {
		n, err := h.conn.Write(data[sent:])
		if n > 0 {
			sent += n
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (h *asyncHttp) goRead() {
	c := newChunkedHttpParse()
	defer h.conn.Close()
	defer func() { close(h.r) }()
	for !c.readOver {
		s, err := h.read()
		if len(s) > 0 {
			c.write(s)
		}

		if err != nil {
			if err == io.EOF {
				c.write(nil)
			} else {
				return
			}
		}

		for {
			response, over := c.tryParse()
			if response != nil {
				h.r <- response
			} else if over {
				return
			} else {
				break
			}
		}
	}
}

func (h *asyncHttp) read() ([]byte, error) {
	buf := make([]byte, 2048)
	n, err := h.conn.Read(buf)
	if n > 0 {
		buf = buf[:n]
	} else {
		buf = nil
	}

	return buf, err
}

func parseUrl(reqUrl string) (hostport, path string, err error) {
	u, err := url.Parse(reqUrl)
	if err != nil {
		return "", "", err
	}

	return u.Host, u.RequestURI(), nil
}

type chunkedHttpParse struct {
	buf          *bytes.Buffer
	statusCode   int
	contentBytes int
	chunked      *bytes.Buffer
	readOver     bool
}

func newChunkedHttpParse() *chunkedHttpParse {
	parse := new(chunkedHttpParse)
	parse.buf = &bytes.Buffer{}
	parse.contentBytes = -1
	parse.chunked = nil
	return parse
}

func (c *chunkedHttpParse) write(newBytes []byte) {
	if len(newBytes) > 0 {
		c.buf.Write(newBytes)
	} else {
		c.readOver = true
	}
}

func (c *chunkedHttpParse) tryParse() (response *AsyncHttpResposne, tcpOver bool) {
	if c.readOver && c.buf.Len() == 0 {
		return nil, true
	}

	if c.contentBytes < 0 {
		content := c.buf.Bytes()
		rnrn := []byte{'\r', '\n', '\r', '\n'}
		h := bytes.Index(content, rnrn)
		if h >= 0 {
			c.parseHeader(c.buf.Next(h + len(rnrn)))
		} else {
			return nil, false
		}
	}

	if c.contentBytes == 0 {
		response := &AsyncHttpResposne{c.statusCode, make([]byte, 0), nil}
		c.statusCode = 0
		c.contentBytes = -1
		c.chunked = nil
		return response, false
	} else if c.chunked != nil {
		finish, err := c.readChunked()
		if err != nil {
			// TODO:
		}

		if finish {
			response := &AsyncHttpResposne{c.statusCode, c.chunked.Bytes(), nil}
			c.statusCode = 0
			c.contentBytes = -1
			c.chunked = nil
			return response, false
		}
	} else if (c.contentBytes > 0 && c.buf.Len() >= c.contentBytes) || c.readOver {
		size := c.contentBytes
		if size > c.buf.Len() {
			size = c.buf.Len()
		}

		response := &AsyncHttpResposne{c.statusCode, c.buf.Next(size), nil}
		c.statusCode = 0
		c.contentBytes = -1
		c.chunked = nil
		return response, false
	}

	return nil, false
}

func (c *chunkedHttpParse) parseHeader(header []byte) error {
	hs := bytes.Split(header, []byte{'\r', '\n'})
	for i, h := range hs {
		if i == 0 {
			s := bytes.SplitN(h, []byte{' '}, 2)
			if len(s) == 2 {
				s = bytes.SplitN(bytes.TrimSpace(s[1]), []byte{' '}, 2)
				if code, err := strconv.Atoi(string(bytes.TrimSpace(s[0]))); err == nil {
					c.statusCode = code
				}
			}

			continue
		}

		kv := bytes.SplitN(h, []byte{':'}, 2)
		if len(kv) != 2 {
			continue
		}

		switch strings.ToLower(strings.TrimSpace(string(kv[0]))) {
		case "content-length":
			length, err := strconv.Atoi(string(bytes.TrimSpace(kv[1])))
			if err != nil || length < 0 {
				return err
			}

			c.contentBytes = length
		case "transfer-encoding":
			if string(bytes.TrimSpace(kv[1])) == "chunked" {
				c.chunked = &bytes.Buffer{}
			}
		}
	}

	if c.contentBytes == -1 {
		c.contentBytes = math.MaxInt32
	}

	return nil
}

func (c *chunkedHttpParse) readChunked() (chunkedOver bool, err error) {
	rn := []byte{'\r', '\n'}
	for c.buf.Len() > 0 {
		all := c.buf.Bytes()
		p := bytes.Index(all, rn)
		if p < 0 {
			break
		}

		size, err := strconv.ParseUint(string(bytes.TrimSpace(all[:p])), 16, 32)
		if err != nil || size < 0 {
			return false, err
		}

		if len(all) >= p+len(rn)+int(size)+len(rn) {
			c.buf.Next(p + len(rn))
			if size > 0 {
				c.chunked.Write(c.buf.Next(int(size)))
			}
			c.buf.Next(len(rn))
			if size == 0 {
				return true, nil
			}
		} else {
			break
		}
	}

	return false, nil
}
