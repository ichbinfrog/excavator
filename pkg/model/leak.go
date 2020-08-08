package model

import (
	"time"
)

// Leak is an interface that represents a possible credential leak
// discovered during the scan phase
type Leak interface {
}

// GitLeak is a potential leak detected associated with a commit
type GitLeak struct {
	// Hash of the commit (SHA-1)
	Commit string `yaml:"commit"`
	// File (path to the file from the root of the git repo)
	File string `yaml:"file"`
	// Line number within the affected file
	Line int `yaml:"line"`

	// Contains a couple of line before and after the commit
	// This allows for displaying the context of the discovery
	// without having to store / display the whole file
	Snippet []string `yaml:"snippet"`
	// Affected is the index of the offending line in the snippet slice
	StartIdx int
	EndIdx   int
	Affected int `yaml:"affected"`

	// TODO: Threat is an estimation of the product
	// 			 of the danger of the leak and the certainty of the analysis
	Threat float32 `yaml:"threat,omitempty"`
	// Stores the name of the author of the commit
	// This could be replaced with the email for better formatting
	Author string `yaml:"author,omitempty"`
	// Time of the commit
	When time.Time `yaml:"commit_date"`

	// Pointer to the offending rule
	IndepParserRule *IndepParserRule `yaml:"-"`
	// Pointer to the offending parser rule
	// The Rule and ParserRule attributes are exclusive
	CtxParserRule *CtxParserRule `yaml:"-"`
	Repo          *Repo          `yaml:"-"`
}

// FileLeak is a potential leak detected in the filesystem
type FileLeak struct {
	// File (path to the file from the root of the execution)
	File string `yaml:"file"`
	Line int    `yaml:"line"`

	// Single offending line
	// Or a concatenation (\n) of multiple lines
	Affected        string           `yaml:"affected"`
	Threat          float32          `yaml:"threat,omitempty"`
	Explain         string           `yaml:"explain"`
	IndepParserRule *IndepParserRule `yaml:"-"`
	CtxParserRule   *CtxParserRule   `yaml:"-"`
}
