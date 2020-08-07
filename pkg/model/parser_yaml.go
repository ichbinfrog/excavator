package model

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type yamlParser struct {
	keyBag *[]string
}

func (y *yamlParser) Parse(reader io.Reader, leakChan chan Leak, file string, rule *CtxParserRule) {
	buf := &bytes.Buffer{}
	buf.ReadFrom(reader)
	var root map[string]interface{}

	if err := yaml.Unmarshal(buf.Bytes(), &root); err != nil {
		log.Error().
			Err(err).
			Str("file", file).
			Msg("Failed to unmarshal")
	}

	// flatten yaml to a single map to search for potential leaks
	// this method does not allow for identifying line numbers
	flattened := map[string]string{}
	for k, v := range root {
		flatten(k, v, flattened)
	}
	for k, v := range flattened {
		for _, key := range *y.keyBag {
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
					CtxParserRule: rule,
				}
			}
		}
	}
}

func flatten(prefix string, value interface{}, res map[string]string) {
	switch submap := value.(type) {
	case map[interface{}]interface{}:
		for k, v := range submap {
			flatten(fmt.Sprintf("%s.%v", prefix, k), v, res)
		}
		return
	case []interface{}:
		for i, v := range submap {
			flatten(fmt.Sprintf("%s[%d]", prefix, i), v, res)
		}
		return
	case map[string]interface{}:
		for k, v := range submap {
			flatten(prefix+"."+k, v, res)
		}
		return
	default:
		res[prefix] = fmt.Sprintf("%v", value)
	}
}

func newYAMLParser(keyBag *[]string) *yamlParser {
	return &yamlParser{
		keyBag: keyBag,
	}
}
