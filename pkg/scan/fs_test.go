package scan

import (
	"testing"
)

func TestFsClient(t *testing.T) {
	f := FsScanner{}
	f.New(".", "../../resources/rules.yaml", &HTMLReport{}, false)
	f.Scan(5)
}
