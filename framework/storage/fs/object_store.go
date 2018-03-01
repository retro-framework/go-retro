package fs

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/retro-framework/go-retro/framework/packing"
)

var (
	ErrUnableToCompletelyWriteObject = errors.New("unable to completely write object")
	ErrUnableToWriteObject           = errors.New("unable to write object")
	ErrUnableToCreateObjectFile      = errors.New("unable to create object file")
	ErrUnableToCreateObjectDir       = errors.New("unable to create object dir")

	ErrNoSuchObject                  = errors.New("no such object in object database")
	ErrBadObjectHashForRetrieve      = errors.New("no valid object hash when looking up packed")
	ErrUnableToDecodeHashForRetrieve = errors.New("unable to decode hash when looking up packed")
	ErrUnableToInflateObject         = errors.New("error running zlib inflate")
	ErrUnsupportedHash               = errors.New("only supports sha256 hash")
	ErrUnableToReadObjectFile        = errors.New("unable to read object file")
)

type ObjectStore struct {
	BasePath string
}

func (s *ObjectStore) mkdirAll(path string) error {
	return os.MkdirAll(path, 0766)
}

func (s *ObjectStore) WritePacked(p packing.PackedObject) (int, error) {

	// TODO: What if basepath points to a _file_ not a dir?
	if _, err := os.Stat(s.BasePath); os.IsNotExist(err) {
		if err := s.mkdirAll(s.BasePath); err != nil {
			return 0, ErrUnableToCreateBaseDir
		}
	}

	var (
		b bytes.Buffer
	)

	w := zlib.NewWriter(&b)
	w.Write(p.Contents())
	w.Close()

	var (
		objPath = filepath.Join(s.BasePath, p.Hash().ToPathName())
		objDir  = filepath.Dir(objPath)
	)

	if _, err := os.Stat(objDir); os.IsNotExist(err) {
		if err := s.mkdirAll(objDir); err != nil {
			return 0, ErrUnableToCreateObjectDir
		}
	}

	if _, err := os.Stat(objPath); os.IsNotExist(err) {
		f, err := os.Create(objPath)
		if err != nil {
			return 0, ErrUnableToCreateObjectFile
		}
		defer f.Close()

		n, err := f.Write(b.Bytes())
		if err != nil {
			return 0, ErrUnableToWriteObject
		}

		if n != len(b.Bytes()) {
			return 0, ErrUnableToCompletelyWriteObject
		}

		return n, nil
	}

	return 0, nil
}

// TODO: should also parse the aglo out of the string and set the PO Hash
// algo/etc to the right values., the new PackedObject could be kept and
// maybe simply take an AlgoName in the second position?
func (s *ObjectStore) RetrievePacked(str string) (*packing.PackedObject, error) {

	parts := strings.Split(str, ":") // ["sha256", "hexbyteshexbtytes"]
	if len(parts) != 2 {
		return nil, ErrBadObjectHashForRetrieve
	}

	dst := make([]byte, hex.DecodedLen(len(parts[1])))
	_, err := hex.Decode(dst, []byte(parts[1]))
	if err != nil {
		return nil, ErrUnableToDecodeHashForRetrieve
	}

	if parts[0] != string(packing.HashAlgoNameSHA256) {
		return nil, ErrUnsupportedHash
	}

	h := packing.Hash{
		AlgoName: packing.HashAlgoNameSHA256,
		Bytes:    dst,
	}

	objPath := filepath.Join(s.BasePath, h.ToPathName())

	content, err := ioutil.ReadFile(objPath)
	if err != nil {
		return nil, ErrUnableToReadObjectFile
	}

	b := bytes.NewReader([]byte(content))
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, ErrUnableToInflateObject
	}
	r.Close()

	// TODO: Check for err here, can this really fail with a local
	// buffer?
	orig, _ := ioutil.ReadAll(r)

	po := packing.NewPackedObject(string(orig))
	return &po, nil
}
