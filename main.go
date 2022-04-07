package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Peltoche/ipfs-gh1000/metadata"
	"github.com/Peltoche/ipfs-gh1000/repo"
)

func main() {
	metaFetcher, err := metadata.NewFetcher("https://gitstar-ranking.com/repositories")
	if err != nil {
		log.Fatalf("failed to create the fetcher: %s", err)
	}

	repoFetcher := repo.NewFetcher("./repositories")
	err = repoFetcher.EnsureFolderExists()
	if err != nil {
		log.Fatalf("failed to ensure that the folder exists: %s", err)
	}

	ctx := context.Background()

	links, err := metaFetcher.FetchLinkPage(ctx)

	for i, link := range links {
		meta, err := metaFetcher.FetchMetadataForLink(ctx, link)
		if err != nil {
			log.Fatalf("failed to fetch metadata for %s: %s", link, err)
		}

		err = repoFetcher.FetchRepository(ctx, meta.RepositoryURL)
		if err != nil {
			log.Fatalf("failed to fetch the repository %q: %s", meta.RepositoryURL, err)
		}
		fmt.Printf("%d -> %+v\n", i, meta)
	}

}
