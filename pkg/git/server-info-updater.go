package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type ServerInfoUpdater struct {
}

func NewServerInfoUpdater() *ServerInfoUpdater {
	return &ServerInfoUpdater{}
}

func (s *ServerInfoUpdater) UpdateServerInfo(storage *filesystem.Storage) error {
	err := storage.Filesystem().MkdirAll("info", 0755)
	if err != nil {
		return fmt.Errorf("failed to create the info folder: %w", err)
	}

	refsFile, err := storage.Filesystem().Create("info/refs")
	if err != nil {
		return fmt.Errorf("failed to create the refs file: %w", err)
	}

	refs, err := storage.IterReferences()
	if err != nil {
		return fmt.Errorf("failed to create an iterator on references: %w", err)
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.HashReference {
			return nil
		}

		newName := strings.Replace(ref.Name().String(), "refs/remotes/origin/", "refs/heads/", -1)

		_, err := refsFile.Write([]byte(fmt.Sprintf("%s\t%s\n", ref.Hash(), newName)))
		if err != nil {
			return fmt.Errorf("failed to append a ref to the \"info/refs\" file: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to populate the \"info/refs\" file: %w", err)
	}

	return nil
}
