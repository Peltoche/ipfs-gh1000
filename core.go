package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/Peltoche/ipfs-gh1000/git"
	"github.com/Peltoche/ipfs-gh1000/ipfs"
	"github.com/Peltoche/ipfs-gh1000/metadata"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

func Run(
	metaFetcher *metadata.Fetcher,
	gitFetcher *git.Fetcher,
	unpacker *git.Unpacker,
	infoUpdater *git.ServerInfoUpdater,
	ipfsUploader *ipfs.Uploader,
	indexer *ipfs.Indexer,
) error {
	ctx := context.Background()

	log.Printf("fetch first page")
	links, err := metaFetcher.FetchLinkPage(ctx)
	if err != nil {
		log.Fatalf("failed to fetch the first page: %s", err)
	}

	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	for i := range links {
		n := r.Intn(len(links) - 1)
		links[i], links[n] = links[n], links[i]
	}

	for _, link := range links {
		fs := memfs.New()
		storage := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())

		log.Printf("fetch metadata for %s", link)
		meta, err := metaFetcher.FetchMetadataForLink(ctx, link)
		if err != nil {
			log.Fatalf("failed to fetch the metadatas for %s: %s", link, err)
		}

		log.Printf("start converting the repo %s", meta.RepositoryURL)

		log.Printf("start pulling repository...")
		err = gitFetcher.CloneRepositoryInto(ctx, meta.RepositoryURL, storage)
		if err != nil {
			log.Fatalf("failed to clone the repository: %s", err)
		}
		log.Println("pull successfull")

		log.Println("unpack repository...")
		err = unpacker.Unpack(storage.Filesystem())
		if err != nil {
			log.Fatalf("failed to unpack the repository: %s", err)
		}
		log.Println("unpack successfull")

		log.Println("start updating server infos")
		err = infoUpdater.UpdateServerInfo(storage)
		if err != nil {
			log.Fatalf("failed to update the server infos: %s", err)
		}
		log.Println("server info updating successfull")

		log.Println("start ipfs uploading")
		repoCID, err := ipfsUploader.UploadRepo(ctx, fs)
		if err != nil {
			log.Fatalf("failed to upload the repo %q into ipfs: %s", meta.RepositoryURL, err)
		}
		log.Printf("ifps uploading successfull: %q", repoCID)

		meta.Repo = repoCID

		log.Println("start updating the index")
		index, err := indexer.RetrieveIndex(ctx)
		if err != nil {
			log.Fatalf("failed to retrieve the index: %s", err)
		}

		log.Printf("index: %#v\n\n", index)
		index[link] = *meta

		err = indexer.SaveIndex(ctx, index)
		if err != nil {
			log.Fatalf("failed to save the new index: %s", err)
		}
	}

	return nil
}
