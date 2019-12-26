package diskqueue

import (
	"sync"
	"time"
)

// Conf for diskqueue
type Conf struct {
	Directory         string
	WriteBatch        int
	WriteMmap         bool
	MaxMsgSize        int
	MaxPutting        int
	EnableWriteBuffer bool
	MaxFileSize       int64
	PersistDuration   time.Duration // GC works at qfile granularity
	// below only valid when EnableWriteBuffer is true
	// unit: second
	CommitInterval  int
	WriteBufferPool *sync.Pool
	writeBufferPool *sync.Pool
}
