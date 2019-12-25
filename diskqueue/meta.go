package diskqueue

import (
	"encoding/binary"
	"sync"
	"unsafe"

	"os"
	"path/filepath"

	"github.com/zhiqiangxu/util"
	"github.com/zhiqiangxu/util/logger"
	"github.com/zhiqiangxu/util/mapped"
	"go.uber.org/zap"
)

type queueMetaROInterface interface {
	NumFiles() int
	Stat() QueueMeta
	FileMeta(idx int) FileMeta
}

type queueMetaInterface interface {
	queueMetaROInterface
	Init() error
	AddFile(f FileMeta)
	UpdateFileStat(idx, n int, endOffset, endTime int64)
	LocateFile(readOffset int64) int
	Sync() error
	Close() error
}

var _ queueMetaInterface = (*queueMeta)(nil)

const (
	maxSizeForMeta = 1024 * 1024
)

var (
	reservedHeaderSize = util.Max(256 /*should be enough*/, int(unsafe.Sizeof(QueueMeta{})))
)

// FileMeta for a single file
type FileMeta struct {
	StartOffset int64 // inclusive
	EndOffset   int64 // exclusive
	StartTime   int64
	EndTime     int64
	MsgCount    uint64
}

// QueueMeta for the queue
type QueueMeta struct {
	FileCount     uint32
	MinValidIndex uint32
}

type queueMeta struct {
	mu          sync.RWMutex
	conf        *Conf
	mappedFile  *mapped.File
	mappedBytes []byte
}

// NewQueueMeta is ctor for queueMeta
func newQueueMeta(conf *Conf) *queueMeta {
	return &queueMeta{conf: conf}
}

const (
	metaFile = "qm"
)

// Init either load or creates the meta file
func (m *queueMeta) Init() (err error) {
	path := filepath.Join(m.conf.Directory, metaFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		m.mappedFile, err = mapped.CreateFile(path, maxSizeForMeta, true, nil)
	} else {
		m.mappedFile, err = mapped.OpenFile(path, maxSizeForMeta, os.O_RDWR, true, nil)
	}
	if err != nil {
		return
	}

	err = m.mappedFile.MLock()
	if err != nil {
		return
	}
	m.mappedBytes = m.mappedFile.MappedBytes()

	return nil
}

func (m *queueMeta) NumFiles() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return int(binary.BigEndian.Uint32(m.mappedBytes))
}

func (m *queueMeta) Stat() QueueMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return QueueMeta{FileCount: binary.BigEndian.Uint32(m.mappedBytes), MinValidIndex: binary.BigEndian.Uint32(m.mappedBytes[4:])}
}

func (m *queueMeta) FileMeta(idx int) (fm FileMeta) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nFiles := int(binary.BigEndian.Uint32(m.mappedBytes))
	if idx >= nFiles {
		logger.Instance().Fatal("FileMeta idx over size", zap.Int("idx", idx), zap.Int("nFiles", nFiles))
	}

	offset := reservedHeaderSize + int(unsafe.Sizeof(FileMeta{}))*idx
	startOffset := int64(binary.BigEndian.Uint64(m.mappedBytes[offset:]))
	endOffset := int64(binary.BigEndian.Uint64(m.mappedBytes[offset+8:]))
	startTime := int64(binary.BigEndian.Uint64(m.mappedBytes[offset+16:]))
	endTime := int64(binary.BigEndian.Uint64(m.mappedBytes[offset+24:]))
	msgCount := binary.BigEndian.Uint64(m.mappedBytes[offset+32:])

	fm = FileMeta{StartOffset: startOffset, EndOffset: endOffset, StartTime: startTime, EndTime: endTime, MsgCount: msgCount}
	return

}

func (m *queueMeta) AddFile(f FileMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nFiles := binary.BigEndian.Uint32(m.mappedBytes)
	binary.BigEndian.PutUint32(m.mappedBytes, nFiles+1)
	offset := reservedHeaderSize + int(unsafe.Sizeof(FileMeta{}))*int(nFiles)
	binary.BigEndian.PutUint64(m.mappedBytes[offset:], uint64(f.StartOffset))
	binary.BigEndian.PutUint64(m.mappedBytes[offset+8:], uint64(f.EndOffset))
	binary.BigEndian.PutUint64(m.mappedBytes[offset+16:], uint64(f.StartTime))
	binary.BigEndian.PutUint64(m.mappedBytes[offset+24:], uint64(f.EndTime))

}

func (m *queueMeta) UpdateFileStat(idx, n int, endOffset, endTime int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nFiles := int(binary.BigEndian.Uint32(m.mappedBytes))
	if idx >= nFiles {
		logger.Instance().Fatal("UpdateFileStat idx over size", zap.Int("idx", idx), zap.Int("nFiles", nFiles))
	}

	offset := reservedHeaderSize + int(unsafe.Sizeof(FileMeta{}))*idx
	endOffset0 := int64(binary.BigEndian.Uint64(m.mappedBytes[offset+8:]))
	startTime0 := int64(binary.BigEndian.Uint64(m.mappedBytes[offset+16:]))
	endTime0 := int64(binary.BigEndian.Uint64(m.mappedBytes[offset+24:]))
	msgCount0 := binary.BigEndian.Uint64(m.mappedBytes[offset+32:])

	if endOffset > endOffset0 {
		binary.BigEndian.PutUint64(m.mappedBytes[offset+8:], uint64(endOffset))
	}

	if endTime < startTime0 {
		binary.BigEndian.PutUint64(m.mappedBytes[offset+16:], uint64(endTime))
	}

	if endTime > endTime0 {
		binary.BigEndian.PutUint64(m.mappedBytes[offset+24:], uint64(endTime))
	}

	binary.BigEndian.PutUint64(m.mappedBytes[offset+32:], msgCount0+uint64(n))

	return

}

func (m *queueMeta) LocateFile(readOffset int64) int {

	m.mu.RLock()
	defer m.mu.RUnlock()

	start := 0
	end := int(binary.BigEndian.Uint32(m.mappedBytes) - 1)

	// 二分查找
	for start <= end {
		target := start + (end-start)/2 // 防止溢出
		offset := reservedHeaderSize + int(unsafe.Sizeof(FileMeta{}))*target
		startOffset := int64(binary.BigEndian.Uint64(m.mappedBytes[offset:]))
		endOffset := int64(binary.BigEndian.Uint64(m.mappedBytes[offset+8:]))
		switch {
		case startOffset <= readOffset && endOffset > readOffset:
			return target
		case readOffset < startOffset:
			end = target - 1
		case readOffset >= endOffset:
			start = target + 1
		}
	}

	return -1

}

func (m *queueMeta) Sync() error {
	return m.mappedFile.Sync()
}

func (m *queueMeta) Close() error {
	m.mappedBytes = nil
	return m.mappedFile.Close()
}
