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

// ParserRule is an union of a definition of the parser
// and it's instantiation. .Parser being the instance
// and (.Type, .Extensions) stores the definition
type ParserRule struct {
	Parser     ContextParser `yaml:"-"`
	Type       string        `yaml:"type"`
	Extensions []string      `yaml:"extensions"`
}

// Init creates a Parser if the .Type is defined
// TODO: Use reflect to make parsers more extensible
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
