package packing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/retro-framework/go-retro/framework/retro"
)

// Hash returns a hashed in raw bytes (not hex encoded)
// and a HashAlgoName alias.
type Hash struct {
	AlgoName HashAlgoName
	B        []byte
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", h.String())), nil
}

func (h Hash) String() string {
	return fmt.Sprintf("%s:%x", h.AlgoName, h.B)
}

func (h Hash) ShortStr() string {
	return fmt.Sprintf("%s:%x", h.AlgoName, h.B[0:8])
}

func (h Hash) Bytes() []byte {
	return h.B
}

func NewHash(n HashAlgoName, b []byte) Hash {
	return Hash{n, b}
}

func HashStrToHash(str string) retro.Hash {
	parts := strings.Split(str, ":")
	decoded, err := hex.DecodeString(parts[1])
	if err != nil {
		// TODO: do this better, not many call sites, maybe :MUST: is wise?
		log.Fatal(err)
	}
	return NewHash(HashAlgoNameSHA256, decoded)
}

// TODO: make this respect algoname in the given string
func hashStr(str string) retro.Hash {
	var s = sha256.Sum256([]byte(str))
	return NewHash(HashAlgoNameSHA256, s[:])
}

func HashStr(str string) retro.Hash {
	return hashStr(str)
}
