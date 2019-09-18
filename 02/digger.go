package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type digger struct {
	fd      int // socket file discriptor
	srcaddr unix.Sockaddr
}

func (d *digger) init() error {
	var err error
	if d.fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP); err != nil {
		return fmt.Errorf("unable to create socket: %w", err)
	}

	d.srcaddr = &unix.SockaddrInet4{}
	if err = unix.Bind(d.fd, d.srcaddr); err != nil {
		fmt.Errorf("unable to bind to socket: %w", err)
	}
	return nil
}

func (d *digger) dig(host string) error {
	return nil
}
