package scan

import "testing"

func TestLocalClient(t *testing.T) {
	s := &Scanner{}
	s.New("https://github.com/eclipse/steady", "test", "../../resources/rules.yaml", &HTMLReport{}, false)
	s.Scan(100)
}

func TestHTMLReport(t *testing.T) {
	s := &Scanner{}
	s.New("https://github.com/eclipse/steady", "test", "../../resources/rules.yaml", &HTMLReport{}, false)
	s.Scan(100)
}
