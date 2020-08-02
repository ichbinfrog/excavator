package model

import (
	"time"
)

type Leak interface {
}

type GitLeak struct {
	Commit   string    `yaml:"commit"`
	File     string    `yaml:"file"`
	Line     int       `yaml:"line"`
	Affected int       `yaml:"affected"`
	Snippet  []string  `yaml:"snippet"`
	Threat   float32   `yaml:"threat,omitempty"`
	Author   string    `yaml:"author,omitempty"`
	When     time.Time `yaml:"commit_date"`
	Rule     *Rule     `yaml:"-"`
	Repo     *Repo     `yaml:"-"`
}

type FileLeak struct {
	File     string  `yaml:"file"`
	Line     int     `yaml:"line"`
	Affected string  `yaml:"affected"`
	Threat   float32 `yaml:"threat,omitempty"`
	Rule     *Rule   `yaml:"-"`
}
