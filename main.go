package main

import (
	"log"
	"os"

	"github.com/apparentlymart/gcc-module-cmi-builder/internal/conversation"
	"github.com/apparentlymart/gcc-module-cmi-builder/internal/protocol"
)

func main() {
	p := protocol.NewParser(os.Stdin)
	w := protocol.NewWriter(os.Stdout)

	c := conversation.New(p, w)

	if cc := os.Getenv("CROSS_COMPILE"); cc != "" {
		c.SetCrossCompilePrefix(cc)
	}

	err := c.Run(p, w)
	if err != nil {
		log.Fatal(err)
	}

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
