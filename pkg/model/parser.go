package model

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

var (
	bag = []string{
		"pass",
		"host",
		"proxy",
	}
)

// ContextParser is an interface to parsers
// that only a full file context for analysis
// in order to search for a potential leak
type ContextParser interface {
	Parse(buf *bufio.Scanner, leakChan chan Leak, file string, rule *ParserRule)
}

type ParserRule struct {
	Parser     ContextParser `yaml:"-"`
	Type       string        `yaml:"type"`
	Extensions []string      `yaml:"extensions"`
}

func (p *ParserRule) Init() {
	switch p.Type {
	case "env":
		p.Parser = NewEnvParser(bag)
		break
	case "dockerfile":
		p.Parser = NewEnvParser(bag)
		break
	default:
		log.Fatal().
			Str("parser_type", p.Type).
			Msg("Unknown parser type, must be (env, dockerfile)")
	}
}
