package scan

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ichbinfrog/excavator/pkg/model"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

// FsScanner stores configuration for scanning a local directory in the fs
type FsScanner struct {
	RulesPath string
	Root      string
	Result    []model.Leak

	Debug       bool
	ProgressBar *progressbar.ProgressBar
	RuleSet     *model.RuleSet
	Output      ReportInterface
}

// Type returns the string type of the scanner ("fs")
func (f FsScanner) Type() string {
	return "fs"
}

// NewFsScanner creates a new FsScanner
func NewFsScanner(root, rulespath string, output ReportInterface, debug bool) *FsScanner {
	f := &FsScanner{
		Root:      root,
		RulesPath: rulespath,
		RuleSet:   &model.RuleSet{},
		Debug:     debug,
		Output:    output,
	}

	f.RuleSet.ParseConfig(f.RulesPath)
	return f
}

func (f *FsScanner) getFiles() []string {
	log.Info().
		Msg("Collecting all files")
	var files []string
	if err := filepath.Walk(f.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, blacklist := range f.RuleSet.BlackListCompiled {
			if blacklist.MatchString(path) {
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}
		switch info.Mode() {
		case os.ModeSymlink:
			log.Trace().
				Str("path", path).
				Msg("Skipping symlink @")
			return nil
		case os.ModeIrregular:
			log.Trace().
				Str("path", path).
				Msg("Skipping irregular file @")
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		log.Fatal().
			Str("path", f.Root).
			Err(err).
			Msg("Failed to recursively files under")
	}
	return files
}

// Scan iterates over each file recursively and use defined rules
// to analyse for possible leaks
func (f *FsScanner) Scan(concurrent int) {
	startTime := time.Now()
	files := f.getFiles()
	chunkSize := len(files) / concurrent
	if chunkSize == 0 {
		log.Fatal().
			Int("concurrent", concurrent).
			Int("n_files", len(files)).
			Msg("Amount of concurrent routines >> number of files")
	}
	log.Info().
		Msg(fmt.Sprintf("Processing %d files with chunk_size = %d", len(files), chunkSize))

	// progress bar initialisation
	if f.Debug {
		f.ProgressBar = progressbar.Default(int64(len(files)), " scanning files")
	}

	// Divide the file list into equal size chunks
	// and for each chunk launch a go routine that analyses
	// each file sequentially for rule breaks.
	var wg sync.WaitGroup
	res := make([][]model.Leak, concurrent+1)

	for i := 0; i < concurrent; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end >= len(files) {
			end = len(files)
		}
		leakChan := make(chan model.Leak)
		doneChan := make(chan bool)
		wg.Add(1)
		go f.scanChunk(start, end, files, leakChan, doneChan)
		go leakReader(leakChan, doneChan, &wg, res, i)
	}
	wg.Wait()

	for _, chunk := range res {
		f.Result = append(f.Result, chunk...)
	}
	if f.Debug {
		f.ProgressBar.Clear()
	}
	log.Info().
		Str("duration", time.Since(startTime).String()).
		Msg("Scan completed in")
	log.Info().
		Int("potential leaks", len(f.Result)).
		Msg("Found")
	f.Output.Write(f)
}

func (f FsScanner) scanChunk(j, e int, files []string, leakChan chan model.Leak, doneChan chan bool) {
	log.Trace().
		Int("start_file", j).
		Int("end_file", e).
		Msg("Routine launched")
	for _, file := range files[j:e] {
		f.RuleSet.Parse(file, leakChan)
		if f.Debug {
			f.ProgressBar.Add(1)
		}
	}
	doneChan <- true
}
