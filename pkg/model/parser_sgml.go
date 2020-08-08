package model

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// sgmlParser is the base parser for Standard Generalized Markup Language (SGML)
// derived languages (json, yaml, xml, ...).
//
// The SGML Parser attempts to flatten the data structure into a key value map
// according to a certain set of rules
//  - the key is built by appending the name of each of the parent node
//  - for arrays the name of the parent node is it's index in the file
//    NOTE that arrays are not ordering in this standard, thus the index
//    is not consistent and only serves to build an unique key
//
// In the end, we end up with a map as such
// "spec.template.spec.containers[0].name" :  "nginx"
//
// TODO: Implement the yaml node visitor to store the line number
//			 as well as to run the key building iteratively for performance
//
//
type sgmlParser struct {
	keyBag     *[]string
	keyMatcher *regexp.Regexp
	// function that unmarshals the data read from the file into
	// either a map[string]interface{} or a []interface{} as an
	// internal representation
	unmarshal func([]byte, interface{}) error
}

func newSGMLParser(keyBag *[]string, unmarshaller func([]byte, interface{}) error) *sgmlParser {
	return &sgmlParser{
		keyBag:     keyBag,
		keyMatcher: regexp.MustCompile(`"([^\t\n]*)"\s*:`),
		unmarshal:  unmarshaller,
	}
}

func (s *sgmlParser) Parse(reader io.Reader, leakChan chan Leak, file string, rule *CtxParserRule) {
	buf := &bytes.Buffer{}
	buf.ReadFrom(reader)
	var data interface{}

	if err := s.unmarshal(buf.Bytes(), &data); err != nil {
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
					Snippet:       []string{k + ":" + v},
					Affected:      0,
					StartIdx:      len(k) + 1,
					EndIdx:        0,
					CtxParserRule: rule,
					Confidence:    "High",
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

func newXMLParser(keyBag *[]string) *sgmlParser {
	return newSGMLParser(keyBag, xml.Unmarshal)
}
