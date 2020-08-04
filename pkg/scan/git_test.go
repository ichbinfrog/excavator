package scan

import (
	"fmt"
	"testing"
)

func TestLocalClient(t *testing.T) {
	g := NewGitScanner("https://github.com/eclipse/steady", "test", "../../resources/rules.yaml", &HTMLReport{}, false)
	g.Scan(100)
}

func TestHTMLReport(t *testing.T) {
	g := NewGitScanner("https://github.com/eclipse/steady", "test", "../../resources/rules.yaml", &HTMLReport{}, false)
	g.Scan(100)
}

func BenchmarkScan(b *testing.B) {
	conccurrent := []int{50}
	for _, i := range conccurrent {
		b.Run(fmt.Sprintf("scan_%d", i), func(b *testing.B) {
			b.StartTimer()
			g := NewGitScanner("https://github.com/eclipse/steady", "test", "../../resources/rules.yaml", &HTMLReport{}, false)
			g.Scan(i)
			b.ResetTimer()
		})
	}
}
