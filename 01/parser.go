package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

type parser struct {
	r      io.Reader
	length int
}

func (p *parser) init(src []byte) {
	p.r = bytes.NewReader(src)
	p.length = len(src)
}

func (p *parser) parsePerFileHeader() PcapHeader {
	var h PcapHeader
	binary.Read(p.r, binary.LittleEndian, &h)
	return h
}

func (p *parser) parseLinkLayer() (LinkLayer, uint8, error) {
	// Cheating here because I know the link layer is Ethernet
	var ether Ethernet
	if err := binary.Read(p.r, binary.BigEndian, &ether); err != nil {
		return nil, 0, err
	}

	//fmt.Printf("%+v\n", ether)

	return &ether, 14, nil
}

func (p *parser) parseNetworkLayer() (NetworkLayer, uint8, error) {
	// cheating here because I know the network layer is IPv4
	var ip IPv4
	if err := binary.Read(p.r, binary.BigEndian, &ip); err != nil {
		return nil, 0, err
	}

	size := ip.IHL()

	if size > 20 {
		// read IP options
		// cheating again as I know we don't have any options
	}

	//fmt.Printf("%+v\n", ip)

	return &ip, size, nil
}

func (p *parser) parseTCPTransportLayer() (*TCP, uint8, error) {
	var tcp TCP
	if err := binary.Read(p.r, binary.BigEndian, &tcp); err != nil {
		return nil, 0, err
	}

	size := tcp.DataOffset()

	if size > 20 {
		// TODO: parse TCP options

		io.CopyN(ioutil.Discard, p.r, int64(size-20))
	}

	//	fmt.Printf("%+v\n", tcp)

	return &tcp, size, nil
}

func (p *parser) parseNextPacket() (*Packet, error) {
	var pkt Packet
	var err error
	if err = binary.Read(p.r, binary.LittleEndian, &pkt.Header); err != nil {
		return nil, err
	}

	var lSize uint8
	if pkt.LinkLayer, lSize, err = p.parseLinkLayer(); err != nil {
		return nil, err
	}

	var nSize uint8
	if pkt.NetworkLayer, nSize, err = p.parseNetworkLayer(); err != nil {
		return nil, err
	}

	var tSize uint8
	switch pkt.NetworkLayer.TransportProtocol() {
	case 0x06: // TCP
		pkt.TransportLayer, tSize, err = p.parseTCPTransportLayer()
	default:
		return nil, errors.New("unknown network layer")
	}
	if err != nil {
		return nil, err
	}

	aSize := pkt.Header.DataLength - uint32(lSize) - uint32(nSize) - uint32(tSize)

	pkt.ApplicationLayer = make([]byte, aSize)

	if err = binary.Read(p.r, binary.BigEndian, &pkt.ApplicationLayer); err != nil {
		return nil, err
	}

	return &pkt, nil
}

func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			return ioutil.ReadAll(s)
		}
		return nil, errors.New("invalid source")
	}
	return ioutil.ReadFile(filename)
}

func ParseFile(filename string, src interface{}) (*PcapFile, error) {
	source, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	p.init(source)

	var file PcapFile
	file.Header = p.parsePerFileHeader()

	for err != io.EOF {
		var pkt *Packet
		pkt, err = p.parseNextPacket()
		if pkt != nil {
			file.Packets = append(file.Packets, *pkt)
		}
	}

	fmt.Println(len(file.Packets))
	return &file, nil
}
