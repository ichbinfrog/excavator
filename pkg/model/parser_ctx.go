package model

import (
	"bufio"

	"github.com/rs/zerolog/log"
)

// ContextParser is an interface to parsers
// that only a full file context for analysis
// in order to search for a potential leak
type ContextParser interface {
	Parse(reader *bufio.Reader, leakChan chan Leak, file string, rule *CtxParserRule)
}

// CtxParserRule is an union of a definition of the parser
// and it's instantiation. .Parser being the instance
// and (.Type, .Extensions) stores the definition
type CtxParserRule struct {
	Parser     ContextParser `yaml:"-"`
	Type       string        `yaml:"type"`
	Extensions []string      `yaml:"extensions"`
	KeyBag     []string      `yaml:"keys"`
}

// Init creates a Parser if the .Type is defined
// TODO: Use reflect to make parsers more extensible
//
func (c *CtxParserRule) Init() {
	if c.KeyBag == nil {
		c.KeyBag = []string{
			"pass",
			"host",
			"proxy",
			"key",
		}
	}

	switch c.Type {
	case "env":
		c.Parser = NewEnvParser(&c.KeyBag)
		break
	case "dockerfile":
		c.Parser = NewDockerFileParser(&c.KeyBag)
		break
	case "properties":
		c.Parser = NewPropertiesParser(&c.KeyBag)
		break
	case "shell":
		c.Parser = NewShParser(&c.KeyBag)
		break
	default:
		log.Fatal().
			Str("parser_type", c.Type).
			Msg("Unknown parser type, must be (env, dockerfile)")
	}
}
