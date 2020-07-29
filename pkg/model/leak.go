package model

import "time"

type Leak struct {
	Commit    string    `yaml:"commit"`
	File      string    `yaml:"file"`
	Line      int       `yaml:"line"`
	Col       int       `yaml:"col"`
	Affected  int       `yaml:"affected"`
	Snippet   []string  `yaml:"snippet"`
	Certainty float32   `yaml:"certainty,omitempty"`
	Author    string    `yaml:"author,omitempty"`
	When      time.Time `yaml:"commit_date"`
	Rule      *Rule     `yaml:"-"`
	Repo      *Repo     `yaml:"-"`
}
