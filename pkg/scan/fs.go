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

type FsScanner struct {
	Scanner *Scanner
	Result  []model.FileLeak
	Root    string
}

func (f *FsScanner) New(root, rulespath string, output ReportInterface, debug bool) {
	*f = FsScanner{
		Root: root,
		Scanner: &Scanner{RulesPath: rulespath,
			RuleSet: &model.RuleSet{},
			Debug:   debug,
			Output:  output,
		},
	}

	f.Scanner.RuleSet.ParseConfig(f.Scanner.RulesPath)
}

func (f *FsScanner) getFiles() []string {
	log.Info().
		Msg("Collecting all files")
	var files []string
	if err := filepath.Walk(f.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, blacklist := range f.Scanner.RuleSet.BlackListCompiled {
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

func (f *FsScanner) Scan(concurrent int) {
	start := time.Now()
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
	if f.Scanner.Debug {
		f.Scanner.ProgressBar = progressbar.Default(int64(len(files)), " scanning files")
	}

	// Divide the commit structure into equal size chunks
	// and for each chunk launch a go routine that analyses
	// each commit sequentially for rule breaks.
	var wg sync.WaitGroup
	res := make([][]model.FileLeak, concurrent+1)

	for i := 0; i < concurrent; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end >= len(files) {
			end = len(files)
		}
		leakChan := make(chan model.FileLeak)
		doneChan := make(chan bool)
		wg.Add(1)
		go f.scanChunk(start, end, files, leakChan, doneChan)
		go fileleakReader(leakChan, doneChan, &wg, res, i)
	}
	wg.Wait()

	for _, chunk := range res {
		f.Result = append(f.Result, chunk...)
	}
	if f.Scanner.Debug {
		f.Scanner.ProgressBar.Clear()
	}
	log.Info().
		Str("duration", time.Since(start).String()).
		Msg("Scan completed in")
	log.Info().
		Int("potential leaks", len(f.Result)).
		Msg("Found")
	f.Scanner.Output.Write(f)
}

func (f *FsScanner) scanChunk(j, e int, files []string, leakChan chan model.FileLeak, doneChan chan bool) {
	log.Info().
		Int("start_file", j).
		Int("end_file", e-1).
		Msg("Routine launched")
	for _, file := range files[j : e-1] {
		f.Scanner.RuleSet.ParseFile(file, leakChan)
	}
	if f.Scanner.Debug {
		f.Scanner.ProgressBar.Add(1)
	}
	doneChan <- true
}

func fileleakReader(leaksChan <-chan model.FileLeak, doneChan <-chan bool, task *sync.WaitGroup, res [][]model.FileLeak, idx int) {
	leaks := []model.FileLeak{}
	for {
		select {
		case leak := <-leaksChan:
			leaks = append(leaks, leak)
		case <-doneChan:
			res[idx] = leaks
			task.Done()
			break
		}
	}
}
