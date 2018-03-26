package packing

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/types"
)

func NewJSONPacker() *JSONPacker {
	return &JSONPacker{
		hashFn: func() hash.Hash { return sha256.New() },
		nowFn:  time.Now,
	}
}

// JSONPacker packs events, affixes and checkpoints as payloads including
// hashing. The packer includes a mutex because the hash engine is shared and
// must be reset before use to clear buffers.
type JSONPacker struct {
	hashFn func() hash.Hash
	nowFn  func() time.Time
}

// PackEvent packs an Event type object into an envelope and returns it with a
// hashed payload attached.  An event is packed with a header and a hint about
// it's type and encoding, followed by the raw bytes, terminated with a null
// byte.
//
// See the tests for an example of how the on-disk format looks.
func (jp *JSONPacker) PackEvent(evName string, ev types.Event) (HashedObject, error) {

	var payload bytes.Buffer

	evB, err := json.Marshal(ev)
	if err != nil {
		return nil, errors.WithMessage(err, "retro-json-pack: can't marshal ev as json")
	}

	payload.WriteString(fmt.Sprintf("%s json %s %d", ObjectTypeEvent, evName, len(evB)))
	payload.WriteString(HeaderContentSepRune)
	payload.Write(evB)

	hash := jp.hashFn()
	hash.Write(payload.Bytes())

	return &PackedEvent{
		PackedObject{
			hash:    Hash{HashAlgoNameSHA256, hash.Sum(nil)},
			payload: payload.Bytes(),
		}}, nil

}

// Unpack event takes a byte slice and returns an event name, and a payload
// which can be unmarahelled into that event tyype, because of limitations of
// the type safety system a event name and payload are the best we can do here.
// The caller should use the event name to request an zero value event with
// that name from the event registry and then decode it.
//
// TODO: ensure that the byte slice given actually contains an event (e.g look
// at the frontmatter)
func (jp *JSONPacker) UnpackEvent(b []byte) (string, []byte, error) {
	var (
		chunks      = bytes.SplitN(b, []byte(HeaderContentSepRune), 2)
		frontMatter = chunks[0]
		payload     = chunks[1]
		parts       = strings.SplitN(string(frontMatter), " ", 4)
	)
	return parts[2], payload, nil
}

// UnpackAffix returns an unpacked affix given a byte stream containing an affix
// TODO: ensure bytes given are actually an affix!
func (jp *JSONPacker) UnpackAffix(b []byte) (Affix, error) {
	var (
		res = Affix{}

		chunks  = bytes.SplitN(b, []byte(HeaderContentSepRune), 2)
		payload = chunks[1]
	)

	scanner := bufio.NewScanner(bytes.NewReader(payload))
	for scanner.Scan() {
		var (
			cols          = strings.SplitN(scanner.Text(), " ", 3)
			partitionName = PartitionName(cols[1])
			evHash        = HashStrToHash(cols[2])
		)
		res[partitionName] = append(res[partitionName], evHash)
	}
	if err := scanner.Err(); err != nil {
		// TODO: Wrap err properly
		return nil, err
	}

	return res, nil
}

func (jp *JSONPacker) UnpackCheckpoint(b []byte) (Checkpoint, error) {
	var (
		res = Checkpoint{Fields: make(map[string]string)}

		kvHeadersRead bool

		chunks  = bytes.SplitN(b, []byte(HeaderContentSepRune), 2)
		payload = chunks[1]
	)

	scanner := bufio.NewScanner(bytes.NewReader(payload))
	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			kvHeadersRead = true
			continue
		}

		if !kvHeadersRead {
			var (
				cols = strings.SplitN(scanner.Text(), " ", 2)
			)
			switch cols[0] {
			case "affix":
				res.AffixHash = HashStrToHash(cols[1])
			case "parent":
				res.ParentHashes = append(res.ParentHashes, HashStrToHash(cols[1]))
			default:
				res.Fields[cols[0]] = cols[1]
			}
			// fmt.Println(cols)
		} else {
			res.CommandDesc = []byte(scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return res, err // TODO: Wrap properly
	}

	return res, nil
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
	for key := range affix {
		partitions = append(partitions, key)
	}
	sort.SliceStable(partitions, func(i, j int) bool {
		return partitions[i] < partitions[j]
	})

	for i, partition := range partitions {
		prefix := fmt.Sprintf("%d %s", i, partition)
		for _, h := range affix[partition] {
			affB.WriteString(fmt.Sprintf("%s %s\n", prefix, h.String()))
		}
	}

	payload.WriteString(fmt.Sprintf("%s %d", ObjectTypeAffix, len(affB.Bytes())))
	payload.WriteString(HeaderContentSepRune)
	payload.Write(affB.Bytes())

	hash := jp.hashFn()
	hash.Write(payload.Bytes())

	return &PackedAffix{
		PackedObject{
			hash:    Hash{HashAlgoNameSHA256, hash.Sum(nil)},
			payload: payload.Bytes(),
		}}, nil
}

// PackCheckpoint packs a checkpoint by rendering an email
// or HTTP style set of headers and injecting the prefix header, etc
func (jp *JSONPacker) PackCheckpoint(cp Checkpoint) (HashedObject, error) {

	var (
		cpB     bytes.Buffer
		payload bytes.Buffer
	)

	cpB.WriteString(fmt.Sprintf("%s %s\n", ObjectTypeAffix, cp.AffixHash.String()))

	for _, parentHash := range cp.ParentHashes {
		cpB.WriteString(fmt.Sprintf("parent %s\n", parentHash.String()))
	}

	for key, value := range cp.Fields {
		cpB.WriteString(fmt.Sprintf("%s %s\n", key, value))
	}

	cpB.WriteString(fmt.Sprintf("\n%s\n", cp.CommandDesc))

	payload.WriteString(fmt.Sprintf("%s %d", ObjectTypeCheckpoint, len(cpB.Bytes())))
	payload.WriteString(HeaderContentSepRune)
	payload.Write(cpB.Bytes())

	hash := jp.hashFn()
	hash.Write(payload.Bytes())

	return &PackedCheckpoint{
		PackedObject{
			hash:    Hash{HashAlgoNameSHA256, hash.Sum(nil)},
			payload: payload.Bytes(),
		}}, nil

}
