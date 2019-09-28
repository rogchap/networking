package main

import (
	"bytes"
	"io/ioutil"
	"strings"
)

const (
	OPTIONS = "OPTIONS"
	HEAD    = "HEAD"
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	PATCH   = "PATCH"
	DELETE  = "DELETE"
)

type Headers map[string]string

func (h Headers) Add(key, value string) {
	// just asuming that we have 1:1 header, but if we had more
	// than one header with the same key (like set-cookie) we should join the headers
	h[key] = strings.TrimSpace(value)
}

type Request struct {
	Verb    string
	Host    string
	Path    string
	version string
	Headers Headers
	Body    []byte
}

func (r *Request) raw() []byte {
	var b bytes.Buffer

	// start-line
	b.WriteString(r.Verb)
	b.WriteRune(' ')
	b.WriteString(r.Path)
	b.WriteRune(' ')
	b.WriteString(r.version)
	b.WriteRune('\r')
	b.WriteRune('\n')

	// headers
	for k, v := range r.Headers {
		b.WriteString(k)
		b.WriteRune(':')
		b.WriteRune(' ')
		b.WriteString(v)
		b.WriteRune('\r')
		b.WriteRune('\n')
	}

	b.WriteRune('\r')
	b.WriteRune('\n')

	// body
	switch r.Verb {
	case POST, PUT, PATCH:
		b.Write(r.Body)
	default:
	}

	return b.Bytes()
}

func parseRequest(b []byte) (*Request, error) {
	req := Request{
		Headers: make(map[string]string),
	}
	buf := bytes.NewBuffer(b)

	// start-line
	sl, err := buf.ReadString('\n')
	if err != nil {
		return nil, err
	}
	// not the best, but lets asume that the HTTP protocol is well defined in this request
	slprt := strings.Split(sl, " ")
	req.Verb = slprt[0]
	req.Path = slprt[1]
	req.version = strings.TrimSpace(slprt[2])

	// scan for the headers
	for {
		txt, _ := buf.ReadString('\n')
		if txt == "\r\n" {
			break
		}
		// this is is not that performant... but works
		header := strings.SplitN(txt, ":", 2)
		req.Headers.Add(header[0], header[1])

		if strings.ToLower(header[0]) == "host" {
			req.Host = header[1]
		}
	}

	if req.Body, err = ioutil.ReadAll(buf); err != nil {
		return nil, err
	}

	return &req, nil
}
