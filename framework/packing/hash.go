package packing

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Hash returns a hashed in raw bytes (not hex encoded)
// and a HashAlgoName alias.
type Hash struct {
	AlgoName HashAlgoName
	Bytes    []byte
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
func (h Hash) ToPathName() string {
	return fmt.Sprintf("%x/%x/%x", h.Bytes[0:2], h.Bytes[2:4], h.Bytes[4:])
}

// TODO: make this more robust
func HashStrToHash(str string) Hash {
	parts := strings.Split(str, ":")
	return Hash{HashAlgoNameSHA256, []byte(parts[1])}
}

func hashStr(str string) Hash {
	var s = sha256.Sum256([]byte("foo"))
	return Hash{AlgoName: HashAlgoNameSHA256, Bytes: s[:]}
}
