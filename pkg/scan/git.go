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

// GitScanner stores configuration for scanning a git repository
type GitScanner struct {
	// Where to store local copy of the repo
	Cache string
	// Path to the rule definition file
	// If set to "" automatically used default resources.yaml embedded by packr
	RulesPath string
	Repo      *model.Repo
	RuleSet   *model.RuleSet
	Result    []model.Leak

	// Whether or not to display progressbar (mainly for testing)
	Debug       bool
	ProgressBar *progressbar.ProgressBar
	// Output writer interface
	Output ReportInterface
}

// Type returns the string type of the scanner ("git")
func (g GitScanner) Type() string {
	return "git"
}

// NewGitScanner creates a new git Scanner
func NewGitScanner(source, cache, rulespath string, output ReportInterface, debug bool) *GitScanner {
	g := &GitScanner{
		Repo:      &model.Repo{},
		Cache:     cache,
		RulesPath: rulespath,
		RuleSet:   &model.RuleSet{},
		Debug:     debug,
		Output:    output,
	}

	g.RuleSet.ParseConfig(g.RulesPath)
	g.Repo.Init(source, g.Cache)
	return g
}

// Scan iterates over each commits and use defined rules
// to analyse for possible leaks
func (g *GitScanner) Scan(concurrent int) {
	startTime := time.Now()
	commits := g.Repo.FetchCommits()
	chunkSize := len(commits) / concurrent
	if chunkSize == 0 {
		log.Fatal().
			Int("concurrent", concurrent).
			Int("n_commits", len(commits)).
			Msg("Amount of concurrent routines >> number of commits")
	}
	log.Info().
		Msg(fmt.Sprintf("Processing %d commits with chunk_size = %d", len(commits), chunkSize))

	// progress bar initialisation
	if g.Debug {
		g.ProgressBar = progressbar.Default(int64(len(commits)), " scanning commits")
	}

	// Divide the commit structure into equal size chunks
	// and for each chunk launch a go routine that analyses
	// each commit sequentially for rule breaks.
	var wg sync.WaitGroup
	res := make([][]model.Leak, concurrent+1)

	for i := 0; i < concurrent; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end >= len(commits) {
			end = len(commits)
		}
		leakChan := make(chan model.Leak)
		doneChan := make(chan bool)
		wg.Add(1)
		go g.scanChunk(start, end, commits, leakChan, doneChan)
		go leakReader(leakChan, doneChan, &wg, res, i)
	}
	wg.Wait()
	for _, chunk := range res {
		g.Result = append(g.Result, chunk...)
	}

	if g.Debug {
		g.ProgressBar.Clear()
	}
	log.Info().
		Str("duration", time.Since(startTime).String()).
		Msg("Scan completed in")
	log.Info().
		Int("potential leaks", len(g.Result)).
		Msg("Found")
	g.Output.Write(g)
}

func (g *GitScanner) scanChunk(j, e int, commits []*object.Commit, leakChan chan model.Leak, doneChan chan bool) {
	log.Trace().
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
			g.RuleSet.ParsePatch(patch, commits[j+idx+1], g.Repo, leakChan)
		}
		if g.Debug {
			g.ProgressBar.Add(1)
		}
	}
	doneChan <- true
}
