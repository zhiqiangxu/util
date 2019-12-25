package diskqueue

import "time"

// Conf for diskqueue
type Conf struct {
	Directory         string
	WriteBatch        int
	WriteMmap         bool
	MaxMsgSize        int
	MaxPutting        int
	EnableWriteBuffer bool
	PersistDuration   time.Duration
	// only valid when EnableWriteBuffer is true
	// unit: second
	CommitInterval int
}
