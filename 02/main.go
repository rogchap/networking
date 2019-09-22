package main

import (
	"fmt"
	"os"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: go run . [@dns-server] [q-type] host\n")
	os.Exit(1)
}

func main() {
	args := os.Args[1:]
	lenArgs := len(args)
	if lenArgs == 0 || lenArgs > 3 {
		usage()
	}

	dnsServer := "1.1.1.1"
	qtype := "A"
	host := "." // Root

	for i := lenArgs - 1; i >= 0; i-- {
		if i == lenArgs-1 {
			host = args[i]
			continue
		}
		if strings.HasPrefix(args[i], "@") {
			dnsServer = string(args[i][1:])
			break
		}
		qtype = args[i]
	}

	var digger Digger
	defer digger.Close()

	if err := digger.Init(dnsServer); err != nil {
		panic(err)
	}

	msg, err := digger.Dig(host, qtype)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, an := range msg.answers {
		fmt.Printf("%s\t%d\t%s\t%s\t%s\n", an.name, an.ttl, an.class, an.typ, an.rdata)
	}
}
