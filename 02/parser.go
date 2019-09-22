package main

import (
	"encoding/binary"
	"strings"
)

type Parser struct {
	src    []byte
	offset int
	names  map[int]string // message compression lookup
}

func (p *Parser) Init(src []byte) {
	p.src = src
	p.offset = 0
	p.names = make(map[int]string)
}

func (p *Parser) nextOctet() uint8 {
	n := p.src[p.offset]
	p.offset += 1
	return uint8(n)
}

func (p *Parser) nextTwoOctets() uint16 {
	n := p.src[p.offset : p.offset+2]
	p.offset += 2
	return binary.BigEndian.Uint16(n)
}

func (p *Parser) nextNOctets(n int) []byte {
	b := p.src[p.offset : p.offset+n]
	p.offset += n
	return b
}

func (p *Parser) parseHeader() (*Header, error) {
	var h Header
	h.id = p.nextTwoOctets()

	bits := p.nextTwoOctets()
	h.qr = (bits>>15)&1 == 1
	h.opcode = Opcode((bits << 1) >> 12)
	h.aa = (bits>>10)&1 == 1
	h.tc = (bits>>9)&1 == 1
	h.rd = (bits>>8)&1 == 1
	h.ra = (bits>>7)&1 == 1
	h.rcode = Rcode(bits & 0x0F)

	h.qdcount = p.nextTwoOctets()
	h.ancount = p.nextTwoOctets()
	h.nscount = p.nextTwoOctets()
	h.arcount = p.nextTwoOctets()

	return &h, nil
}

func (p *Parser) parseDomainName() string {
	var sb strings.Builder
	offset := p.offset
	for o := p.nextOctet(); o != 0; o = p.nextOctet() {
		n := o &^ 0xc0
		if o&0xc0 == 0xc0 {
			dnOffset := binary.BigEndian.Uint16([]byte{n, p.nextOctet()})
			sb.WriteString(p.names[int(dnOffset)])
			break
		}
		sb.Write(p.nextNOctets(int(n)))
		sb.WriteRune('.')
	}
	dn := sb.String()
	p.names[offset] = dn
	return dn
}

func (p *Parser) parseQuestion() Question {
	var q Question
	q.qname = p.parseDomainName()
	q.qtype = Qtype(p.nextTwoOctets())
	q.qclass = Qclass(p.nextTwoOctets())
	return q
}

func (p *Parser) parseARecord() *ARecord {
	var ar ARecord
	copy(ar.Addr[:], p.nextNOctets(4))
	return &ar
}

func (p *Parser) parseNSRecord() *NSRecord {
	var ns NSRecord
	ns.Name = p.parseDomainName()
	return &ns
}

func (p *Parser) parseTXTRecord(l uint16) *TXTRecord {
	var t TXTRecord
	t.Text = string(p.nextNOctets(int(l)))
	return &t
}

func (p *Parser) parseResourceRecord() ResourceRecord {
	var rr ResourceRecord
	rr.name = p.parseDomainName()
	rr.typ = Type(p.nextTwoOctets())
	rr.class = Class(p.nextTwoOctets())
	rr.ttl = binary.BigEndian.Uint32(p.nextNOctets(4))
	rr.rdlength = p.nextTwoOctets()

	switch rr.typ {
	case A:
		rr.rdata = p.parseARecord()
	case NS:
		rr.rdata = p.parseNSRecord()
	case TXT:
		rr.rdata = p.parseTXTRecord(rr.rdlength)
	default:
		panic("rdata type not implemented")
	}
	return rr
}

func (p *Parser) Parse() (*Message, error) {
	var msg Message
	var err error

	if msg.header, err = p.parseHeader(); err != nil {
		return nil, err
	}

	for i := uint16(0); i < msg.header.qdcount; i++ {
		msg.questions = append(msg.questions, p.parseQuestion())
	}

	for i := uint16(0); i < msg.header.ancount; i++ {
		msg.answers = append(msg.answers, p.parseResourceRecord())
	}

	return &msg, nil

}
