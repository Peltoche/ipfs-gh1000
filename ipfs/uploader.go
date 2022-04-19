package ipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-git/go-billy/v5"
	shell "github.com/ipfs/go-ipfs-api"
	files "github.com/ipfs/go-ipfs-files"
)

type Uploader struct {
	shell *shell.Shell
}

func NewUploader() *Uploader {
	return &Uploader{
		shell: shell.NewLocalShell(),
	}
}

func (u *Uploader) UploadRepo(ctx context.Context, fs billy.Filesystem) (string, error) {
	root := "/"

	stat, err := fs.Stat(root)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve root stats: %w", err)
	}

	node, err := NewSerialFile(root, fs, stat)
	if err != nil {
		return "", err
	}

	rootDir := files.NewSliceDirectory([]files.DirEntry{files.FileEntry("/", node)})
	reader := files.NewMultiFileReader(rootDir, true)

	resp, err := u.shell.Request("add").
		Option("recursive", true).
		Option("pin", true).
		Body(reader).
		Send(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to save the data into ipfs: %w", err)
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", resp.Error
	}

	dec := json.NewDecoder(resp.Output)
	type object struct {
		Hash string
	}
	var final string
	for {
		var out object
		err = dec.Decode(&out)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		final = out.Hash
	}

	return final, nil
}
