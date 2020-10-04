package model

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestRuleMarshal(t *testing.T) {
	rs := RuleSet{}
	rs.ParseConfig("../../resources/rules.yaml")
	fmt.Printf("%+v\n", rs)
}

func TestArchive(t *testing.T) {
	fd, err := os.Open("tests/test_archive.tar")
	if err != nil {
		fmt.Println(err)
		return
	}
	tarReader := tar.NewReader(fd)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		fmt.Println(header)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
