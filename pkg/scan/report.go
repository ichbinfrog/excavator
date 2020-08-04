package scan

import (
	"html/template"
	"os"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/gobuffalo/packr/v2"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// ReportInterface modules writer behaviour for different reports
type ReportInterface interface {
	Write(Scanner)
}

// HTMLReport implements the ReportInterface to write html reports
type HTMLReport struct {
	Outfile  string
	Template *template.Template
}

// YamlReport implements the ReportInterface to write yaml reports
type YamlReport struct {
	Outfile string
}

func createFile(path string) *os.File {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal().
			Str("output", path).
			Err(err).
			Msg("Unable to create file")
	}
	return f
}

func (h HTMLReport) Write(s Scanner) {
	h.Outfile = "index.html"
	f := createFile(h.Outfile)
	defer f.Close()

	box := packr.New(".", "./static")
	report, err := box.FindString("report.gohtml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load static box")
	}

	h.Template = template.Must(template.New("report.gohtml").Funcs(
		sprig.FuncMap(),
	).Parse(report))
	if err := h.Template.Execute(f, s); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to execute template")
	}

	log.Info().
		Str("path", h.Outfile).
		Msg("Output has been written to")
}

func (y YamlReport) Write(s Scanner) {
	y.Outfile = time.Now().Format(time.RFC3339) + ".yaml"
	f := createFile(y.Outfile)
	defer f.Close()

	data, err := yaml.Marshal(&s)
	if err != nil || data == nil {
		log.Fatal().
			Err(err).
			Msg("Unable to marshal structure to yaml")
	}

	if _, err := f.Write(data); err != nil {
		log.Fatal().
			Str("output", y.Outfile).
			Err(err).
			Msg("Unable to write to file")
	}
	log.Info().
		Str("path", y.Outfile).
		Msg("Output has been written to")
}
