package main

import (
	"context"
	"log"
	"net/url"

	"github.com/Peltoche/ipfs-gh1000/git"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

func main() {
	ctx := context.Background()

	/* metaFetcher, err := metadata.NewFetcher("https://gitstar-ranking.com/repositories")
	if err != nil {
		log.Fatalf("failed to create the fetcher: %s", err)
	} */

	// memFS := memfs.New()
	memFS := osfs.New("/tmp/foobar")
	storage := filesystem.NewStorage(memFS, cache.NewObjectLRUDefault())

	unpacker := git.NewUnpacker()
	gitFetcher := git.NewFetcher()
	infoUpdater := git.NewServerInfoUpdater()

	repoURL, _ := url.Parse("https://github.com/vuejs/vue")

	err := gitFetcher.CloneRepositoryInto(ctx, repoURL, storage)
	if err != nil {
		log.Fatalf("failed to clone the repository: %s", err)
	}

	err = unpacker.Unpack(storage.Filesystem())
	if err != nil {
		log.Fatalf("failed to unpack the repository: %s", err)
	}

	err = infoUpdater.UpdateServerInfo(storage)
	if err != nil {
		log.Fatalf("failed to update the server infos: %s", err)
	}
}
