package main

import (
	"fmt"
	"os"
	"syscall"
)

type RequestHandler func(ResponseWriter, *Request)

type Server struct {
	fd      int
	Handler RequestHandler
}

func recvall(fd int) ([]byte, error) {

	// This is pretty rubbish!
	// I should really read the lines to get the content length
	// find where the headers end, aand wait until I've got all the
	// bytes I need/want.

	var ret []byte
	const chunk = 4 << 10 // 4kB
	for {
		b := make([]byte, chunk)
		n, _, err := syscall.Recvfrom(fd, b, 0)
		if err != nil {
			return ret, err
		}
		ret = append(ret, b[:n]...)

		// there is a chance I've not got all the data :(
		if n < chunk {
			break
		}
	}
	return ret, nil
}

func notFoundHandler(rw ResponseWriter, _ *Request) {
	rw.SetStatus(404)
}

func (s *Server) handleIncoming(fd int, sa syscall.Sockaddr) {
	defer syscall.Close(fd) // no keep-alive

	b, err := recvall(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error recvall: %v", err)
		return
	}

	req, err := parseRequest(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parseRequest: %v", err)
	}

	res := &Response{
		headers: make(map[string]string),
	}

	if s.Handler == nil {
		s.Handler = notFoundHandler
	}
	s.Handler(res, req)

	// reply to caller
	if err := syscall.Sendto(fd, res.raw(), 0, sa); err != nil {
		panic(err)
	}

}

// Serve listens and accepts messages on a given port
func (s *Server) Serve(port int) error {
	var err error
	if s.fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0); err != nil {
		return err
	}

	sa := &syscall.SockaddrInet4{
		Port: port,
		Addr: [...]byte{0, 0, 0, 0},
	}

	if err := syscall.Bind(s.fd, sa); err != nil {
		return err
	}

	if err := syscall.Listen(s.fd, 10); err != nil {
		return err
	}

	errc := make(chan error, 1)
	go func() {
		for {
			fd, sa, err := syscall.Accept(s.fd)
			if err != nil {
				errc <- err
			}
			go s.handleIncoming(fd, sa)
		}
	}()

	return <-errc
}

func (s *Server) Close() {
	syscall.Close(s.fd)
}
