package main

import (
	"context"
	"log"

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
) error {
	ctx := context.Background()

	fs := memfs.New()
	storage := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())

	log.Printf("fetch first page")
	links, err := metaFetcher.FetchLinkPage(ctx)
	if err != nil {
		log.Fatalf("failed to fetch the first page: %s", err)
	}

	repoMetadataList := make([]metadata.RepoMetadata, len(links))

	for i, link := range links {
		log.Printf("fetch metadata for %s", link)
		meta, err := metaFetcher.FetchMetadataForLink(ctx, link)
		if err != nil {
			log.Fatalf("failed to fetch the metadatas for %s: %s", link, err)
		}

		repoMetadataList[i] = *meta
	}

	for _, repo := range repoMetadataList {
		log.Printf("start converting the repo %s", repo.RepositoryURL)

		log.Printf("start pulling repository...")
		err = gitFetcher.CloneRepositoryInto(ctx, repo.RepositoryURL, storage)
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
		hash, err := ipfsUploader.UploadRepo(ctx, fs)
		if err != nil {
			log.Fatalf("failed to upload the repo %q into ipfs: %s", repo.RepositoryURL, err)
		}
		log.Printf("ifps uploading successfull: %q", hash)
	}

	return nil
}