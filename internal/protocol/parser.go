package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var tokenRE = regexp.MustCompile(
	`^[ \t]*(\n|;|'(?:\\'|[^'])*'|[-+_/%.A-Za-z0-9]+)`,
)

type Message []string

type Block []Message

type Parser struct {
	sc *bufio.Scanner
}

func NewParser(r io.Reader) Parser {
	sc := bufio.NewScanner(r)
	sc.Split(splitFunc)
	return Parser{sc}
}

func (p Parser) NextBlock() (Block, error) {
	continuation := false
	var words []string
	var messages []Message
	for p.sc.Scan() {
		tok := p.sc.Bytes()
		if len(tok) == 1 && tok[0] == '\n' {
			if continuation {
				words = words[:len(words)-1] // trim continuation marker
			}
			messages = append(messages, Message(words))
			words = words[len(words):]
			if !continuation {
				return Block(messages), nil
			}
			continuation = false
			continue
		}
		continuation = false
		word := parseWord(tok)
		words = append(words, word)
		switch word {
		case ";":
			continuation = true
		}
	}
	return nil, p.sc.Err()
}

func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if !atEOF {
		if !bytes.ContainsRune(data, '\n') {
			return 0, nil, nil
		}
	} else if len(data) == 0 {
		return 0, nil, nil
	}
	idxs := tokenRE.FindSubmatchIndex(data)
	if len(idxs) != 4 {
		return 0, nil, fmt.Errorf("syntax error at %q", data)
	}
	advance = idxs[1]
	token = data[idxs[2]:idxs[3]]
	err = nil
	return
}

func parseWord(raw []byte) string {
	var buf strings.Builder
	if len(raw) >= 2 && raw[0] == '\'' {
		raw = raw[1 : len(raw)-1]
	}
	esc := false
	for _, b := range raw {
		switch b {
		case '\\':
			esc = true
		default:
			switch {
			case esc:
				switch b {
				case 'n':
					buf.WriteByte('\n')
				case 't':
					buf.WriteByte('\t')
				default:
					buf.WriteByte(b)
				}
				esc = false
			default:
				switch b {
				case '\\':
					esc = true
				default:
					buf.WriteByte(b)
				}
			}
		}
	}
	return buf.String()
}
