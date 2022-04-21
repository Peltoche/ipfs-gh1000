package git

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/idxfile"
	"github.com/go-git/go-git/v5/plumbing/format/objfile"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
)

type Unpacker struct {
}

func NewUnpacker() *Unpacker {
	return &Unpacker{}
}

func (u *Unpacker) Unpack(storage billy.Filesystem) error {
	dir := dotgit.New(storage)

	packsHashs, err := dir.ObjectPacks()
	if err != nil {
		return fmt.Errorf("failed to retrieve the pack hashs: %w", err)
	}

	for _, packHash := range packsHashs {
		packFil, err := dir.ObjectPack(packHash)
		if err != nil {
			return fmt.Errorf("failed to retrieve the packfile %q: %w", packHash, err)
		}

		idxFile, err := dir.ObjectPackIdx(packHash)
		if err != nil {
			return fmt.Errorf("failed to retrieve the idx file for pack %q: %w", packHash, err)
		}

		idx := idxfile.NewMemoryIndex()
		err = idxfile.NewDecoder(idxFile).Decode(idx)
		if err != nil {
			return fmt.Errorf("failed to decode the idx file for pack %q: %w", packHash, err)
		}

		pack := packfile.NewPackfile(idx, nil, packFil)
		objs, err := pack.GetAll()
		if err != nil {
			return fmt.Errorf("failed to fetch all objects from pack %q: %w", packHash, err)
		}

		err = objs.ForEach(func(obj plumbing.EncodedObject) error {
			file, err := u.createObjectFileForHash(storage, obj.Hash())
			if err != nil {
				return err
			}

			writer := objfile.NewWriter(file)
			writer.WriteHeader(obj.Type(), obj.Size())

			objReader, err := obj.Reader()
			if err != nil {
				return fmt.Errorf("failed to retrieve writer for object %q: %w", obj.Hash(), err)
			}

			buf, err := ioutil.ReadAll(objReader)
			if err != nil {
				return fmt.Errorf("failed to write the object %q inside a buffer: %w", obj.Hash(), err)
			}

			err = objReader.Close()
			if err != nil {
				return fmt.Errorf("failed to close the object: %w", err)
			}

			if int64(len(buf)) != obj.Size() {
				return fmt.Errorf("the buffer is not completly full after receiving the data from object %q (%d != %d)", obj.Hash(), len(buf), obj.Size())
			}

			writer.Write(buf)
			err = writer.Close()
			if err != nil {
				return fmt.Errorf("failed to close the object writer: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to unpack the objects: %w", err)
		}

		err = dir.DeleteOldObjectPackAndIndex(packHash, time.Time{})
		if err != nil {
			return fmt.Errorf("failed to delete the packfile %q and the corresponding index: %w", packHash, err)
		}
	}

	return nil
}

func (u *Unpacker) createObjectFileForHash(fs billy.Filesystem, hash plumbing.Hash) (billy.File, error) {
	prefix := hash.String()[0:2]
	fileName := hash.String()[2:]

	err := fs.MkdirAll(path.Join("objects/", prefix), 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create the dir for hash %q: %w", hash, err)
	}

	filePath := path.Join("objects/", prefix, fileName)

	_, err = fs.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		file, err := fs.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create the file for object %q: %w", hash, err)
		}

		return file, nil
	}

	file, err := fs.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file for object %q: %w", hash, err)
	}

	return file, nil
}
