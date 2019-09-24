package main

import (
	"os"
	"os/signal"
	"syscall"
)

func echo(fd int, sa syscall.Sockaddr) {
	defer syscall.Close(fd)

	b := make([]byte, 1500)
	if _, _, err := syscall.Recvfrom(fd, b, 0); err != nil {
		panic(err)
	}

	if err := syscall.Sendto(fd, b, 0, sa); err != nil {
		panic(err)
	}
}

func main() {

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(fd)

	sa := &syscall.SockaddrInet4{
		Port: 8080,
		Addr: [...]byte{0, 0, 0, 0},
	}

	if err := syscall.Bind(fd, sa); err != nil {
		panic(err)
	}

	if err := syscall.Listen(fd, 1); err != nil {
		panic(err)
	}

	go func() {
		for {
			nfd, dsa, err := syscall.Accept(fd)
			if err != nil {
				panic(err)
			}
			go echo(nfd, dsa)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-stop

}
