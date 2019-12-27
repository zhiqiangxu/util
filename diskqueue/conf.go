package diskqueue

import (
	"context"
	"sync"
	"time"
)

// CustomDecoder for customized packets
type CustomDecoder func(context.Context, *QfileSizeReader) ([]byte, error)

// Conf for diskqueue
type Conf struct {
	Directory         string
	WriteBatch        int
	WriteMmap         bool
	MaxMsgSize        int
	CustomDecoder     CustomDecoder
	MaxPutting        int
	EnableWriteBuffer bool
	MaxFileSize       int64
	PersistDuration   time.Duration // GC works at qfile granularity
	// below only valid when EnableWriteBuffer is true
	// unit: second
	CommitInterval  int
	WriteBufferPool *sync.Pool
	// below are modified internally for cache
	writeBufferPool *sync.Pool
	customDecoder   bool
}
