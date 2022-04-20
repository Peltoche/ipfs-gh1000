package main

import (
	"log"

	"github.com/Peltoche/ipfs-gh1000/git"
	"github.com/Peltoche/ipfs-gh1000/ipfs"
	"github.com/Peltoche/ipfs-gh1000/metadata"
)

func main() {
	metaFetcher, err := metadata.NewFetcher("https://gitstar-ranking.com/repositories")
	if err != nil {
		log.Fatalf("failed to create the fetcher: %s", err)
	}

	unpacker := git.NewUnpacker()
	gitFetcher := git.NewFetcher()
	infoUpdater := git.NewServerInfoUpdater()
	ipfsUploader := ipfs.NewUploader()

	err = Run(metaFetcher, gitFetcher, unpacker, infoUpdater, ipfsUploader)
	if err != nil {
		log.Fatal(err)
	}
}
