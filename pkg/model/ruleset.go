package model

import (
	"bytes"
	"encoding/hex"
	"hash/fnv"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gobuffalo/packr/v2"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	// contextsize is the amount of lines (before and after)
	// the violating line that will be added to the snippet
	contextSize = 4
)

// RuleSet groups all Rules and Parsers interpreted from the user defined file
//
// - Rules represent parsers that are context independant
//   it can parse a file line by line to precisely find the leak
// - Parsers are parsers that need the entire file as a context
//   to analyse for leaks correctly (TODO: rename)
//
type RuleSet struct {
	// Version of the configuration file
	// Not used currently but for future proofing
	APIVersion string `yaml:"apiVersion"`

	// FNV hash of the configuration file
	// Useful for determining whether or not the definition file
	// has been changed. (for future uses)
	Checksum          string
	ReadAt            time.Time
	IndepParsers      []IndepParserRule `yaml:"rules"`
	CtxParsers        []CtxParserRule   `yaml:"parsers"`
	BlackList         []string          `yaml:"black_list"`
	BlackListCompiled []*regexp.Regexp  `yaml:"-"`
}

// ParseConfig reads the user defined configuration file
func (r *RuleSet) ParseConfig(file string) {
	var data []byte
	var err error

	if len(file) == 0 {
		box := packr.New("rules", "../../resources")
		data, err = box.Find("rules.yaml")
		if err != nil {
			log.Fatal().
				Str("path", file).
				Err(err).
				Msg("Failed to read static binary definition")
		}
	} else {
		data, err = ioutil.ReadFile(file)
		if err != nil {
			log.Fatal().
				Str("path", file).
				Err(err).
				Msg("Failed to read rules definition @")
		}
	}
	if err := yaml.Unmarshal(data, &r); err != nil {
		log.Fatal().
			Str("path", file).
			Err(err).
			Msg("Failed to unmarshal yaml @")
	}
	r.Checksum = hex.EncodeToString(fnv.New32().Sum(data))[:10]
	r.ReadAt = time.Now()

	for idx, rule := range r.IndepParsers {
		r.IndepParsers[idx].Compiled = regexp.MustCompile(rule.Definition)
	}
	for idx := range r.CtxParsers {
		r.CtxParsers[idx].Init()
	}

	for _, bl := range r.BlackList {
		r.BlackListCompiled = append(r.BlackListCompiled, regexp.MustCompile(bl))
	}
}

// ParsePatch iterates over each chunk of the patch object
// and applies all context indenpendant rules
// TODO: allow context dependant rules
//
func (r *RuleSet) ParsePatch(patch *object.Patch, commit *object.Commit, repo *Repo, leakChan chan Leak) {
	for _, filePatch := range patch.FilePatches() {
		if filePatch.IsBinary() {
			break
		}
		_, to := filePatch.Files()
		if to == nil {
			continue
		}
		for _, blacklist := range r.BlackListCompiled {
			if blacklist.MatchString(to.Path()) {
				break
			}
		}

		for _, chunk := range filePatch.Chunks() {
			if chunk.Type() == diff.Add {
				lines := strings.Split(chunk.Content(), "\n")
				for idx, line := range lines {
					for _, rule := range r.IndepParsers {
						match := rule.Compiled.FindStringIndex(line)
						if len(match) > 0 {
							start := idx - contextSize
							end := idx + contextSize
							if start < 0 {
								start = 0
							}
							if end >= len(lines) {
								end = len(lines) - 1
							}
							disc := GitLeak{
								Line:            idx,
								Affected:        idx - start,
								File:            to.Path(),
								StartIdx:        match[0],
								EndIdx:          match[1],
								Author:          commit.Author.Name,
								When:            commit.Author.When,
								Commit:          to.Hash().String(),
								Repo:            repo,
								IndepParserRule: &rule,
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

// ParseFile reads a given file and applies all rules given
func (r *RuleSet) ParseFile(file string, leakChan chan Leak) {
	fd, err := os.Open(file)
	if err != nil {
		log.Trace().
			Str("file", file).
			Err(err).
			Msg("Failed to read")
		return
	}
	defer fd.Close()

	for _, rule := range r.CtxParsers {
		for _, ext := range rule.Extensions {
			if strings.HasSuffix(file, ext) {
				rule.Parser.Parse(fd, leakChan, file, &rule)
				return
			}
		}
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(fd)
	lines := strings.Split(buf.String(), "\n")
	for idx, line := range lines {
		if !utf8.ValidString(line) {
			continue
		}
		for _, rule := range r.IndepParsers {
			match := rule.Compiled.FindStringIndex(line)

			if len(match) > 0 {
				start := idx - contextSize
				end := idx + contextSize
				if start < 0 {
					start = 0
				}
				if end >= len(lines) {
					end = len(lines) - 1
				}
				disc := FileLeak{
					File:            file,
					StartIdx:        match[0],
					EndIdx:          match[1],
					Line:            idx,
					Affected:        idx - start,
					IndepParserRule: &rule,
				}
				disc.Snippet = make([]string, len(lines[start:end]))
				copy(disc.Snippet, lines[start:end])
				leakChan <- disc
				break
			}
		}
	}
}
