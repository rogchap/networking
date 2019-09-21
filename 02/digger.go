package main

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sys/unix"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type digger struct {
	fd      int // socket file discriptor
	srcaddr unix.Sockaddr
	dstaddr unix.Sockaddr
}

func (d *digger) init() error {
	var err error
	if d.fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP); err != nil {
		return fmt.Errorf("unable to create socket: %w", err)
	}

	d.srcaddr = &unix.SockaddrInet4{}
	/*	if err = unix.Bind(d.fd, d.srcaddr); err != nil {
			return fmt.Errorf("unable to bind to socket: %w", err)
		}
	*/
	d.dstaddr = &unix.SockaddrInet4{
		Port: 53,
		Addr: [4]byte{0x8, 0x8, 0x8, 0x8},
	}
	return nil
}

func (d *digger) close() error {
	return unix.Close(d.fd)
}

func generateID() uint16 {
	return uint16(r.Uint32())
}

func askQuestion(host string, qtype Qtype) Message {
	return Message{
		header: Header{
			id:      generateID(),
			rd:      true,
			qdcount: 1,
		},
		questions: []Question{
			Question{
				qname:  host,
				qtype:  qtype,
				qclass: Qclass(IN),
			},
		},
	}
}

func (d *digger) dig(host string) error {

	msg := askQuestion(host, Qtype(A))

	if err := unix.Sendto(d.fd, msg.Marshall(), 0, d.dstaddr); err != nil {
		return err
	}

	// 1500 bytes is the size of an Enternet frame
	b := make([]byte, 1500)
	if _, _, err := unix.Recvfrom(d.fd, b, 0); err != nil {
		return err
	}

	fmt.Printf("% x\n", b)

	return nil
}
