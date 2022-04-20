package main

import (
	"log"

	"github.com/Peltoche/ipfs-gh1000/git"
	"github.com/Peltoche/ipfs-gh1000/ipfs"
	"github.com/Peltoche/ipfs-gh1000/metadata"
	shell "github.com/ipfs/go-ipfs-api"
)

func main() {
	metaFetcher, err := metadata.NewFetcher("https://gitstar-ranking.com/repositories")
	if err != nil {
		log.Fatalf("failed to create the fetcher: %s", err)
	}

	shell := shell.NewLocalShell()
	unpacker := git.NewUnpacker()
	gitFetcher := git.NewFetcher()
	infoUpdater := git.NewServerInfoUpdater()
	ipfsUploader := ipfs.NewUploader(shell)

	ipfsIndexer, err := ipfs.NewIndexer(shell, "gh1000")
	if err != nil {
		log.Fatalf("failed to initiate the indexer: %s", err)
	}

	err = Run(metaFetcher, gitFetcher, unpacker, infoUpdater, ipfsUploader, ipfsIndexer)
	if err != nil {
		log.Fatal(err)
	}
}
