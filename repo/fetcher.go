package repo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

type Fetcher struct {
	path string
}

func NewFetcher(cachePath string) *Fetcher {
	return &Fetcher{
		path: cachePath,
	}
}

func (f *Fetcher) EnsureFolderExists() error {
	fileInfo, err := os.Stat(f.path)
	if os.IsNotExist(err) {
		err = os.Mkdir(f.path, 0755)
		if err != nil {
			return fmt.Errorf("failed to create the directory %q: %w", f.path, err)
		}

		return nil
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("%q is not a directory", f.path)
	}

	return nil
}

func (f *Fetcher) FetchRepository(ctx context.Context, repoURL *url.URL) error {
	log.Printf("start fetching %q\n", repoURL)

	exists, err := f.repoExists(repoURL)
	if err != nil {
		return fmt.Errorf("failed to check if the folder already exists: %w", err)
	}

	var repo *git.Repository

	if !exists {
		repo, err = f.initRepository(repoURL)
	} else {
		repo, err = f.openRepository(repoURL)
	}
	if err != nil {
		return fmt.Errorf("failed to open or init the repostory: %w", err)
	}

	err = repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to pull the latest changes: %w", err)
	}

	return nil
}

func (f *Fetcher) openRepository(url *url.URL) (*git.Repository, error) {
	folder := f.folderFromURL(url)

	repo, err := git.PlainOpen(folder)
	if err != nil {
		return nil, fmt.Errorf("failed to open the existing repository: %w", err)
	}

	return repo, nil
}

func (f *Fetcher) initRepository(url *url.URL) (*git.Repository, error) {
	folder := f.folderFromURL(url)

	err := os.MkdirAll(folder, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create the folder %q: %w", folder, err)
	}

	repo, err := git.PlainInit(folder, true)
	if err != nil {
		return nil, fmt.Errorf("failed to init the new repository: %w", err)
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{url.String()},
		Fetch: []config.RefSpec{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create the new remote: %w", err)
	}

	return repo, nil
}

func (f *Fetcher) repoExists(url *url.URL) (bool, error) {
	folder := f.folderFromURL(url)

	_, err := os.Stat(folder)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check the repository folder exists: %w", err)
	}

	return true, nil
}

func (f *Fetcher) folderFromURL(url *url.URL) string {
	return path.Join(f.path, url.Path)
}
