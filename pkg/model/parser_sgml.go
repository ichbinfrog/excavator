package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type sgmlParser struct {
	keyBag     *[]string
	keyMatcher *regexp.Regexp
	Unmarshal  func([]byte, interface{}) error
}

func newSGMLParser(keyBag *[]string, unmarshaller func([]byte, interface{}) error) *sgmlParser {
	return &sgmlParser{
		keyBag:     keyBag,
		keyMatcher: regexp.MustCompile(`"([^\t\n]*)"\s*:`),
		Unmarshal:  unmarshaller,
	}
}

func (s *sgmlParser) Parse(reader io.Reader, leakChan chan Leak, file string, rule *CtxParserRule) {
	buf := &bytes.Buffer{}
	buf.ReadFrom(reader)
	var data interface{}

	if err := s.Unmarshal(buf.Bytes(), &data); err != nil {
		log.Trace().
			Err(err).
			Str("file", file).
			Msg("Failed to unmarshal")
		return
	}

	flattened := map[string]string{}
	switch root := data.(type) {
	// !!map document node type
	case map[string]interface{}:
		for k, v := range root {
			flattenMap(k, v, flattened)
		}
		break
	// !!seq document node type
	case []interface{}:
		for k, v := range root {
			flattenMap(strconv.Itoa(k), v, flattened)
		}
		break
	}

	for k, v := range flattened {
		for _, key := range *s.keyBag {
			last := k
			lastIndex := strings.LastIndex(k, ".")
			if lastIndex != -1 {
				last = k[lastIndex:]
			}
			if strings.Contains(
				strings.ToLower(last),
				key,
			) {
				leakChan <- FileLeak{
					File:          file,
					Line:          0,
					Affected:      k + ":" + v,
					StartIdx:      len(k) + 1,
					EndIdx:        0,
					CtxParserRule: rule,
				}
			}
		}
	}
}

func flattenMap(prefix string, value interface{}, res map[string]string) {
	switch submap := value.(type) {
	case map[interface{}]interface{}:
		for k, v := range submap {
			flattenMap(fmt.Sprintf("%s.%v", prefix, k), v, res)
		}
		return
	case []interface{}:
		for i, v := range submap {
			flattenMap(fmt.Sprintf("%s[%d]", prefix, i), v, res)
		}
		return
	case map[string]interface{}:
		for k, v := range submap {
			flattenMap(prefix+"."+k, v, res)
		}
		return
	default:
		res[prefix] = fmt.Sprintf("%v", value)
	}
}
func newJSONParser(keyBag *[]string) *sgmlParser {
	return newSGMLParser(keyBag, json.Unmarshal)
}

func newYAMLParser(keyBag *[]string) *sgmlParser {
	return newSGMLParser(keyBag, yaml.Unmarshal)
}
