package retro

import "strings"

// PartitionName is a typealias for string
// to help keep the internals of the library
// clear about whether we're dealing with
// any old string or specifically a partition
// name
type PartitionName string

func (pn PartitionName) Dirname() string {
	return strings.Split(string(pn), "/")[0]
}

func (pn PartitionName) ID() string {
	return strings.Split(string(pn), "/")[1]
}

func (pn PartitionName) String() string {
	return string(pn)
}
