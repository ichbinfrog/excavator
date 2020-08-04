package scan

import (
	"testing"
)

func TestFsClient(t *testing.T) {
	f := NewFsScanner(".", "../../resources/rules.yaml", &HTMLReport{}, false)
	f.Scan(5)
}
