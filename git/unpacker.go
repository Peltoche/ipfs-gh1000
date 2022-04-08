package git

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/objfile"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type Unpacker struct {
}

func NewUnpacker() *Unpacker {
	return &Unpacker{}
}

func (u *Unpacker) Unpack(storage *filesystem.Storage) error {
	objs, err := storage.IterEncodedObjects(plumbing.AnyObject)
	if err != nil {
		return fmt.Errorf("failed to create an iterator for the encoded objects: %w", err)
	}

	err = objs.ForEach(func(obj plumbing.EncodedObject) error {
		file, err := u.createObjectFileForHash(storage.Filesystem(), obj.Hash())
		if err != nil {
			return err
		}

		writer := objfile.NewWriter(file)
		writer.WriteHeader(obj.Type(), obj.Size())

		objReader, err := obj.Reader()
		if err != nil {
			return fmt.Errorf("failed to retrieve writer for object %q: %w", obj.Hash(), err)
		}

		buf := make([]byte, obj.Size())
		n, err := objReader.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to write the object %q inside a buffer: %w", obj.Hash(), err)
		}
		err = objReader.Close()
		if err != nil {
			return fmt.Errorf("failed to close the object: %w", err)
		}

		if int64(n) != obj.Size() {
			return fmt.Errorf("the buffer is not completly full after receiving the data from object %q (%d != %d)", obj.Hash(), n, obj.Size())
		}

		writer.Write(buf)
		err = writer.Close()
		if err != nil {
			return fmt.Errorf("failed to close the object writer: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to unpach the objects: %w", err)
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
