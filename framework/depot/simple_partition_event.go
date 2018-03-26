package depot

import "time"

type SimplePartitionEvent struct{}

func (s *SimplePartitionEvent) Time() time.Time {
	return time.Time{}
}

func (s *SimplePartitionEvent) Name() string {
	return "demo"
}

func (s *SimplePartitionEvent) Bytes() []byte {
	return []byte{}
}
