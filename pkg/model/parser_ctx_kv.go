package model

import (
	"bufio"
	"bytes"
	"io"
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
//  - matcher: for <key><equal_operator><value>
//  - varMatcher: for internal variable reference
//		for example, DB_PASSWORD='${DB_HOST}'
//
// For POSIX compliant .env files:
//		<key>: [a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}
// 		<equal>: '='
// 		<value>: \S*
type kvParser struct {
	keyBag     *[]string
	threshold  float32
	matcher    *regexp.Regexp
	varMatcher *regexp.Regexp
}

// newKVParser returns a new KV file leak parser
func newKVParser(keyRegexp, equal, valRegexp, varRegexp string, bag *[]string) *kvParser {
	return &kvParser{
		matcher:    regexp.MustCompile(keyRegexp + equal + valRegexp),
		varMatcher: regexp.MustCompile(varRegexp),
		keyBag:     bag,
	}
}

// Parse reads file line by line to scan the KV file
func (k *kvParser) Parse(reader io.Reader, leakChan chan Leak, file string, rule *CtxParserRule) {
	pairs := make(map[string]pair)
	lineNum := 0
	buf := bufio.NewScanner(reader)
	for buf.Scan() {
		lineNum++
		line := buf.Bytes()
		match := k.matcher.FindSubmatch(line)
		if len(match) <= 2 {
			continue
		}
		value := bytes.TrimSpace(match[2])
		if len(value) == 0 {
			continue
		}
		pairs[string(match[1])] = pair{
			value:    match[2],
			line:     lineNum,
			affected: line,
		}
	}
	for key, pair := range pairs {
		for _, keyword := range *k.keyBag {
			if strings.Contains(
				strings.ToLower(key),
				keyword,
			) {
				npair := &pair
				matches := k.varMatcher.FindAllSubmatch(pair.value, -1)
				innerCalls := 0
				for _, match := range matches {
					if len(match) >= 2 {
						for key := range pairs {
							if strings.Compare(key, string(match[2])) == 0 {
								innerCalls++
								break
							}
						}
					}
				}
				npair.threats = innerCalls
				if npair.threats == 0 {
					npair.threats = 1
				}
				pairs[key] = *npair
				break
			}
		}
	}
	for _, pair := range pairs {
		if pair.threats != 0 {
			leakChan <- FileLeak{
				File:          file,
				Line:          pair.line,
				Affected:      string(pair.affected),
				Threat:        pair.threats,
				CtxParserRule: rule,
			}
		}
	}
}

// NewEnvParser returns a new environ file leak parser
func newEnvParser(keyBag *[]string) *kvParser {
	return newKVParser(
		`([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})`,
		`=`,
		`(.*)`,
		`\$(\{([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\}|([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}))`,
		keyBag,
	)
}

// NewDockerFileParser returns a new dockerfile leak parser
// this comes from the asumption that some ENV declarations
// can be left in the file during developpement cycles
func newDockerFileParser(keyBag *[]string) *kvParser {
	return newKVParser(
		`(ENV|ARGS)\s+([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})`,
		`=?`,
		`(.*)`,
		`\$(\{([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\}|([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}))`,
		keyBag,
	)
}

// NewPropertiesParser returns a new .properties leak parser
func newPropertiesParser(keyBag *[]string) *kvParser {
	return newKVParser(
		`([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\s*`,
		`=`,
		`(.*)`,
		`\$(\{([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\}|([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}))`,
		keyBag,
	)
}

// NewShParser returns a new .properties leak parser
func newShParser(keyBag *[]string) *kvParser {
	return newKVParser(
		`\s*(export)?([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\s*`,
		`=`,
		`(.*)`,
		`\$(\{([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\}|([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}))`,
		keyBag,
	)
}
