package main

import (
	"fmt"
	"log"
	"os"

	"github.com/apparentlymart/gcc-module-cmi-builder/internal/protocol"
)

func main() {
	p := protocol.NewParser(os.Stdin)
	for {
		block, err := p.NextBlock()
		if err != nil {
			log.Fatal(err)
		}
		if len(block) == 0 {
			break
		}
		fmt.Printf("%#v\n", block)
	}
}
