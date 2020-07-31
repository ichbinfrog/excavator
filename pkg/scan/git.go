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

type GitScanner struct {
	Cache   string
	Scanner *Scanner
	Repo    *model.Repo
	Result  []model.GitLeak
}

func (g *GitScanner) New(source, cache, rulespath string, output ReportInterface, debug bool) {
	*g = GitScanner{
		Repo:  &model.Repo{},
		Cache: cache,
		Scanner: &Scanner{
			RulesPath: rulespath,
			RuleSet:   &model.RuleSet{},
			Debug:     debug,
			Output:    output,
		},
	}

	g.Scanner.RuleSet.ParseConfig(g.Scanner.RulesPath)
	g.Repo.Init(source, g.Cache)
}

func (g *GitScanner) Scan(concurrent int) {
	start := time.Now()

	commits := g.Repo.FetchCommits()
	chunkSize := len(commits) / concurrent
	log.Info().
		Msg(fmt.Sprintf("Processing %d commits with chunk_size = %d", len(commits), chunkSize))

	// progress bar initialisation
	if g.Scanner.Debug {
		g.Scanner.ProgressBar = progressbar.Default(int64(len(commits)), " scanning commits")
	}

	// Divide the commit structure into equal size chunks
	// and for each chunk launch a go routine that analyses
	// each commit sequentially for rule breaks.
	var wg sync.WaitGroup
	res := make([][]model.GitLeak, concurrent+1)

	for i := 0; i < concurrent; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end >= len(commits) {
			end = len(commits)
		}
		leakChan := make(chan model.GitLeak)
		doneChan := make(chan bool)
		wg.Add(1)
		go g.scanChunk(start, end, commits, leakChan, doneChan)
		go gitLeakReader(leakChan, doneChan, &wg, res, i)
	}
	wg.Wait()

	for _, chunk := range res {
		g.Result = append(g.Result, chunk...)
	}
	if g.Scanner.Debug {
		g.Scanner.ProgressBar.Clear()
	}
	log.Info().
		Str("duration", time.Since(start).String()).
		Msg("Scan completed in")
	log.Info().
		Int("potential leaks", len(g.Result)).
		Msg("Found")
	g.Scanner.Output.Write(g)
}

func (g *GitScanner) scanChunk(j, e int, commits []*object.Commit, leakChan chan model.GitLeak, doneChan chan bool) {
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
			g.Scanner.RuleSet.ParsePatch(patch, commits[j+idx+1], g.Repo, leakChan)
		}
		if g.Scanner.Debug {
			g.Scanner.ProgressBar.Add(1)
		}
	}
	doneChan <- true
}

func gitLeakReader(leaksChan <-chan model.GitLeak, doneChan <-chan bool, task *sync.WaitGroup, res [][]model.GitLeak, idx int) {
	leaks := []model.GitLeak{}
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
