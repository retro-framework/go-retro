package fs

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/retro-framework/go-retro/framework/packing"
	"github.com/retro-framework/go-retro/framework/types"
)

var (
	ErrUnableToCompletelyWriteRef = errors.New("unable to completely write ref")
	ErrUnableToWriteRef           = errors.New("unable to write ref")
	ErrUnableToCreateRefFile      = errors.New("unable to create ref file")
	ErrUnableToCreateRefDir       = errors.New("unable to create ref dir")

	ErrNoSuchRef           = errors.New("no such ref in database")
	ErrUnableToReadRefFile = errors.New("unable to read ref file")
	ErrBadHashForRetrieve  = errors.New("no valid hash in ref ")
)

type RefStore struct {
	BasePath string
}

func (r *RefStore) mkdirAll(path string) error {
	return os.MkdirAll(path, 0766)
}

func (s *RefStore) Ls() (map[string]types.Hash, error) {
	var hashes = make(map[string]types.Hash)
	files, err := filepath.Glob(fmt.Sprintf("%s/refs/**/*", s.BasePath))
	if err != nil {
		return hashes, err // TODO: Wrap me
	}
	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			return nil, err
		}
		switch mode := fi.Mode(); {
		case mode.IsRegular():
			contents, err := ioutil.ReadFile(file)
			if err != nil {
				return nil, err // TODO: wrap me
			}
			hashes[strings.TrimPrefix(file, fmt.Sprintf("%s/", s.BasePath))] = packing.HashStrToHash(string(contents))
		}
	}
	return hashes, nil
}

func (s *RefStore) Write(name string, hash types.Hash) (bool, error) {

	// TODO: What if basepath points to a _file_ not a dir?
	if _, err := os.Stat(s.BasePath); os.IsNotExist(err) {
		if err := s.mkdirAll(s.BasePath); err != nil {
			return false, ErrUnableToCreateBaseDir
		}
	}

	var (
		refPath = filepath.Join(s.BasePath, name)
		refDir  = filepath.Dir(refPath)
	)

	var writeRefFile = func() error {
		f, err := os.Create(refPath)
		if err != nil {
			return ErrUnableToCreateRefFile
		}
		defer f.Close()
		n, err := f.WriteString(hash.String())
		if err != nil {
			return ErrUnableToWriteRef
		}

		if n != len(hash.String()) {
			return ErrUnableToCompletelyWriteRef
		}

		return nil
	}

	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		if err := s.mkdirAll(refDir); err != nil {
			return false, ErrUnableToCreateRefDir
		}
		return true, writeRefFile()
	}

	// Check if we have to write the file, read it first
	fileData, err := ioutil.ReadFile(refPath)
	if err != nil {
		return false, ErrUnableToReadRefFile
	}
	if string(fileData) != hash.String() {
		return true, writeRefFile()
	}
	return false, nil

}

func (s *RefStore) WriteSymbolic(name, ref string) (bool, error) {

	// TODO: What if basepath points to a _file_ not a dir?
	if _, err := os.Stat(s.BasePath); os.IsNotExist(err) {
		if err := s.mkdirAll(s.BasePath); err != nil {
			return false, ErrUnableToCreateBaseDir
		}
	}

	var (
		symRefPath = filepath.Join(s.BasePath, name)
		symRefDir  = filepath.Dir(symRefPath)
	)

	symRefContents := fmt.Sprintf("ref: %s", ref)

	var writeRefFile = func() error {
		f, err := os.Create(symRefPath)
		if err != nil {
			return ErrUnableToCreateRefFile
		}
		defer f.Close()

		n, err := f.WriteString(symRefContents)
		if err != nil {
			return ErrUnableToWriteRef
		}

		if n != len(ref)+5 {
			return ErrUnableToCompletelyWriteRef
		}

		return nil
	}

	if _, err := os.Stat(symRefPath); os.IsNotExist(err) {
		if err := s.mkdirAll(symRefDir); err != nil {
			return false, ErrUnableToCreateRefDir
		}
		return true, writeRefFile()
	}

	// Check if we have to write the file, read it first
	fileData, err := ioutil.ReadFile(symRefPath)
	if err != nil {
		return false, ErrUnableToReadRefFile
	}
	if string(fileData) != symRefContents {
		return true, writeRefFile()
	}
	return false, nil
}

func (s *RefStore) Retrieve(name string) (types.Hash, error) {

	var refPath = filepath.Join(s.BasePath, name)

	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		return nil, ErrNoSuchRef
	}

	hashData, err := ioutil.ReadFile(refPath)
	if err != nil {
		return nil, ErrUnableToReadRefFile
	}

	parts := strings.Split(string(hashData), ":") // ["sha256", "hexbyteshexbtytes"]
	if len(parts) != 2 {
		return nil, ErrBadHashForRetrieve
	}

	dst := make([]byte, hex.DecodedLen(len(parts[1])))
	_, err = hex.Decode(dst, []byte(parts[1]))
	if err != nil {
		return nil, ErrUnableToDecodeHashForRetrieve
	}

	return packing.NewHash(packing.HashAlgoNameSHA256, dst), nil
}

func (s *RefStore) RetrieveSymbolic(name string) (string, error) {

	var symRefPath = filepath.Join(s.BasePath, name)

	if _, err := os.Stat(symRefPath); os.IsNotExist(err) {
		return "", ErrNoSuchRef
	}

	hashData, err := ioutil.ReadFile(symRefPath)
	if err != nil {
		return "", ErrUnableToReadRefFile
	}

	parts := strings.Split(string(hashData), ": ") // ["sha256", "hexbyteshexbtytes"]
	if len(parts) != 2 {
		return "", ErrBadHashForRetrieve
	}

	return parts[1], nil

}
