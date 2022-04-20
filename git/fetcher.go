package git

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type Fetcher struct {
	storage *filesystem.Storage
}

func NewFetcher() *Fetcher {
	return &Fetcher{}
}

func (f *Fetcher) CloneRepositoryInto(ctx context.Context, repoURL string, storage *filesystem.Storage) error {
	repo, err := git.Init(storage, nil)
	if err != nil {
		return fmt.Errorf("failed to init the new repository: %w", err)
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{repoURL},
		Fetch: []config.RefSpec{},
	})
	if err != nil {
		return fmt.Errorf("failed to create the new remote: %w", err)
	}

	err = repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName:      "origin",
		RefSpecs:        []config.RefSpec{},
		Depth:           0,
		Auth:            nil,
		Progress:        io.Discard,
		Tags:            0,
		Force:           false,
		InsecureSkipTLS: false,
		CABundle:        []byte{},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to pull the latest changes: %w", err)
	}

	return nil
}
