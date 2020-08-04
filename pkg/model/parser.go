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
	Parse(reader bufio.Reader, leakChan chan Leak, file string, rule *CtxParserRule)
}

// CtxParserRule is an union of a definition of the parser
// and it's instantiation. .Parser being the instance
// and (.Type, .Extensions) stores the definition
type CtxParserRule struct {
	Parser     ContextParser `yaml:"-"`
	Type       string        `yaml:"type"`
	Extensions []string      `yaml:"extensions"`
}

// Init creates a Parser if the .Type is defined
// TODO: Use reflect to make parsers more extensible
// TODO: ParserPool instead of instantiating new one every call
//
func (c *CtxParserRule) Init() {
	switch c.Type {
	case "env":
		c.Parser = NewEnvParser(bag)
		break
	case "dockerfile":
		c.Parser = NewDockerFileParser(bag)
		break
	case "properties":
		c.Parser = NewPropertiesParser(bag)
		break
	default:
		log.Fatal().
			Str("parser_type", c.Type).
			Msg("Unknown parser type, must be (env, dockerfile)")
	}
}
