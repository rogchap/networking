package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"golang.org/x/sys/unix"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var qtypes = map[string]Qtype{
	"A":   Qtype(A),
	"NS":  Qtype(NS),
	"TXT": Qtype(TXT),
}

type Digger struct {
	fd      int // socket file discriptor
	dstaddr *unix.SockaddrInet4
}

func (d *Digger) Init(server string) error {
	var err error
	if d.fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP); err != nil {
		return fmt.Errorf("unable to create socket: %w", err)
	}

	// quick way to parse an IPv4 address
	ip := net.ParseIP(server).To4()
	d.dstaddr = &unix.SockaddrInet4{Port: 53}
	copy(d.dstaddr.Addr[:], ip)
	return nil
}

func (d *Digger) Close() error {
	return unix.Close(d.fd)
}

func generateID() uint16 {
	return uint16(r.Uint32())
}

func askQuestion(host string, qtype Qtype) Message {
	return Message{
		header: &Header{
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

func (d *Digger) Dig(host, qtype string) (*Message, error) {
	qt, ok := qtypes[qtype]
	if !ok {
		qt = Qtype(A)
	}

	msg := askQuestion(host, qt)
	if err := unix.Sendto(d.fd, msg.Marshall(), 0, d.dstaddr); err != nil {
		return nil, err
	}

	done := make(chan struct{}, 1)
	recvErr := make(chan error, 1)

	// 1500 bytes is the size of an Enternet frame
	b := make([]byte, 1500)

	go func() {
		if _, _, err := unix.Recvfrom(d.fd, b, 0); err != nil {
			recvErr <- err
			return
		}
		done <- struct{}{}
	}()

	select {
	case err := <-recvErr:
		if err != nil {
			return nil, err
		}
	case <-time.After(5 * time.Second):
		return nil, errors.New("connection timed out")
	case <-done:
	}

	var parser Parser
	parser.Init(b)
	return parser.Parse()
}
