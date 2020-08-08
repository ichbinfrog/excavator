package model

import (
	"io"

	"github.com/rs/zerolog/log"
)

// contextParser is an interface to parsers
// that only a full file context for analysis
// in order to search for a potential leak
type contextParser interface {
	Parse(reader io.Reader, leakChan chan Leak, file string, rule *CtxParserRule)
}

// CtxParserRule is an union of a definition of the parser
// and it's instantiation.
type CtxParserRule struct {
	// Instance of the parser
	Parser contextParser `yaml:"-"`
	// Name of the parser type
	Type string `yaml:"type"`
	// Extensions which the parser takes into consideration
	Extensions []string `yaml:"extensions"`
	// Bag of words used mainly to identify keys/values
	// that are potential leaks
	KeyBag []string `yaml:"keys"`
	// Confidence of the assessment
	// - "High" : for context based parsers
	// - "Low" : for regexp/context insensitive parsers
	Confidence string `yaml:"confidence"`
}

// Init creates a Parser if the .Type is defined
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
		c.Parser = newEnvParser(&c.KeyBag)
		break
	case "dockerfile":
		c.Parser = newDockerFileParser(&c.KeyBag)
		break
	case "properties":
		c.Parser = newPropertiesParser(&c.KeyBag)
		break
	case "shell":
		c.Parser = newShParser(&c.KeyBag)
		break
	case "json":
		c.Parser = newJSONParser(&c.KeyBag)
		break
	case "yaml":
		c.Parser = newYAMLParser(&c.KeyBag)
		break
	case "xml":
		c.Parser = newXMLParser(&c.KeyBag)
		break
	default:
		log.Fatal().
			Str("parser_type", c.Type).
			Msg("Unknown parser type, must be (env, dockerfile)")
	}
}
