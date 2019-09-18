package main

type Message struct {
	header      Header
	questions   []Question
	answers     []ResourceRecord
	authorities []ResourceRecord
	aditional   []ResourceRecord
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
	QSTAR // * All records
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
	CSTAR Qclass = 255 // * any class
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

type Question struct {
	qname  string
	qtype  Qtype
	qclass Qclass
}

type ResourceRecord struct {
	name     string
	typ      Type
	class    Class
	ttl      uint32
	rdlength uint16
	rdata    []byte // This may be better as an interface
}
