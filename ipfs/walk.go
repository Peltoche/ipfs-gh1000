package ipfs

import (
	"path/filepath"

	files "github.com/ipfs/go-ipfs-files"
)

func Walk(nd files.Node, cb func(fpath string, nd files.Node) error) error {
	var helper func(string, files.Node) error
	helper = func(path string, nd files.Node) error {
		if err := cb(path, nd); err != nil {
			return err
		}
		dir, ok := nd.(files.Directory)
		if !ok {
			return nil
		}
		iter := dir.Entries()
		for iter.Next() {
			if err := helper(filepath.Join(path, iter.Name()), iter.Node()); err != nil {
				return err
			}
		}
		return iter.Err()
	}
	return helper("", nd)
}
