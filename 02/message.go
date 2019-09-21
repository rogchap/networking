//go:generate stringer -type=Type,Class

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type Message struct {
	header      *Header
	questions   []Question
	answers     []ResourceRecord
	authorities []ResourceRecord
	additional  []ResourceRecord
}

func (m Message) Marshall() []byte {
	b := m.header.Marshall()

	for _, q := range m.questions {
		b = append(b, q.Marshall()...)
	}

	// Skiping other parts as not needed to construct a query

	return b
}

type Opcode int

const (
	QUERY Opcode = iota
	IQUERY
	STATUS
)

type Rcode int

const (
	NO_ERR Rcode = iota
	FORMAT_ERR
	SERVER_FAIL
	NAME_ERR
	NOT_IMPLEMENTED
	REFUSED
)

type Type int

const (
	A Type = iota + 1
	NS
	MD
	MF
	CNAME
	SOA
	MB
	MG
	MR
	NULL
	WKS
	PTR
	HINFO
	MINFO
	MX
	TXT
)

type Qtype Type

const (
	AXFR Qtype = iota + 252
	MAILB
	MAILA
	ANY // * All records
)

type Class int

const (
	IN Class = iota + 1
	CS
	CH
	HS
)

type Qclass Class

const (
	CANY Qclass = 255 // * any class
)

type Header struct {
	id      uint16
	qr      bool // query (0) or response (1)
	opcode  Opcode
	aa      bool // authoritative answer
	tc      bool // truncation
	rd      bool // recursion desired
	ra      bool // recursion available
	z       bool // 0b000
	rcode   Rcode
	qdcount uint16
	ancount uint16
	nscount uint16
	arcount uint16
}

func (h Header) Marshall() []byte {
	b := &bytes.Buffer{}
	binary.Write(b, binary.BigEndian, h.id)

	flags := uint16(h.opcode)<<11 | uint16(h.rcode&0xF)

	if h.rd {
		flags |= 1 << 8
	}
	// TODO: Set the other flags in the same way
	binary.Write(b, binary.BigEndian, flags)
	binary.Write(b, binary.BigEndian, h.qdcount)
	binary.Write(b, binary.BigEndian, h.ancount)
	binary.Write(b, binary.BigEndian, h.nscount)
	binary.Write(b, binary.BigEndian, h.arcount)

	return b.Bytes()
}

type Question struct {
	qname  string
	qtype  Qtype
	qclass Qclass
}

func (q Question) Marshall() []byte {
	b := &bytes.Buffer{}

	names := strings.Split(q.qname, ".")
	for _, n := range names {
		qname := []byte(n)
		binary.Write(b, binary.BigEndian, uint8(len(qname)))
		binary.Write(b, binary.BigEndian, qname)
	}
	binary.Write(b, binary.BigEndian, uint8(0))
	binary.Write(b, binary.BigEndian, uint16(q.qtype))
	binary.Write(b, binary.BigEndian, uint16(q.qclass))
	return b.Bytes()
}

type RData interface {
	fmt.Stringer
	rDataNode()
}

type IPv4Address [4]byte

func (i IPv4Address) String() string {
	return net.IPv4(i[0], i[1], i[2], i[3]).String()
}

type ARecord struct {
	Addr IPv4Address
}

func (*ARecord) rDataNode() {}
func (a *ARecord) String() string {
	return a.Addr.String()
}

type ResourceRecord struct {
	name     string
	typ      Type
	class    Class
	ttl      uint32
	rdlength uint16
	rdata    RData
}
