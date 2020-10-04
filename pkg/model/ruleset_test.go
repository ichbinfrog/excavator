package model

import (
	"fmt"
	"testing"
)

func TestRuleMarshal(t *testing.T) {
	rs := RuleSet{}
	rs.ParseConfig("../../resources/rules.yaml")
	fmt.Printf("%+v\n", rs)
}
