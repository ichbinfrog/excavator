package scan

import (
	"fmt"
	"testing"
)

func TestFsClient(t *testing.T) {
	f := NewFsScanner(".", "../../resources/rules.yaml", &HTMLReport{}, false)
	f.Scan(5)
}
func BenchmarkFSScan(b *testing.B) {
	conccurrent := []int{50}
	for _, i := range conccurrent {
		b.Run(fmt.Sprintf("fsscan_%d", i), func(b *testing.B) {
			b.StartTimer()
			g := NewFsScanner("../..", "../../resources/rules.yaml", &HTMLReport{}, false)
			g.Scan(i)
			b.ResetTimer()
		})
	}
}
