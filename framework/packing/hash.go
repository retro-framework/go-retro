package packing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/retro-framework/go-retro/framework/types"
)

// Hash returns a hashed in raw bytes (not hex encoded)
// and a HashAlgoName alias.
type Hash struct {
	AlgoName HashAlgoName
	Bytes    []byte
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", h.String())), nil
}

func (h Hash) String() string {
	return fmt.Sprintf("%s:%x", h.AlgoName, h.Bytes)
}

// ToPathName yields a pathname such as
// "f63b/82de/c4c45a502655369ca20af061d08c4459b108f87a108aa1d1dd4c02a0"
// which is intended to avoid having filesystem directories containing
// millions of entries. Git uses a similar scheme using the first bytes
// for a two-level hierarchy. Because Retro prefers SHA256 which has a
// bigger space (longer hashes) a two-level hierarchy seemed prudent.
// func (h Hash) ToPathName() string {
// 	return fmt.Sprintf("%x/%x/%x", h.Bytes[0:2], h.Bytes[2:4], h.Bytes[4:])
// }

func HashStrToHash(str string) types.Hash {
	parts := strings.Split(str, ":")
	decoded, err := hex.DecodeString(parts[1])
	if err != nil {
		// TODO: do this better, not many call sites, maybe :MUST: is wise?
		log.Fatal(err)
	}
	return Hash{AlgoName: HashAlgoNameSHA256, Bytes: decoded}
}

// TODO: make this respect algoname in the given string
func hashStr(str string) types.Hash {
	var s = sha256.Sum256([]byte(str))
	return Hash{AlgoName: HashAlgoNameSHA256, Bytes: s[:]}
}

func HashStr(str string) types.Hash {
	return hashStr(str)
}
