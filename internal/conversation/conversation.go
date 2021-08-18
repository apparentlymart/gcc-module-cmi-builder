package conversation

import (
	"crypto/sha256"
	"fmt"

	"github.com/apparentlymart/gcc-module-cmi-builder/internal/protocol"
)

// Run holds a conversation with a compiler over the given parser and
// writer, handling incoming messages and returning responses until the
// client (the compiler) closes our incoming message stream or until we
// encounter an I/O error.
func Run(in protocol.Parser, out protocol.Writer) error {
	handshook := false
	var compilerName string
	var buildIdent string

	moduleKey := func(name string) string {
		hash := sha256.New()
		hash.Write([]byte(name))
		if buildIdent != "" {
			return fmt.Sprintf("%x-%s-%s", hash.Sum(nil), buildIdent, compilerName)
		}
		return fmt.Sprintf("%x-%s", hash.Sum(nil), compilerName)
	}

	for {
		reqBlock, err := in.NextBlock()
		if err != nil {
			return err
		}
		if len(reqBlock) == 0 {
			return nil // all done!
		}
		respBlock := make(protocol.Block, 0, len(reqBlock))

		for _, reqMsg := range reqBlock {
			if len(reqMsg) < 1 {
				return fmt.Errorf("invalid empty message from client")
			}

			if reqMsg[0] == "HELLO" {
				if len(reqMsg) < 4 {
					respBlock = append(respBlock, protocol.ErrorMessage("unsupported handshake format"))
					continue
				}

				compilerName = reqMsg[2]
				buildIdent = reqMsg[3]
				handshook = true

				respBlock = append(respBlock, protocol.HandshakeResponseMessage(1, "gcc-module-cmi-builder", 0))

				continue
			}

			if !handshook {
				// All of the other messages require a handshake first
				respBlock = append(respBlock, protocol.ErrorMessage("message before valid handshake"))
				continue
			}

			switch reqMsg[0] {
			case "MODULE-REPO":
				respBlock = append(respBlock, protocol.PathNameMessage(""))
			case "MODULE-IMPORT":
				if len(reqMsg) < 2 {
					respBlock = append(respBlock, protocol.ErrorMessage("invalid MODULE-IMPORT message"))
					continue
				}
				name := reqMsg[1]
				key := moduleKey(name)
				// TODO: Properly implement
				respBlock = append(respBlock, protocol.ErrorMessage("don't yet know how to build module %q as %s", name, key))
			default:
				respBlock = append(respBlock, protocol.ErrorMessage("unsupported message type %q", reqMsg[0]))
			}
		}

		// Once we get here we should have enough entries in respBlock to
		// match all of the requests in reqBlock, so we can respond.
		err = out.WriteBlock(respBlock)
		if err != nil {
			return err
		}
	}
}
