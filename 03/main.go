package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func handler(fd int, sa syscall.Sockaddr) {
	defer syscall.Close(fd)

	// get mesage from caller
	b := make([]byte, 1500)
	if _, _, err := syscall.Recvfrom(fd, b, 0); err != nil {
		panic(err)
	}

	// proxy the message
	nfd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(nfd)

	nsa := &syscall.SockaddrInet4{
		Port: 9000,
		Addr: [...]byte{0, 0, 0, 0},
	}
	if err := syscall.Connect(nfd, nsa); err != nil {
		panic(err)
	}
	if err := syscall.Sendto(nfd, b, 0, nsa); err != nil {
		panic(err)
	}

	res := make([]byte, 1500)
	if _, _, err := syscall.Recvfrom(nfd, res, syscall.MSG_WAITALL); err != nil {
		panic(err)
	}
	// respond to the caller
	if err := syscall.Sendto(fd, res, 0, sa); err != nil {
		panic(err)
	}
}

func proxy(rw ResponseWriter, r *Request) {
	// TODO change req Host to be that of end server
	var c Client
	res, err := c.Do(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in client: %v", err)
		return
	}

	// TODO I should copy the headers from client res to rw
	rw.Write(res.Body())
}

func main() {
	var s Server
	s.Handler = proxy

	go s.Serve(8080)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-stop

	s.Close()
	os.Exit(0)
}
