package git

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/format/objfile"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type Fetcher struct {
	path string
}

func NewFetcher(cachePath string) *Fetcher {
	return &Fetcher{
		path: cachePath,
	}
}

func (f *Fetcher) CloneRepository(ctx context.Context, repoURL *url.URL, fs billy.Filesystem) (string, error) {
	log.Printf("start fetching %q\n", repoURL)

	storage := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())

	repo, err := git.Init(storage, nil)
	if err != nil {
		return "", fmt.Errorf("failed to init the new repository: %w", err)
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{repoURL.String()},
		Fetch: []config.RefSpec{},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create the new remote: %w", err)
	}

	err = repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return "", fmt.Errorf("failed to pull the latest changes: %w", err)
	}

	objs, err := storage.IterEncodedObjects(plumbing.AnyObject)
	if err != nil {
		return "", fmt.Errorf("failed to create an iterator for the encoded objects: %w", err)
	}

	err = objs.ForEach(func(obj plumbing.EncodedObject) error {
		log.Printf("obj %s %s %d\n", obj.Hash(), obj.Type(), obj.Size())
		file, err := f.createObjectFileForHash(storage.Filesystem(), obj.Hash())
		if err != nil {
			return err
		}

		writer := objfile.NewWriter(file)
		writer.WriteHeader(obj.Type(), obj.Size())

		objReader, err := obj.Reader()
		if err != nil {
			return fmt.Errorf("failed to retrieve writer for object %q: %w", obj.Hash(), err)
		}

		buf := make([]byte, obj.Size())
		n, err := objReader.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to write the object %q inside a buffer: %w", obj.Hash(), err)
		}
		err = objReader.Close()
		if err != nil {
			return fmt.Errorf("failed to close the object: %w", err)
		}

		if int64(n) != obj.Size() {
			return fmt.Errorf("the buffer is not completly full after receiving the data from object %q (%d != %d)", obj.Hash(), n, obj.Size())
		}

		writer.Write(buf)
		err = writer.Close()
		if err != nil {
			return fmt.Errorf("failed to close the object writer: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to unpach the objects: %w", err)
	}

	return "some-CID", nil
}

func (f *Fetcher) createObjectFileForHash(fs billy.Filesystem, hash plumbing.Hash) (billy.File, error) {
	prefix := hash.String()[0:2]
	fileName := hash.String()[2:]

	err := fs.MkdirAll(path.Join("objects/", prefix), 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create the dir for hash %q: %w", hash, err)
	}

	filePath := path.Join("objects/", prefix, fileName)

	_, err = fs.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		file, err := fs.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create the file for object %q: %w", hash, err)
		}

		return file, nil
	}

	file, err := fs.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file for object %q: %w", hash, err)
	}

	return file, nil
}
