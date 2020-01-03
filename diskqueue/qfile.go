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
	Read(ctx context.Context, offset int64) ([]byte, error)
	StreamRead(ctx context.Context, offset int64, ch chan<- StreamBytes) (bool, error)
	StreamOffsetRead(ctx context.Context, offset int64, offsetCh <-chan int64, ch chan<- StreamBytes) (bool, int64, error)
	Sync() error
	Close() error
}

// qfile has no write-write races, but has read-write races
type qfile struct {
	ref            int32
	q              *Queue
	idx            int
	startOffset    int64
	mappedFile     *mapped.File
	notLatest      bool
	readLockedFunc func(ctx context.Context, r *QfileSizeReader) (otherFile bool, startOffset int64, dataBytes []byte, err error)
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
	if err != nil {
		return
	}
	qf.init()

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
	qf.init()

	if q.meta.NumFiles() != idx {
		logger.Instance().Fatal("createQfile idx != NumFiles", zap.Int("NumFiles", q.meta.NumFiles()), zap.Int("idx", idx))
	}

	nowNano := NowNano()
	q.meta.AddFile(FileMeta{StartOffset: startOffset, EndOffset: startOffset, StartTime: nowNano, EndTime: nowNano})

	return
}

func (qf *qfile) init() {
	if qf.q.conf.customDecoder {
		qf.readLockedFunc = qf.readLockedCustom
	} else {
		qf.readLockedFunc = qf.readLockedDefault
	}
}

func (qf *qfile) IncrRef() int32 {
	return atomic.AddInt32(&qf.ref, 1)
}

func (qf *qfile) DecrRef() (newRef int32) {
	newRef = atomic.AddInt32(&qf.ref, -1)
	if newRef > 0 {
		return
	}

	err := qf.Close()
	if err != nil {
		logger.Instance().Error("qf.Close", zap.Error(err))
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

// isLatest can be called concurrently :)
func (qf *qfile) isLatest() bool {
	if qf.notLatest {
		return false
	}
	isLatest := qf.idx == qf.q.meta.NumFiles()-1
	if !isLatest {
		qf.notLatest = true
	}
	return isLatest
}

func (qf *qfile) readLockedCustom(ctx context.Context, r *QfileSizeReader) (otherFile bool, startOffset int64, dataBytes []byte, err error) {

	startOffset = r.NextOffset()
	otherFile, dataBytes, err = qf.q.conf.CustomDecoder(ctx, r)
	if err == mapped.ErrReadBeyond && !qf.isLatest() {
		otherFile = true
	}
	return
}

func (qf *qfile) readLockedDefault(ctx context.Context, r *QfileSizeReader) (otherFile bool, startOffset int64, dataBytes []byte, err error) {

	startOffset = r.NextOffset()
	var sizeBytes [sizeLength]byte
	err = r.Read(ctx, sizeBytes[:])
	if err != nil {
		if err == mapped.ErrReadBeyond && !qf.isLatest() {
			otherFile = true
		}
		return
	}

	size := int(binary.BigEndian.Uint32(sizeBytes[:]))
	if size > qf.q.conf.MaxMsgSize {
		err = errInvalidOffset
		return
	}

	dataBytes = make([]byte, size)
	err = r.Read(ctx, dataBytes)
	return
}

func (qf *qfile) calcFileOffset(offset int64) (fileOffset int64, err error) {
	fileOffset = offset - qf.startOffset
	if fileOffset < 0 {
		logger.Instance().Error("calcFileOffset negative fileOffset", zap.Int64("offset", offset), zap.Int64("startOffset", qf.startOffset))
		err = errInvalidOffset
		return
	}

	return
}

func (qf *qfile) Read(ctx context.Context, offset int64) (data []byte, err error) {
	fileOffset, err := qf.calcFileOffset(offset)
	if err != nil {
		return
	}

	qf.mappedFile.RLock()
	defer qf.mappedFile.RUnlock()

	r := qf.getSizeReader(fileOffset)
	_, _, data, err = qf.readLockedFunc(ctx, r)
	qf.putSizeReader(r)
	return
}

var sizeReaderPool = sync.Pool{
	New: func() interface{} {
		return &QfileSizeReader{}
	},
}

func (qf *qfile) getSizeReader(fileOffset int64) *QfileSizeReader {
	r := sizeReaderPool.Get().(*QfileSizeReader)
	r.qf = qf
	r.fileOffset = fileOffset
	r.isLatest = qf.isLatest()
	return r
}

func (qf *qfile) putSizeReader(r *QfileSizeReader) {
	r.qf = nil
	sizeReaderPool.Put(r)
}

// when StreamRead returns , err is guaranteed not nil
func (qf *qfile) StreamRead(ctx context.Context, offset int64, ch chan<- StreamBytes) (otherFile bool, err error) {
	fileOffset, err := qf.calcFileOffset(offset)
	if err != nil {
		logger.Instance().Fatal("calcFileOffset err", zap.Int64("offset", offset), zap.Int64("startOffset", qf.startOffset))
		return
	}

	qf.mappedFile.RLock()
	defer qf.mappedFile.RUnlock()

	r := qf.getSizeReader(fileOffset)
	defer qf.putSizeReader(r)

	var (
		dataBytes   []byte
		startOffset int64
	)

	for {

		otherFile, startOffset, dataBytes, err = qf.readLockedFunc(ctx, r)
		if err != nil {
			return
		}

		select {
		case ch <- StreamBytes{Bytes: dataBytes, Offset: startOffset}:
		case <-ctx.Done():
			err = ctx.Err()
			return
		}

	}

}

func (qf *qfile) StreamOffsetRead(ctx context.Context, offset int64, offsetCh <-chan int64, ch chan<- StreamBytes) (otherFile bool, lastOffset int64, err error) {
	fileOffset, err := qf.calcFileOffset(offset)
	if err != nil {
		logger.Instance().Fatal("calcFileOffset err", zap.Int64("offset", offset), zap.Int64("startOffset", qf.startOffset))
		return
	}

	qf.mappedFile.RLock()
	defer qf.mappedFile.RUnlock()

	r := qf.getSizeReader(fileOffset)
	defer func() {
		lastOffset = r.fileOffset + qf.startOffset
		qf.putSizeReader(r)
	}()

	var (
		dataBytes   []byte
		nextOffset  int64
		ok          bool
		startOffset int64
	)

	for {

		otherFile, startOffset, dataBytes, err = qf.readLockedFunc(ctx, r)
		if err != nil {
			return
		}

		select {
		case ch <- StreamBytes{Bytes: dataBytes, Offset: startOffset}:
		case <-ctx.Done():
			err = ctx.Err()
			return
		}

		select {
		case nextOffset, ok = <-offsetCh:
			if !ok {
				err = errOffsetChClosed
				return
			}
			r.fileOffset = nextOffset - qf.startOffset
			if r.fileOffset < 0 {
				err = errInvalidOffset
				otherFile = true
				return
			}
		case <-ctx.Done():
			err = ctx.Err()
			return
		}

	}
}

func (qf *qfile) Shrink() error {
	return qf.mappedFile.Shrink()
}

func (qf *qfile) Sync() error {
	return qf.mappedFile.Sync()
}

func (qf *qfile) Close() error {
	return qf.mappedFile.Close()
}

func (qf *qfile) remove() (err error) {
	err = qf.mappedFile.Remove()
	return
}
