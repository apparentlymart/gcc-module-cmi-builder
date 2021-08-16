package protocol

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var wordRE = regexp.MustCompile(
	`^[-+_/%.A-Za-z0-9]+$`,
)

var escReplacer = strings.NewReplacer(
	`\`, `\\`,
	`'`, `\'`,
	"\n", `\n`,
	"\t", `\t`,
)

type Writer struct {
	w *bufio.Writer
}

func NewWriter(w io.Writer) Writer {
	return Writer{bufio.NewWriter(w)}
}

func (w Writer) WriteBlock(block Block) error {
	for mi, msg := range block {
		for wi, word := range msg {
			if wi != 0 {
				if err := w.w.WriteByte(' '); err != nil {
					return err
				}
			}
			if !wordRE.MatchString(word) {
				word = "'" + escReplacer.Replace(word) + "'"
			}
			if _, err := w.w.WriteString(word); err != nil {
				return err
			}
		}
		if mi != len(block)-1 {
			if _, err := w.w.WriteString(" ;"); err != nil {
				return err
			}
		}
		if err := w.w.WriteByte('\n'); err != nil {
			return err
		}
	}
	return w.w.Flush()
}
