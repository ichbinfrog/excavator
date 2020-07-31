package model

import "time"

type GitLeak struct {
	Commit    string    `yaml:"commit"`
	File      string    `yaml:"file"`
	Line      int       `yaml:"line"`
	Affected  int       `yaml:"affected"`
	Snippet   []string  `yaml:"snippet"`
	Certainty float32   `yaml:"certainty,omitempty"`
	Author    string    `yaml:"author,omitempty"`
	When      time.Time `yaml:"commit_date"`
	Rule      *Rule     `yaml:"-"`
	Repo      *Repo     `yaml:"-"`
}

type FileLeak struct {
	File      string  `yaml:"file"`
	Line      int     `yaml:"line"`
	Size      int64   `yaml:"size"`
	Affected  string  `yaml:"affected"`
	Certainty float32 `yaml:"certainty,omitempty"`
	Rule      *Rule   `yaml:"-"`
}
