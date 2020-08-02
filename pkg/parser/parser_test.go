package parser

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog/log"
)

func TestKVParser(t *testing.T) {
	type Case struct {
		Parser   *KVParser
		TestFile []string
		Name     string
	}

	testCases := []Case{
		{
			Parser: NewEnvParser([]string{
				"password",
				"host",
				"proxy",
			}),
			TestFile: []string{"tests/laravel.env"},
			Name:     "env",
		},
		{
			Parser: NewDockerFileParser([]string{
				"password",
				"host",
				"proxy",
			}),
			TestFile: []string{"tests/Dockerfile"},
			Name:     "dockerfile",
		},
	}

	for _, testCase := range testCases {
		for _, file := range testCase.TestFile {
			fd, err := os.Open(file)
			if err != nil {
				log.Fatal().Str("path", file).Msg("Unable to open test case")
			}
			t.Run(fmt.Sprintf("kv_parser_"+testCase.Name+"_"+file), func(t *testing.T) {
				testCase.Parser.Parse(bufio.NewScanner(bufio.NewReader(fd)))
				fmt.Printf("%+v\n", testCase.Parser)
			})
			fd.Close()
		}
	}
}