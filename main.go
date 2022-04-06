package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Peltoche/ipfs-gh1000/fetcher"
)

func main() {
	fetcher, err := fetcher.New("https://gitstar-ranking.com/repositories")
	if err != nil {
		log.Fatalf("failed to create the fetcher: %s", err)
	}

	ctx := context.Background()

	links, err := fetcher.FetchLinkPage(ctx)

	for i, link := range links {
		meta, err := fetcher.FetchMetadataForLink(ctx, link)
		if err != nil {
			log.Fatalf("failed to fetch metadata for %s: %s", link, err)
		}

		fmt.Printf("%d -> %+v\n", i, meta)
	}

}
