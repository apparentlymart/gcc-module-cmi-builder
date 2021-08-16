package main

import (
	"log"
	"os"

	"github.com/apparentlymart/gcc-module-cmi-builder/internal/protocol"
)

func main() {
	p := protocol.NewParser(os.Stdin)
	w := protocol.NewWriter(os.Stdout)
	for {
		block, err := p.NextBlock()
		if err != nil {
			log.Fatal(err)
		}
		if len(block) == 0 {
			break
		}
		err = w.WriteBlock(block)
		if err != nil {
			log.Fatal(err)
		}
	}
}
