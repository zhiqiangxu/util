package diskqueue

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/zhiqiangxu/util/logger"
	"github.com/zhiqiangxu/util/mapped"
	"go.uber.org/zap"
)

type qfileInterface interface {
	Shrink() error
	writeBuffers(buffs *net.Buffers) (int64, error)
	WrotePosition() int64
	ReturnWriteBuffer()
	Commit()
	Read(offset int64) ([]byte, error)
	StreamRead(offset int64) (ch chan []byte, err error)
	Sync() error
	Close() error
}

type qfile struct {
	q           *Queue
	idx         int
	startOffset int64
	mappedFile  *mapped.File
}

const (
	qfSubDir  = "qf"
	qfileSize = 1024 * 1024 * 1024
)

var _ qfileInterface = (*qfile)(nil)

func qfilePath(startOffset int64, conf *Conf) string {
	return filepath.Join(conf.Directory, qfSubDir, fmt.Sprintf("%020d", startOffset))
}

func openQfile(q *Queue, idx int, isLatest bool) (qf *qfile, err error) {
	fm := q.meta.FileMeta(idx)

	qf = &qfile{q: q, idx: idx, startOffset: fm.StartOffset}
	var pool *sync.Pool
	if isLatest && q.conf.EnableWriteBuffer {
		pool = &writerBufferPool
	}
	qf.mappedFile, err = mapped.OpenFile(qfilePath(fm.StartOffset, &q.conf), int64(fm.EndOffset-fm.StartOffset), os.O_RDWR, q.conf.WriteMmap, pool)
	return
}

var writerBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, qfileSize))
	},
}

func createQfile(q *Queue, idx int, startOffset int64) (qf *qfile, err error) {
	qf = &qfile{q: q, idx: idx, startOffset: startOffset}
	var pool *sync.Pool
	if q.conf.EnableWriteBuffer {
		pool = &writerBufferPool
	}
	qf.mappedFile, err = mapped.CreateFile(qfilePath(startOffset, &q.conf), qfileSize, q.conf.WriteMmap, pool)
	if err != nil {
		return
	}

	if q.meta.NumFiles() != idx {
		logger.Instance().Fatal("createQfile idx != NumFiles", zap.Int("NumFiles", q.meta.NumFiles()), zap.Int("idx", idx))
	}

	nowNano := NowNano()
	q.meta.AddFile(FileMeta{StartOffset: startOffset, EndOffset: startOffset, StartTime: nowNano, EndTime: nowNano})
	return
}

func (qf *qfile) writeBuffers(buffs *net.Buffers) (n int64, err error) {
	n, err = qf.mappedFile.WriteBuffers(buffs)
	return
	// n, err = buffs.WriteTo(qf.mappedFile)
	// return
}

func (qf *qfile) WrotePosition() int64 {
	return qf.mappedFile.GetWrotePosition()
}

func (qf *qfile) ReturnWriteBuffer() {
	qf.mappedFile.ReturnWriteBuffer()
}

func (qf *qfile) Commit() {
	qf.mappedFile.Commit()
}

func (qf *qfile) Read(offset int64) (dataBytes []byte, err error) {
	fileOffset := offset - qf.startOffset
	if fileOffset < 0 {
		logger.Instance().Error("Read negative fileOffset", zap.Int64("offset", offset), zap.Int64("startOffset", qf.startOffset))
		err = errInvalidOffset
		return
	}

	qf.mappedFile.RLock()
	defer qf.mappedFile.RUnlock()

	sizeBytes := qf.q.syncByteArena.AllocBytes(sizeLength)
	_, err = qf.mappedFile.ReadRLocked(fileOffset, sizeBytes)
	if err != nil {
		return
	}

	size := int(binary.BigEndian.Uint32(sizeBytes))
	if size > qf.q.conf.MaxMsgSize {
		err = errInvalidOffset
		return
	}
	dataBytes = qf.q.syncByteArena.AllocBytes(size)
	_, err = qf.mappedFile.ReadRLocked(fileOffset+sizeLength, dataBytes)
	return
}

func (qf *qfile) StreamRead(offset int64) (ch chan []byte, err error) {
	fileOffset := offset - qf.startOffset
	if fileOffset < 0 {
		logger.Instance().Error("StreamRead negative fileOffset", zap.Int64("offset", offset), zap.Int64("startOffset", qf.startOffset))
		err = errInvalidOffset
		return
	}

	qf.mappedFile.RLock()
	defer qf.mappedFile.RUnlock()

	sizeBytes := qf.q.syncByteArena.AllocBytes(sizeLength)
	_, err = qf.mappedFile.ReadRLocked(fileOffset, sizeBytes)
	if err != nil {
		return
	}

	size := int(binary.BigEndian.Uint32(sizeBytes))
	if size > qf.q.conf.MaxMsgSize {
		err = errInvalidOffset
		return
	}

	return
}

func (qf *qfile) Shrink() error {
	return qf.mappedFile.Shrink()
}

func (qf *qfile) Sync() error {
	return nil
}

func (qf *qfile) Close() error {
	return nil
}
