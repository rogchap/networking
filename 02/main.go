package main

import (
	"fmt"
	"os"
)

func main() {
	host := "."

	for idx, arg := range os.Args {
		switch idx {
		case 1:
			host = arg
		}
	}

	var digger Digger
	defer digger.Close()

	if err := digger.Init(); err != nil {
		panic(err)
	}

	msg, err := digger.Dig(host)
	if err != nil {
		panic(err)
	}

	for _, an := range msg.answers {
		fmt.Printf("%s\t%d\t%s\t%s\t%s\n", an.name, an.ttl, an.class, an.typ, an.rdata)
	}
}
