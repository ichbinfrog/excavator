package model

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
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
	var data interface{}

	if err := yaml.Unmarshal(buf.Bytes(), &data); err != nil {
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

func newYAMLParser(keyBag *[]string) *yamlParser {
	return &yamlParser{
		keyBag: keyBag,
	}
}
