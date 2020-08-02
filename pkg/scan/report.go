package scan

import (
	"html/template"
	"os"
	"reflect"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/gobuffalo/packr"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type ReportInterface interface {
	Write(interface{})
}

type HTMLReport struct {
	Outfile  string
	Template *template.Template
}
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

func (h *HTMLReport) Write(g interface{}) {
	h.Outfile = "index.html"
	f := createFile(h.Outfile)
	defer f.Close()

	switch v := g.(type) {
	case *GitScanner:
	case *FsScanner:
		box := packr.NewBox(".")
		report, err := box.FindString("report.gohtml")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load static box")
		}

		h.Template = template.Must(template.New("report.gohtml").Funcs(
			sprig.FuncMap(),
		).Parse(report))
		if err := h.Template.Execute(f, v); err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to execute template")
		}
		// box := packr.NewBox(".")
		// report, err := box.FindString("report_" + v.Name + ".gohtml")
		// if err != nil {
		// 	log.Fatal().Err(err).Msg("Failed to load static box")
		// }

		// h.Template = template.Must(template.New("report_" + v.Name + ".gohtml").Funcs(
		// 	sprig.FuncMap(),
		// ).Parse(report))
		// if err := h.Template.Execute(f, v); err != nil {
		// 	log.Fatal().
		// 		Err(err).
		// 		Msg("Failed to execute template")
		// }
		break
	default:
		log.Fatal().
			Str("type", reflect.TypeOf(v).String()).
			Msg("Failed to write report, unknown type")
	}
	log.Info().
		Str("path", h.Outfile).
		Msg("Output has been written to")
}

func (y *YamlReport) Write(g interface{}) {
	y.Outfile = time.Now().Format(time.RFC3339) + ".yaml"
	f := createFile(y.Outfile)
	defer f.Close()

	var data []byte
	switch v := g.(type) {
	case *GitScanner:
	case *FsScanner:
		data, err := yaml.Marshal(&v)
		if err != nil || data == nil {
			log.Fatal().
				Err(err).
				Msg("Unable to marshal structure to yaml")
		}
		break
	default:
		log.Fatal().
			Msg("Failed to write report, unknown type")
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
