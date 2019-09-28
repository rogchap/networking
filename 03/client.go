package main

import "syscall"

type Client struct{}

func (c *Client) Do(r *Request) (*Response, error) {

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.Close(fd)

	// TODO get fro r.Host
	sa := &syscall.SockaddrInet4{
		Port: 9000,
		Addr: [...]byte{0, 0, 0, 0},
	}

	if err := syscall.Connect(fd, sa); err != nil {
		return nil, err
	}

	if err := syscall.Sendto(fd, r.raw(), 0, sa); err != nil {
		return nil, err
	}

	b, err := recvall(fd)
	if err != nil {
		return nil, err
	}

	return parseResponse(b)

}
