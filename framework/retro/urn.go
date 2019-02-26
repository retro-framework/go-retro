package retro

// PartitionName is a typealias for string
// to help keep the internals of the library
// clear about whether we're dealing with
// any old string or specifically a partition
// name
type URN string

func (urn URN) TrimPrefix() string {
	if len(urn) == 0 {
		return ""
	}
	return string(urn[4:])
}

func (urn URN) String() string {
	return string(urn)
}
