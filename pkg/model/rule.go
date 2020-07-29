package model

import (
	"encoding/hex"
	"hash/fnv"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

const (
	// contextsize is the amount of lines (before and after)
	// the violating line that will be added to the snippet
	contextSize = 2
)

type RuleSet struct {
	ApiVersion string `yaml:"apiVersion"`
	Checksum   string
	ReadAt     time.Time
	Rules      []Rule `yaml:"rules"`
}

type Rule struct {
	Definition  string  `yaml:"definition"`
	Description string  `yaml:"description,omitempty"`
	Category    string  `yaml:"category,omitempty"`
	Weight      float32 `yaml:"weight"`
	Compiled    *regexp.Regexp
}

func (r *RuleSet) ParseConfig(file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal().
			Str("path", file).
			Err(err).
			Msg("Failed to read rules definition @")
	}

	if err := yaml.Unmarshal(data, &r); err != nil {
		log.Fatal().
			Str("path", file).
			Err(err).
			Msg("Failed to unmarshal yaml @")
	}

	r.Checksum = hex.EncodeToString(fnv.New32().Sum(data))[:10]
	r.ReadAt = time.Now()

	for idx, rule := range r.Rules {
		r.Rules[idx].Compiled = regexp.MustCompile(rule.Definition)
	}
}

func (r *RuleSet) ParsePatch(patch *object.Patch, commit *object.Commit, repo *Repo, leakChan chan Leak) {
	for _, filePatch := range patch.FilePatches() {
		_, to := filePatch.Files()
		if to == nil {
			continue
		}
		for _, chunk := range filePatch.Chunks() {
			if chunk.Type() == diff.Add {
				lines := strings.Split(strings.Replace(chunk.Content(), "\r\n", "\n", -1), "\n")
				for idx, line := range lines {
					for _, rule := range r.Rules {
						match := rule.Compiled.FindStringSubmatchIndex(line)
						if match != nil {
							start := idx - contextSize
							end := idx + contextSize
							if start < 0 {
								start = 0
							}
							if end >= len(lines) {
								end = len(lines) - 1
							}
							disc := Leak{
								Line:     idx,
								Col:      match[0],
								Affected: idx - start,
								File:     to.Path(),
								Author:   commit.Author.Name,
								When:     commit.Author.When,
								Commit:   to.Hash().String(),
								Repo:     repo,
								Rule:     &rule,
							}
							disc.Snippet = make([]string, len(lines[start:end]))
							copy(disc.Snippet, lines[start:end])
							leakChan <- disc
							break
						}
					}
				}
			}
		}
	}
}
