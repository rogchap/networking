package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

type ResponseWriter interface {
	io.Writer
	Headers() Headers
	SetStatus(code int)
}

type Response struct {
	headers Headers
	body    bytes.Buffer
	status  int
}

var statusCodes = map[int]string{
	200: "OK",
	404: "Not Found",
}

func (r *Response) Write(p []byte) (n int, err error) {
	return r.body.Write(p)
}

func (r *Response) Headers() Headers {
	return r.headers
}

func (r *Response) SetStatus(code int) {
	r.status = code
}

func (r *Response) Body() []byte {
	return r.body.Bytes()
}

func (r *Response) raw() []byte {
	if r.status == 0 {
		r.status = 200
	}
	var b bytes.Buffer

	// start-line
	b.WriteString("HTTP/1.1 ")
	b.WriteString(strconv.Itoa(r.status))
	b.WriteRune(' ')
	b.WriteString(statusCodes[r.status])
	b.WriteRune('\r')
	b.WriteRune('\n')

	for k, v := range r.headers {
		b.WriteString(k)
		b.WriteRune(':')
		b.WriteRune(' ')
		b.WriteString(v)
		b.WriteRune('\r')
		b.WriteRune('\n')
	}

	b.WriteRune('\r')
	b.WriteRune('\n')

	b.Write(r.Body())

	return b.Bytes()
}

func parseResponse(b []byte) (*Response, error) {
	res := Response{
		headers: make(map[string]string),
	}
	buf := bytes.NewBuffer(b)

	// start-line
	sl, err := buf.ReadString('\n')
	if err != nil {
		return nil, err
	}

	slprt := strings.Split(sl, " ")
	res.status, _ = strconv.Atoi(slprt[1])

	// scan for the headers
	for {
		txt, _ := buf.ReadString('\n')
		if txt == "\r\n" {
			break
		}
		// this is is not that performant... but works
		header := strings.Split(txt, ":")
		res.headers.Add(header[0], header[1])
	}

	body, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	res.Write(body)

	return &res, nil
}
