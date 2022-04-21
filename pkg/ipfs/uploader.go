package ipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/go-git/go-billy/v5"
	cid "github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	files "github.com/ipfs/go-ipfs-files"
)

type Uploader struct {
	shell *shell.Shell
}

func NewUploader(shell *shell.Shell) *Uploader {
	return &Uploader{shell}
}

func (u *Uploader) UploadRepo(ctx context.Context, fs billy.Filesystem) (*cid.Cid, error) {
	root := "/"

	stat, err := fs.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve root stats: %w", err)
	}

	node, err := NewSerialFile(root, fs, stat)
	if err != nil {
		return nil, err
	}

	rootDir := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("/", node)})
	reader := files.NewMultiFileReader(rootDir, true)

	resp, err := u.shell.Request("add").
		Option("recursive", true).
		Body(reader).
		Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to save the data into ipfs: %w", err)
	}
	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	dec := json.NewDecoder(resp.Output)
	type object struct {
		Hash cid.Cid
	}
	var final cid.Cid
	for {
		var out object
		err = dec.Decode(&out)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		final = out.Hash
	}

	log.Printf("pin the repository: %s", final)
	err = u.shell.Pin(final.String())
	if err != nil {
		return nil, fmt.Errorf("failed to pin the repo: %w", err)
	}

	return &final, nil
}
