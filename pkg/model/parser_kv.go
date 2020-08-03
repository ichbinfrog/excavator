package model

import (
	"bufio"
	"regexp"
	"strings"
)

type pair struct {
	line     int
	value    []byte
	threats  int
	affected []byte
}

// KVParser is the base parser structure for a key value file
// Has two different regexes:
//  - Matcher: for <key><equal_operator><value>
//  - VarMatcher: for internal variable reference
//		for example, DB_PASSWORD='${DB_HOST}'
//
// For POSIX compliant .env files:
//		<key>: [a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}
// 		<equal>: '='
// 		<value>: \S*
type KVParser struct {
	Pairs      map[string]pair
	Bag        []string
	Threshold  float32
	Matcher    *regexp.Regexp
	VarMatcher *regexp.Regexp
}

// NewKVParser returns a new KV file leak parser
func NewKVParser(keyRegexp, equal, valRegexp, varRegexp string, bag []string) *KVParser {
	return &KVParser{
		Matcher:    regexp.MustCompile(keyRegexp + equal + valRegexp),
		VarMatcher: regexp.MustCompile(varRegexp),
		Bag:        bag,
		Pairs:      make(map[string]pair),
	}
}

// Parse reads file line by line to scan the KV file
func (k *KVParser) Parse(reader bufio.Reader, leakChan chan Leak, file string, rule *ParserRule) {
	lineNum := 0
	buf := bufio.NewScanner(&reader)
	for buf.Scan() {
		lineNum++
		line := buf.Bytes()
		match := k.Matcher.FindSubmatch(line)
		if len(match) <= 2 {
			continue
		}
		k.Pairs[string(match[1])] = pair{
			value:    match[2],
			line:     lineNum,
			affected: line,
		}
	}
	for key, pair := range k.Pairs {
		for _, keyword := range k.Bag {
			if strings.Contains(
				strings.ToLower(key),
				keyword,
			) {
				npair := &pair
				npair.threats = k.parseVariable(pair.value)
				if npair.threats == 0 {
					npair.threats = 1
				}
				k.Pairs[key] = *npair
				break
			}
		}
	}
	for _, pair := range k.Pairs {
		if pair.threats != 0 {
			leakChan <- FileLeak{
				File:     file,
				Line:     pair.line,
				Affected: string(pair.affected),
				Threat:   pair.threats,
				Parser:   rule,
			}
		}
	}
}

func (k *KVParser) parseVariable(value []byte) int {
	matches := k.VarMatcher.FindAllSubmatch(value, -1)
	innerCalls := 0
	for _, match := range matches {
		if len(match) >= 2 {
			for key := range k.Pairs {
				if strings.Compare(
					key,
					string(match[2]),
				) == 0 {
					innerCalls++
					break
				}
			}
		}
	}
	return innerCalls
}

// NewEnvParser returns a new environ file leak parser
func NewEnvParser(bag []string) *KVParser {
	return NewKVParser(
		`([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})`,
		`=`,
		`(.*)`,
		`\$(\{([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\}|([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}))`,
		bag,
	)
}

// NewDockerFileParser returns a new dockerfile leak parser
// this comes from the asumption that some ENV declarations
// can be left in the file during developpement cycles
func NewDockerFileParser(bag []string) *KVParser {
	return NewKVParser(
		`(ENV|ARGS)\s+([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})`,
		`=?`,
		`(.*)`,
		`\$(\{([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\}|([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}))`,
		bag,
	)
}
