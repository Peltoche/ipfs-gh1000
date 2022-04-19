package ipfs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	files "github.com/ipfs/go-ipfs-files"
)

// serialFile implements Node, and reads from a path on the OS filesystem.
// No more than one file will be opened at a time.
type serialFile struct {
	fs    billy.Filesystem
	path  string
	files []os.FileInfo
	stat  os.FileInfo
}

type serialIterator struct {
	fs    billy.Filesystem
	files []os.FileInfo
	path  string

	curName string
	curFile files.Node

	err error
}

func NewSerialFile(path string, fs billy.Filesystem, stat os.FileInfo) (files.Node, error) {
	switch mode := stat.Mode(); {
	case mode.IsRegular():
		file, err := fs.Open(path)
		if err != nil {
			return nil, err
		}
		return files.NewReaderPathFile(path, file, stat)

	case mode.IsDir():
		// for directories, stat all of the contents first, so we know what files to
		// open when Entries() is called
		files, err := fs.ReadDir(path)
		if err != nil {
			return nil, err
		}

		return &serialFile{fs, path, files, stat}, nil

	case mode&os.ModeSymlink != 0:
		target, err := fs.Readlink(path)
		if err != nil {
			return nil, err
		}
		return files.NewLinkFile(target, stat), nil
	default:
		return nil, fmt.Errorf("unrecognized file type for %s: %s", path, mode.String())
	}
}

func (sf *serialFile) Close() error {
	return nil
}

func (sf *serialFile) Stat() os.FileInfo {
	return sf.stat
}

func (sf *serialFile) Size() (int64, error) {
	if !sf.stat.IsDir() {
		//something went terribly, terribly wrong
		return 0, errors.New("serialFile is not a directory")
	}

	log.Println("not implemented yet")

	return 42, nil

}

func (sf *serialFile) Entries() files.DirIterator {
	return &serialIterator{
		fs:    sf.fs,
		files: sf.files,
		path:  sf.path,
	}
}

func (it *serialIterator) Name() string {
	return it.curName
}

func (it *serialIterator) Node() files.Node {
	return it.curFile
}

func (it *serialIterator) Next() bool {
	// if there aren't any files left in the root directory, we're done
	if len(it.files) == 0 {
		return false
	}

	stat := it.files[0]
	it.files = it.files[1:]

	// open the next file
	filePath := filepath.ToSlash(filepath.Join(it.path, stat.Name()))

	// recursively call the constructor on the next file
	// if it's a regular file, we will open it as a ReaderFile
	// if it's a directory, files in it will be opened serially
	sf, err := NewSerialFile(filePath, it.fs, stat)
	if err != nil {
		it.err = err
		return false
	}

	it.curName = stat.Name()
	it.curFile = sf
	return true
}

func (di *serialIterator) Err() error {
	return di.err
}

var _ files.Directory = &serialFile{}
var _ files.DirIterator = &serialIterator{}
