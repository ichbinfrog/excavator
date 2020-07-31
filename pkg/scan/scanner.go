package scan

import (
	"github.com/ichbinfrog/excavator/pkg/model"
	"github.com/schollz/progressbar/v3"
)

type Scanner struct {
	RulesPath string

	Debug       bool
	ProgressBar *progressbar.ProgressBar
	RuleSet     *model.RuleSet
	Output      ReportInterface
}
