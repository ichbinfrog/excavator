package model

import "regexp"

// IndepParserRule represents a context independant parser
type IndepParserRule struct {
	Definition  string  `yaml:"definition"`
	Description string  `yaml:"description,omitempty"`
	Category    string  `yaml:"category,omitempty"`
	Weight      float32 `yaml:"weight"`
	Compiled    *regexp.Regexp
}
