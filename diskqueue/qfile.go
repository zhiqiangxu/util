package diskqueue

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/zhiqiangxu/util/logger"
	"github.com/zhiqiangxu/util/mapped"
	"go.uber.org/zap"
)

type refCountInterface interface {
	IncrRef() int32
	DecrRef() int32
}
type qfileInterface interface {
	refCountInterface
	Shrink() error
	writeBuffers(buffs *net.Buffers) (int64, error)
	WrotePosition() int64
	DoneWrite() int64
	Commit() int64
	Read(offset int64) ([]byte, error)
	StreamRead(ctx context.Context, offset int64, ch chan []byte) error
	Sync() error
}

// qfile has no write-write races, but has read-write races
type qfile struct {
	ref         int32
	q           *Queue
	idx         int
	startOffset int64
	mappedFile  *mapped.File
}

const (
	qfSubDir         = "qf"
	qfileDefaultSize = 1024 * 1024 * 1024
)

var _ qfileInterface = (*qfile)(nil)

func qfilePath(startOffset int64, conf *Conf) string {
	return filepath.Join(conf.Directory, qfSubDir, fmt.Sprintf("%020d", startOffset))
}

func openQfile(q *Queue, idx int, isLatest bool) (qf *qfile, err error) {
	fm := q.meta.FileMeta(idx)

	qf = &qfile{q: q, idx: idx, startOffset: fm.StartOffset, ref: 1}
	var pool *sync.Pool
	if isLatest {
		pool = q.writeBufferPool()
	}
	qf.mappedFile, err = mapped.OpenFile(qfilePath(fm.StartOffset, &q.conf), int64(fm.EndOffset-fm.StartOffset), os.O_RDWR, q.conf.WriteMmap, pool)
	return
}

func createQfile(q *Queue, idx int, startOffset int64) (qf *qfile, err error) {
	qf = &qfile{q: q, idx: idx, startOffset: startOffset, ref: 1}
	var pool *sync.Pool
	if q.conf.EnableWriteBuffer {
		pool = q.writeBufferPool()
	}
	qf.mappedFile, err = mapped.CreateFile(qfilePath(startOffset, &q.conf), q.conf.MaxFileSize, q.conf.WriteMmap, pool)
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

func (qf *qfile) IncrRef() int32 {
	return atomic.AddInt32(&qf.ref, 1)
}

func (qf *qfile) DecrRef() (newRef int32) {
	newRef = atomic.AddInt32(&qf.ref, -1)
	if newRef > 0 {
		return
	}

	err := qf.close()
	if err != nil {
		logger.Instance().Error("qf.close", zap.Error(err))
	}

	err = qf.remove()
	if err != nil {
		logger.Instance().Error("qf.remove", zap.Error(err))
	}

	return
}

func (qf *qfile) writeBuffers(buffs *net.Buffers) (n int64, err error) {
	n, err = qf.mappedFile.WriteBuffers(buffs)
	return
	// n, err = buffs.WriteTo(qf.mappedFile)
	// return
}

func (qf *qfile) WrotePosition() int64 {
	return qf.startOffset + qf.mappedFile.GetWrotePosition()
}

func (qf *qfile) DoneWrite() int64 {
	return qf.startOffset + qf.mappedFile.DoneWrite()
}

func (qf *qfile) Commit() int64 {
	return qf.startOffset + qf.mappedFile.Commit()
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

	sizeBytes := make([]byte, sizeLength)
	_, err = qf.mappedFile.ReadRLocked(fileOffset, sizeBytes)
	if err != nil {
		return
	}

	size := int(binary.BigEndian.Uint32(sizeBytes))
	if size > qf.q.conf.MaxMsgSize {
		err = errInvalidOffset
		return
	}
	dataBytes = make([]byte, size)
	_, err = qf.mappedFile.ReadRLocked(fileOffset+sizeLength, dataBytes)
	return
}

// when StreamRead returns , err is guaranteed not nil
func (qf *qfile) StreamRead(ctx context.Context, offset int64, ch chan []byte) (err error) {
	fileOffset := offset - qf.startOffset
	if fileOffset < 0 {
		logger.Instance().Error("StreamRead negative fileOffset", zap.Int64("offset", offset), zap.Int64("startOffset", qf.startOffset))
		err = errInvalidOffset
		return
	}

	qf.mappedFile.RLock()
	defer qf.mappedFile.RUnlock()

	isLatest := qf.idx == qf.q.meta.NumFiles()-1

	for {
		sizeBytes := make([]byte, sizeLength)
		_, err = qf.mappedFile.ReadRLocked(fileOffset, sizeBytes)
		if err != nil {
			if !isLatest {
				return
			}
			err = qf.q.wm.Wait(ctx, qf.startOffset+fileOffset+sizeLength)
			if err != nil {
				return
			}
			_, err = qf.mappedFile.ReadRLocked(fileOffset, sizeBytes)
			if err != nil {
				// 说明换文件了
				return
			}
		}

		size := int(binary.BigEndian.Uint32(sizeBytes))
		if size > qf.q.conf.MaxMsgSize {
			err = errInvalidOffset
			return
		}
		dataBytes := make([]byte, size)
		_, err = qf.mappedFile.ReadRLocked(fileOffset+sizeLength, dataBytes)
		if err != nil {
			if !isLatest {
				return
			}
			err = qf.q.wm.Wait(ctx, qf.startOffset+fileOffset+sizeLength+int64(size))
			if err != nil {
				return
			}
			_, err = qf.mappedFile.ReadRLocked(fileOffset, sizeBytes)
			if err != nil {
				// 说明换文件了
				return
			}
		}

		select {
		case ch <- dataBytes:
		case <-ctx.Done():
		}
		fileOffset += sizeLength + int64(size)

	}

}

func (qf *qfile) Shrink() error {
	return qf.mappedFile.Shrink()
}

func (qf *qfile) Sync() error {
	return qf.mappedFile.Sync()
}

func (qf *qfile) close() error {
	return qf.mappedFile.Close()
}

func (qf *qfile) remove() (err error) {
	err = qf.mappedFile.Remove()
	return
}
