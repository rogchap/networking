package main

import (
	"encoding/hex"
	"fmt"
	"net"
)

type PcapHeader struct {
	MagicNumber         uint32
	MajorVersion        uint16
	MinorVersion        uint16
	TimezoneOffset      uint32
	TimestampAccuracy   uint32
	SnapshotLength      uint32
	LinkLayerHeaderType uint32
}

type PacketHeader struct {
	TimestampSeconds  uint32
	TimestampNanos    uint32
	DataLength        uint32
	UntruncatedLength uint32
}

type LinkLayer interface {
	linkLayerNode()
}

type MACAddress [6]byte

func (m MACAddress) String() string {
	buf := make([]byte, 17)
	hex.Encode(buf[0:2], m[:1])
	buf[2] = ':'
	hex.Encode(buf[3:5], m[1:2])
	buf[5] = ':'
	hex.Encode(buf[6:8], m[2:3])
	buf[8] = ':'
	hex.Encode(buf[9:11], m[3:4])
	buf[11] = ':'
	hex.Encode(buf[12:14], m[4:5])
	buf[14] = ':'
	hex.Encode(buf[15:], m[5:])
	return string(buf)
}

type EtherType uint16

func (e EtherType) String() string {
	switch e {
	case 0x0800:
		return "IPv4"
	case 0x86DD:
		return "IPv6"
	default:
		return "Other" // https://en.wikipedia.org/wiki/EtherType
	}
}

type Ethernet struct {
	DestinationMAC MACAddress
	SourceMAC      MACAddress
	Type           EtherType
}

func (Ethernet) linkLayerNode() {}

type NetworkLayer interface {
	networkLayerNode()
	TransportProtocol() Protocol
}

type IPv4Address [4]byte

func (i IPv4Address) String() string {
	return net.IPv4(i[0], i[1], i[2], i[3]).String()
}

type Protocol uint8

func (p Protocol) String() string {
	switch p {
	case 0x06:
		return "TCP"
	case 0x11:
		return "UDP"
	default:
		return "Other" // https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
	}
}

type VersionAndIHL uint8

func (v VersionAndIHL) Version() uint8 {
	return uint8(v) & 0xF0 >> 4
}

func (v VersionAndIHL) IHL() uint8 {
	return uint8(v) & 0x0F * 4
}

func (v VersionAndIHL) String() string {
	return fmt.Sprintf("version:%d, length: %d", v.Version(), v.IHL())
}

type IPv4 struct {
	VersionAndIHL          VersionAndIHL
	DSCPAndECN             uint8
	TotalLength            uint16
	Identification         uint16
	FlagsAndFragmentOffset uint16
	TimeToLive             uint8
	Protocol               Protocol
	HeaderChecksum         uint16
	SourceIPAddress        IPv4Address
	DestinationIPAddress   IPv4Address
}

func (i IPv4) Version() uint8 {
	return i.VersionAndIHL.Version()
}

func (i IPv4) IHL() uint8 {
	return i.VersionAndIHL.IHL()
}

func (IPv4) networkLayerNode() {}

func (i *IPv4) TransportProtocol() Protocol {
	return i.Protocol
}

type TransportLayer interface {
	transportLayerNode()
}

type DataOffsetAndNSFlag uint8

func (d DataOffsetAndNSFlag) DataOffset() uint8 {
	return uint8(d) & 0xF0 >> 4 * 4
}

type TCP struct {
	SourcePort           uint16
	DestinationPort      uint16
	SequenceNumber       uint32
	AcknowledgmentNumber uint32
	DataOffsetAndFlags   DataOffsetAndNSFlag
	Flags                uint8
	WindowSize           uint16
	Checksum             uint16
	UrgentPointer        uint16
}

func (t TCP) DataOffset() uint8 {
	return t.DataOffsetAndFlags.DataOffset()
}

func (TCP) transportLayerNode() {}

type Packet struct {
	Header           PacketHeader
	LinkLayer        LinkLayer
	NetworkLayer     NetworkLayer
	TransportLayer   TransportLayer
	ApplicationLayer []byte
}

type PcapFile struct {
	Header  PcapHeader
	Packets []Packet
}
