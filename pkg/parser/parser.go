package parser

import (
	"bufio"

	"github.com/ichbinfrog/excavator/pkg/model"
)

var (
	bag = []string{
		"password",
		"host",
		"proxy",
	}
)

// ContextParser is an interface to parsers
// that only a full file context for analysis
// in order to search for a potential leak
type ContextParser interface {
	Parse(buf *bufio.Scanner)
}

func Parse(extension string, buf *bufio.Scanner, leakChan chan model.FileLeak) bool {
	switch extension {
	case ".env":
	case "Dockerfile":
	default:
		return false
	}
	return true
}
