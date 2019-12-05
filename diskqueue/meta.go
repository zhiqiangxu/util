package diskqueue

import (
	"encoding/binary"
	"fmt"
	"sync"
	"unsafe"

	"os"

	"github.com/zhiqiangxu/util/mapped"
)

type queueMetaInterface interface {
	LoadOrCreate() error
	NumFiles() uint32
	FileMeta(idx int) FileMeta
	AddFile(f FileMeta)
	UpdateFileTime(idx int, startTime, endTime uint64) error
	UpdateFileEndOffset(idx int, endOffset uint64) error
	Sync() error
	Close() error
}

const (
	maxSizeForMeta = 1024 * 1024
)

// FileMeta for a single file
type FileMeta struct {
	StartOffset uint64
	EndOffset   uint64
	StartTime   uint64
	EndTime     uint64
}

type queueMeta struct {
	mu          sync.RWMutex
	path        string
	mappedFile  *mapped.File
	mappedBytes []byte
}

// NewQueueMeta is ctor for queueMeta
func newQueueMeta(path string) *queueMeta {
	return &queueMeta{path: path}
}

func (m *queueMeta) LoadOrCreate() (err error) {
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		m.mappedFile, err = mapped.CreateFile(m.path, maxSizeForMeta, true, nil)
	} else {
		m.mappedFile, err = mapped.OpenFile(m.path, maxSizeForMeta, os.O_RDWR, true, nil)
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

func (m *queueMeta) NumFiles() uint32 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return binary.BigEndian.Uint32(m.mappedBytes)
}

func (m *queueMeta) FileMeta(idx int) FileMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	offset := 4 + int(unsafe.Sizeof(FileMeta{}))*idx
	startOffset := binary.BigEndian.Uint64(m.mappedBytes[offset:])
	endOffset := binary.BigEndian.Uint64(m.mappedBytes[offset+8:])
	startTime := binary.BigEndian.Uint64(m.mappedBytes[offset+16:])
	endTime := binary.BigEndian.Uint64(m.mappedBytes[offset+24:])

	return FileMeta{StartOffset: startOffset, EndOffset: endOffset, StartTime: startTime, EndTime: endTime}
}

func (m *queueMeta) AddFile(f FileMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nFiles := binary.BigEndian.Uint32(m.mappedBytes)
	binary.BigEndian.PutUint32(m.mappedBytes, nFiles+1)
	offset := 4 + int(unsafe.Sizeof(FileMeta{}))*int(nFiles)
	binary.BigEndian.PutUint64(m.mappedBytes[offset:], f.StartOffset)
	binary.BigEndian.PutUint64(m.mappedBytes[offset+8:], f.EndOffset)
	binary.BigEndian.PutUint64(m.mappedBytes[offset+16:], f.StartTime)
	binary.BigEndian.PutUint64(m.mappedBytes[offset+24:], f.EndTime)
}

func (m *queueMeta) UpdateFileTime(idx int, startTime, endTime uint64) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nFiles := binary.BigEndian.Uint32(m.mappedBytes)
	if uint32(idx) >= nFiles {
		err = fmt.Errorf("UpdateFileTime beyond idx:%d nFiles:%d", idx, nFiles)
		return
	}

	offset := 4 + int(unsafe.Sizeof(FileMeta{}))*idx
	os := binary.BigEndian.Uint64(m.mappedBytes[offset+16:])
	oe := binary.BigEndian.Uint64(m.mappedBytes[offset+24:])

	if startTime < os {
		binary.BigEndian.PutUint64(m.mappedBytes[offset+16:], startTime)
	}

	if endTime > oe {
		binary.BigEndian.PutUint64(m.mappedBytes[offset+24:], endTime)
	}

	return

}

func (m *queueMeta) UpdateFileEndOffset(idx int, endOffset uint64) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nFiles := binary.BigEndian.Uint32(m.mappedBytes)
	if uint32(idx) >= nFiles {
		err = fmt.Errorf("UpdateFileEndOffset beyond idx:%d nFiles:%d", idx, nFiles)
		return
	}

	offset := 4 + int(unsafe.Sizeof(FileMeta{}))*idx
	oe := binary.BigEndian.Uint64(m.mappedBytes[offset+8:])

	if endOffset > oe {
		binary.BigEndian.PutUint64(m.mappedBytes[offset+8:], endOffset)
	}

	return
}

func (m *queueMeta) Sync() error {
	return m.mappedFile.Sync()
}

func (m *queueMeta) Close() error {
	m.mappedBytes = nil
	return m.mappedFile.Close()
}
