package conversation

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apparentlymart/gcc-module-cmi-builder/internal/protocol"
)

type Conversation struct {
	in  protocol.Parser
	out protocol.Writer

	crossCompile string
}

func New(in protocol.Parser, out protocol.Writer) *Conversation {
	return &Conversation{
		in:  in,
		out: out,
	}
}

// Run holds a conversation with a compiler over the given parser and
// writer, handling incoming messages and returning responses until the
// client (the compiler) closes our incoming message stream or until we
// encounter an I/O error.
func (c *Conversation) Run(in protocol.Parser, out protocol.Writer) error {
	handshook := false
	var compilerName string
	var buildIdent string

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
				repoPath, err := filepath.Abs(".modules-cmi")
				if err != nil {
					respBlock = append(respBlock, protocol.ErrorMessage("failed to build repository path: %s", err))
					continue
				}
				err = os.MkdirAll(repoPath, 0755)
				if err != nil {
					respBlock = append(respBlock, protocol.ErrorMessage("failed to create repository dir: %s", err))
					continue
				}
				respBlock = append(respBlock, protocol.PathNameMessage(repoPath))
			case "MODULE-EXPORT":
				if len(reqMsg) < 2 {
					respBlock = append(respBlock, protocol.ErrorMessage("invalid MODULE-EXPORT message"))
					continue
				}
				name := reqMsg[1]
				cmiFileRel := c.cmiFilename(name, compilerName, buildIdent)
				respBlock = append(respBlock, protocol.PathNameMessage(cmiFileRel))
			case "MODULE-IMPORT":
				if len(reqMsg) < 2 {
					respBlock = append(respBlock, protocol.ErrorMessage("invalid MODULE-IMPORT message"))
					continue
				}
				name := reqMsg[1]
				cmiFileRel := c.cmiFilename(name, compilerName, buildIdent)
				cmiDirRel := filepath.Dir(cmiFileRel)
				cmiDir := filepath.Join(".modules-cmi", cmiDirRel)
				cmiFile := filepath.Join(".modules-cmi", cmiFileRel)
				err = os.MkdirAll(cmiDir, 0755)
				if err != nil {
					respBlock = append(respBlock, protocol.ErrorMessage("failed to create repository dir: %s", err))
					continue
				}

				sourceFile, mode := sourceForModule(name)

				cmiInfo, err := os.Stat(cmiFile)
				if err != nil {
					if !os.IsNotExist(err) {
						respBlock = append(respBlock, protocol.ErrorMessage("invalid existing CMI file %s: %s", cmiFile, err))
						continue
					}
					cmiInfo = nil
				}
				if cmiInfo != nil {
					sourceInfo, err := os.Stat(sourceFile)
					if err != nil {
						respBlock = append(respBlock, protocol.ErrorMessage("invalid source file %s: %s", sourceFile, err))
						continue
					}

					if cmiInfo.ModTime().After(sourceInfo.ModTime()) {
						// The CMI file is already up-to-date, so we can just
						// return it without doing any more builds.
						respBlock = append(respBlock, protocol.PathNameMessage(cmiFileRel))
						continue
					}
				}

				cmdLine := c.gccBaseArgs(mode)
				cmdLine = append(cmdLine, sourceFile)
				log.Printf("compile %#v", cmdLine)
				cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
				cmd.Env = os.Environ()
				cmd.Stdout = os.Stderr // don't disturb the protocol on stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				if err != nil {
					respBlock = append(respBlock, protocol.ErrorMessage("failed to build %s: %s", name, err))
					continue
				}

				respBlock = append(respBlock, protocol.PathNameMessage(cmiFileRel))
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

func (c *Conversation) SetCrossCompilePrefix(prefix string) {
	c.crossCompile = prefix
}

func (c *Conversation) cmiFilename(mod string, compilerName string, buildIdent string) string {
	hash := sha256.New()
	hash.Write([]byte(mod))
	if c.crossCompile != "" {
		hash.Write([]byte{0})
		hash.Write([]byte(c.crossCompile))
	}
	if buildIdent != "" {
		return fmt.Sprintf("%x-%s-%s", hash.Sum(nil), buildIdent, compilerName)
	}
	key := fmt.Sprintf("%x-%s", hash.Sum(nil), compilerName)
	return filepath.Join(key[:2], key[2:4], key+".cmi")
}

type buildMode rune

const buildModuleUnit buildMode = 'M'
const buildHeaderUnit buildMode = 'H'

func (c *Conversation) gccBaseArgs(mode buildMode) []string {
	ret := make([]string, 0, 7)
	ret = append(ret, c.crossCompile+"g++", "-std=c++20", "-fmodule-mapper=|"+os.Args[0])
	switch mode {
	case buildModuleUnit:
		ret = append(ret, "-fmodules-ts", "-c")
	case buildHeaderUnit:
		ret = append(ret, "-fmodules-ts", "-x", "c++-system-header")
	default:
		panic("unsupported build mode")
	}
	return ret
}

func sourceForModule(mod string) (string, buildMode) {
	if filepath.IsAbs(mod) {
		return mod, buildHeaderUnit
	}
	if strings.HasPrefix(filepath.ToSlash(mod), "./") {
		return mod, buildHeaderUnit
	}

	return mod + ".cpp", buildModuleUnit
}
