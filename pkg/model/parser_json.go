package model

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
)

type jsonParser struct {
	keyBag     *[]string
	keyMatcher *regexp.Regexp
}

func (j *jsonParser) Parse(reader io.Reader, leakChan chan Leak, file string, rule *CtxParserRule) {
	lineNum := 0
	buf := bufio.NewScanner(reader)
	affected := &bytes.Buffer{}
	affectedLine := 0
	for buf.Scan() {
		lineNum++
		line := buf.Bytes()
		// Add everything until the next key match
		matches := j.keyMatcher.FindAllIndex(line, -1)
		switch len(matches) {
		// no matches
		case 0:
			if affected.Len() > 0 {
				// if there's a running value matcher
				affected.Write(line)
			}
			break
		// single match
		case 1:
			if affected.Len() > 0 {
				// if there's a running value matcher
				leakChan <- FileLeak{
					File:          file,
					Line:          affectedLine,
					Affected:      affected.String(),
					CtxParserRule: rule,
				}
				affected.Reset()
				affectedLine = lineNum
			}
			for _, key := range *j.keyBag {
				if bytes.Contains(
					bytes.ToLower(line[matches[0][0]:matches[0][1]]),
					[]byte(key),
				) {
					// add discovery if there's no running value matcher
					if bytes.ContainsRune(line[matches[0][1]:], '{') {
						// ignores beginning of objects
						// because they don't usually contain potential leaks
						// that aren't present in keys that are deeper in the structure
						// This also fixes duplicate discoveries
						break
					}
					affectedLine = lineNum
					affected.Write(line)
					break
				}
			}
			break
		default:
			// more than one match in the same line
			for _, key := range *j.keyBag {
				for i := 0; i < len(matches); i++ {
					if bytes.Contains(
						bytes.ToLower(line[matches[0][0]:matches[0][1]]),
						[]byte(key),
					) {
						// add discovery if there's no running value matcher
						leakChan <- FileLeak{
							File:          file,
							Line:          lineNum,
							Affected:      string(line[matches[i][0]:matches[i+1][0]]),
							CtxParserRule: rule,
						}
						if i == len(matches)-1 {
							affected.Write(line)
							affectedLine = lineNum
						}
					}
				}
			}
		}
	}
}

func newJSONParser(keyBag *[]string) *jsonParser {
	return &jsonParser{
		keyBag:     keyBag,
		keyMatcher: regexp.MustCompile(`"([^\t\n]*)"\s*:`),
	}
}
