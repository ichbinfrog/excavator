package model

import (
	"net/url"
	"path"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
)

type Repo struct {
	Source string `yaml:"source"`
	Path   string
	Since  time.Time
	Storer *git.Repository
}

func (r *Repo) Init(source, cache string) {
	r.Source = source

	if _, err := url.Parse(source); err != nil {
		log.Warn().
			Err(err).
			Msg("Git source given is not a remote URL, interpreting as local")
		r.Path, _ = filepath.Abs(filepath.Base(source))
		r.Storer = r.PlainOpen(source)
	} else {
		r.Path = path.Join(cache, path.Base(r.Source))
		log.Info().
			Str("repository", r.Path).
			Msg("Cloning")

		ref, err := git.PlainClone(r.Path, false, &git.CloneOptions{
			URL: r.Source,
		})
		if err != nil {
			if err == git.ErrRepositoryAlreadyExists {
				r.Storer = r.PlainOpen(r.Path)
				return
			}
			log.Fatal().
				Err(err).
				Msg("Failed to open repository")
		}
		r.Storer = ref
	}
}

func (r *Repo) PlainOpen(repoPath string) *git.Repository {
	ref, err := git.PlainOpen(r.Path)
	if err != nil {
		log.Fatal().
			Str("path", r.Path).
			Err(err).
			Msg("Repo failed to open @")
	}
	return ref
}

func (r *Repo) FetchCommits() []*object.Commit {
	// Fetch repository head
	ref, err := r.Storer.Head()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Unable to fetch head")
	}
	commitIter, _ := r.Storer.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})

	// Append all commits to a slice
	// (Allows for better concurrent access than the base commitIter)
	commits := []*object.Commit{}
	commitIter.ForEach(func(o *object.Commit) error {
		commits = append(commits, o)
		return nil
	})
	commitIter.Close()
	return commits
}
