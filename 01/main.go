package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
)

type ByTCPSequence []Packet

func (s ByTCPSequence) Len() int      { return len(s) }
func (s ByTCPSequence) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByTCPSequence) Less(i, j int) bool {
	// cheating because I know they are all TCP
	return (s[i].TransportLayer.(*TCP)).SequenceNumber < (s[j].TransportLayer.(*TCP)).SequenceNumber
}

func main() {
	f, err := ParseFile(os.Args[1], nil)
	if err != nil {
		fmt.Errorf("failed to parse file", err)
		os.Exit(1)
	}

	// sort the packets
	sort.Sort(ByTCPSequence(f.Packets))

	var httpData []byte
	dup := make(map[uint32]struct{})

	for _, pkt := range f.Packets {
		tcp := pkt.TransportLayer.(*TCP)

		if tcp.Flags.SYN() {
			continue
		}

		sn := tcp.SequenceNumber
		if _, ok := dup[sn]; ok {
			continue
		}
		dup[sn] = struct{}{}

		httpData = append(httpData, pkt.ApplicationLayer...)

		if tcp.Flags.FIN() {
			break
		}
	}

	resIdx := bytes.Index(httpData, []byte("\r\n\r\n"))

	fmt.Println(string(httpData[:resIdx]))

	imgFile, _ := os.Create("image.jpg")
	imgFile.Write(httpData[resIdx+4:])
}
