package scan

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/ichbinfrog/excavator/pkg/model"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

type Scanner struct {
	Cache     string
	RulesPath string

	Debug       bool
	ProgressBar *progressbar.ProgressBar
	Repo        *model.Repo
	RuleSet     *model.RuleSet
	Output      ReportInterface

	Result []model.Leak
}

func (s *Scanner) New(source, cache, rulespath string, output ReportInterface, debug bool) {
	*s = Scanner{
		Cache:     cache,
		RulesPath: rulespath,
		Repo:      &model.Repo{},
		RuleSet:   &model.RuleSet{},
		Debug:     debug,
		Output:    output,
	}

	s.RuleSet.ParseConfig(s.RulesPath)
	s.Repo.Init(source, s.Cache)
}

func (s *Scanner) Scan(concurrent int) {
	start := time.Now()

	commits := s.Repo.FetchCommits()
	chunkSize := len(commits) / concurrent
	log.Info().Msg(fmt.Sprintf("Processing %d commits with chunk_size = %d", len(commits), chunkSize))

	// progress bar initialisation
	if s.Debug {
		s.ProgressBar = progressbar.Default(int64(len(commits)), " scanning commits")
	}

	// Divide the commit structure into equal size chunks
	// and for each chunk launch a go routine that analyses
	// each commit sequentially for rule breaks.
	var wg sync.WaitGroup
	res := make([][]model.Leak, concurrent+1)

	for i := 0; i <= concurrent; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end >= len(commits) {
			end = len(commits)
		}
		leakChan := make(chan model.Leak)
		doneChan := make(chan bool)
		wg.Add(1)
		go s.scanChunk(start, end, commits, leakChan, doneChan)
		go leakReader(leakChan, doneChan, &wg, res, i)
	}
	wg.Wait()

	for _, chunk := range res {
		s.Result = append(s.Result, chunk...)
	}
	if s.Debug {
		s.ProgressBar.Clear()
	}
	log.Info().
		Str("duration", time.Since(start).String()).
		Msg("Scan completed in")
	log.Info().
		Int("potential leaks", len(s.Result)).
		Msg("Found")
	s.Output.Write(s)
}

func (s *Scanner) scanChunk(j, e int, commits []*object.Commit, leakChan chan model.Leak, doneChan chan bool) {
	log.Info().
		Int("start_commit", j).
		Int("end_commit", e-1).
		Msg("Routine launched")
	for idx, commit := range commits[j : e-1] {
		currentTree, err := commit.Tree()
		if err != nil {
			log.Error().
				Int("commit_index", j+idx).
				Err(err).
				Msg("Unable to fetch tree from current commit")
			continue
		}
		nextTree, err := commits[j+idx+1].Tree()
		if err != nil {
			log.Error().
				Int("commit_index", j+idx+1).
				Err(err).
				Msg("Unable to fetch tree from next commit")
			continue
		}
		changes, err := currentTree.Diff(nextTree)
		if err != nil {
			log.Error().
				Int("first_tree", j+idx).
				Int("next_tree", j+idx+1).
				Err(err).Msg("Unable to generate a tree diff")
			continue
		}
		for _, ch := range changes {
			patch, _ := ch.Patch()
			s.RuleSet.ParsePatch(patch, commits[j+idx+1], s.Repo, leakChan)
		}
		if s.Debug {
			s.ProgressBar.Add(1)
		}
	}
	doneChan <- true
}

func leakReader(leaksChan <-chan model.Leak, doneChan <-chan bool, task *sync.WaitGroup, res [][]model.Leak, idx int) {
	leaks := []model.Leak{}
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
