package drones

import (
	"github.com/benbearchen/antpost"

	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

type NextHttp func(h *HttpReq, ok bool, statusCode int, header http.Header, data []byte) *HttpReq

type HttpReq struct {
	Url    string
	Method string
	Header http.Header
	Data   []byte
	Next   NextHttp
	Arg    interface{}
}

func NewHttpGetReq(url string, next NextHttp, arg interface{}) *HttpReq {
	return &HttpReq{url, "GET", nil, nil, next, arg}
}

func NewHttpPostReq(url string, data []byte, next NextHttp, arg interface{}) *HttpReq {
	return &HttpReq{url, "POST", nil, data, next, arg}
}

func NewHttpDrone(h *HttpReq) antpost.Drone {
	d := new(httpDrone)
	d.http = h
	return d
}

type httpDrone struct {
	http *HttpReq
	next *HttpReq
}

func (h *httpDrone) Run(context *antpost.Context) antpost.DroneResult {
	resp, err := h.http.req()
	context.Step(antpost.StepConnected)
	if err != nil {
		return antpost.ResultConnectFail
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	context.Step(antpost.StepResponsed)

	if h.http.Next != nil {
		ok := err == nil
		h.next = h.http.Next(h.http, ok, resp.StatusCode, resp.Header, data)
	}

	context.Bool("conn", err == nil)

	if err != nil {
		return antpost.ResultResponseBroken
	} else {
		return antpost.ResultOK
	}
}

func (h *httpDrone) Next() antpost.Drone {
	if h.next != nil {
		return NewHttpDrone(h.next)
	} else if h.http.Next != nil {
		return NewHttpDrone(h.http.Next(h.http, false, 0, nil, nil))
	} else {
		return h
	}
}

func (h *HttpReq) req() (*http.Response, error) {
	client := &http.Client{}

	var body io.Reader = nil
	if h.Method == "POST" {
		body = bytes.NewReader(h.Data)
	}

	req, err := http.NewRequest(h.Method, h.Url, body)
	if err != nil {
		return nil, err
	}

	if h.Header != nil {
		req.Header = h.Header
	}

	return client.Do(req)
}
