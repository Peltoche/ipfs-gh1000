package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"

	"github.com/Peltoche/ipfs-gh1000/git"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

func main() {
	ctx := context.Background()

	/* metaFetcher, err := metadata.NewFetcher("https://gitstar-ranking.com/repositories")
	if err != nil {
		log.Fatalf("failed to create the fetcher: %s", err)
	} */

	memFS := memfs.New()
	storage := filesystem.NewStorage(memFS, cache.NewObjectLRUDefault())

	unpacker := git.NewUnpacker()
	gitFetcher := git.NewFetcher()

	repoURL, _ := url.Parse("https://github.com/Peltoche/ipfs-gh1000")

	err := gitFetcher.CloneRepositoryInto(ctx, repoURL, storage)
	if err != nil {
		log.Fatalf("failed to clone the repository: %s", err)
	}

	err = unpacker.Unpack(storage)
	if err != nil {
		log.Fatalf("failed to unpack the repository: %s", err)
	}

	entries, _ := memFS.ReadDir("/objects")
	for _, entry := range entries {
		if !entry.IsDir() {
			fmt.Printf("file: %v %s\n", entry.IsDir(), entry.Name())
			continue
		}

		entries2, _ := memFS.ReadDir(path.Join("/objects", entry.Name()))
		for _, entry2 := range entries2 {
			fmt.Printf("file: %v %s\n", entry.IsDir(), path.Join("/objects", entry.Name(), entry2.Name()))
		}

	}
}
