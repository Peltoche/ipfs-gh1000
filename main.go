package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"

	"github.com/Peltoche/ipfs-gh1000/git"
	"github.com/go-git/go-billy/v5/memfs"
)

func main() {
	ctx := context.Background()

	/* metaFetcher, err := metadata.NewFetcher("https://gitstar-ranking.com/repositories")
	if err != nil {
		log.Fatalf("failed to create the fetcher: %s", err)
	} */

	memFS := memfs.New()

	gitFetcher := git.NewFetcher("./repositories")

	repoURL, _ := url.Parse("https://github.com/Peltoche/ipfs-gh1000")
	// repoURL, _ := url.Parse("https://github.com/996icu/996.ICU")

	CID, err := gitFetcher.CloneRepository(ctx, repoURL, memFS)
	if err != nil {
		log.Fatalf("failed to clone the repository: %s", err)
	}

	fmt.Printf("CID: %s\n", CID)

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
