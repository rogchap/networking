package main

import (
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

	var digger digger
	if err := digger.init(); err != nil {
		panic(err)
	}

	err := digger.dig(host)
	if err != nil {
		panic(err)
	}
}
