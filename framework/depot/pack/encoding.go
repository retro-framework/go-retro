package pack

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"sort"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/types"
)

func NewJSONPacker() *JSONPacker {
	return &JSONPacker{func() hash.Hash { return sha256.New() }}
}

// JSONPacker packs events, affixes and checkpoints
// as payloads including hashing. The packer includes
// a mutex because the hash engine is shared and must
// be reset before use to clear buffers.
type JSONPacker struct {
	hFactory func() hash.Hash
}

// PackEvent packs an Event type object into an envelope
// and returns it with a hashed payload attached.
// An event is packed with a header and a hint about it's
// type and encoding, followed by the raw bytes, terminated
// with a null byte.
//
// See the tests for an example of how the on-disk format
// looks.
func (jp *JSONPacker) PackEvent(evName string, ev types.Event) (HashedObject, error) {

	var payload bytes.Buffer

	evB, err := json.Marshal(ev)
	if err != nil {
		return nil, errors.WithMessage(err, "retro-json-pack: can't marshal ev as json")
	}

	payload.WriteString(fmt.Sprintf("event json %s %d", evName, len(evB)))
	payload.WriteString(HeaderContentSepRune)
	payload.Write(evB)

	hash := jp.hFactory()
	hash.Write(payload.Bytes())

	// Assemble PackedEvent
	// rPayload.Rewind()
	return &PackedEvent{
		PackedObject{
			hash:    Hash{HashAlgoNameSHA256, hash.Sum(nil)},
			payload: payload.Bytes(),
		}}, nil

}

// PackAffix packs an affix by rendering a text-table
// and injecting the prefix header, etc
func (jp *JSONPacker) PackAffix(affix Affix) (HashedObject, error) {

	var (
		affB       bytes.Buffer
		payload    bytes.Buffer
		partitions []PartitionName
	)

	// Write the affix text representation in lexographical
	// order, probably. Go 1.0+ _intentionally_ randomizes has
	// iteration order.
	for key, _ := range affix {
		partitions = append(partitions, key)
	}
	sort.SliceStable(partitions, func(i, j int) bool {
		return partitions[i] < partitions[j]
	})

	for i, partition := range partitions {
		prefix := fmt.Sprintf("%d %s", i, partition)
		for _, h := range affix[partition] {
			affB.WriteString(fmt.Sprintf("%s %s:%x\n", prefix, h.AlgoName, h.Bytes))
		}
	}

	payload.WriteString(fmt.Sprintf("affix %d", len(affB.Bytes())))
	payload.WriteString(HeaderContentSepRune)
	payload.Write(affB.Bytes())

	hash := jp.hFactory()
	hash.Write(payload.Bytes())

	return &PackedAffix{
		PackedObject{
			hash:    Hash{HashAlgoNameSHA256, hash.Sum(nil)},
			payload: payload.Bytes(),
		}}, nil
}

// func (jp *JSONPacker) PackAffix(ev Affix) (HashedObject, error) {
//
// }
