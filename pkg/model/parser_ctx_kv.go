package model

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

type pair struct {
	line     int
	indexes  [4]int
	threat   float32
	affected string
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
		line := buf.Text()
		match := k.matcher.FindStringSubmatchIndex(line)
		// FindStringSubmatchIndex should return an []int with
		//	 0     1            2               3              4               5
		// [ 0 {len(line)} {start idx key} {end idx key} {start idx val} {end idx val} ]
		if len(match) < 6 || match[2] == -1 || match[4] == -1 {
			continue
		}
		value := line[match[4]:match[5]]
		if len(value) == 0 {
			continue
		}
		pairs[line[match[2]:match[3]]] = pair{
			line:     lineNum,
			affected: line,
			indexes: [4]int{
				match[2],
				match[3],
				match[4],
				match[5],
			},
		}
	}
	for key, pair := range pairs {
		for _, keyword := range *k.keyBag {
			if strings.Contains(
				strings.ToLower(key),
				keyword,
			) {
				npair := &pair
				matches := k.varMatcher.FindAllStringSubmatch(pair.affected[pair.indexes[2]:pair.indexes[3]], -1)
				innerCall := false
				for _, match := range matches {
					if len(match) >= 2 || innerCall {
						for key := range pairs {
							if strings.Compare(key, match[2]) == 0 {
								innerCall = true
								break
							}
						}
					}
				}
				if innerCall {
					// Higher threat if the potential leak is hardcoded
					npair.threat = 0.7
				} else {
					npair.threat = 1
				}
				pairs[key] = *npair
				break
			}
		}
	}
	for _, pair := range pairs {
		if pair.threat != 0.0 {
			leakChan <- FileLeak{
				File:          file,
				Line:          pair.line,
				Snippet:       []string{pair.affected},
				Affected:      0,
				StartIdx:      pair.indexes[2],
				EndIdx:        pair.indexes[3],
				Threat:        pair.threat,
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
		`\s*(.*)\s*`,
		`\$(\{([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\}|([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}))`,
		keyBag,
	)
}

// NewDockerFileParser returns a new dockerfile leak parser
// this comes from the asumption that some ENV declarations
// can be left in the file during developpement cycles
func newDockerFileParser(keyBag *[]string) *kvParser {
	return newKVParser(
		`(?:ENV|ARGS)\s+([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})`,
		`=?`,
		`\s*(.*)\s*`,
		`\$(\{([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\}|([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}))`,
		keyBag,
	)
}

// NewPropertiesParser returns a new .properties leak parser
func newPropertiesParser(keyBag *[]string) *kvParser {
	return newKVParser(
		`([a-zA-Z_]{1,}[a-zA-Z0-9_]{0,})\s*`,
		`=`,
		`\s*(.*)\s*`,
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
