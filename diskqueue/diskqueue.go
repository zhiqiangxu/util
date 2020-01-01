package diskqueue

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"sync/atomic"

	"github.com/zhiqiangxu/util"
	"github.com/zhiqiangxu/util/closer"
	"github.com/zhiqiangxu/util/logger"
	"github.com/zhiqiangxu/util/mapped"
	"github.com/zhiqiangxu/util/wm"
	"go.uber.org/zap"
)

type queueInterface interface {
	queueMetaROInterface
	Put([]byte) (int64, error)
	Read(ctx context.Context, offset int64) ([]byte, error)
	StreamRead(ctx context.Context, offset int64) (<-chan []byte, error)
	StreamOffsetRead(offsetCh <-chan int64) (<-chan []byte, error)
	Close()
	GC() (int, error)
	Delete() error
}

var _ queueInterface = (*Queue)(nil)

// Queue for diskqueue
type Queue struct {
	putting    int32
	gcFlag     uint32
	closeState uint32
	closer     *closer.Signal
	meta       *queueMeta
	conf       Conf
	writeCh    chan *writeRequest
	writeReqs  []*writeRequest
	writeBuffs net.Buffers
	sizeBuffs  []byte
	gcCh       chan *gcRequest
	// guards files,minValidIndex
	flock         sync.RWMutex
	files         []*qfile
	minValidIndex int
	once          sync.Once
	wm            *wm.Offset // maintains commit offset
}

const (
	defaultWriteBatch      = 100
	defaultMaxMsgSize      = 512 * 1024 * 1024
	defaultMaxPutting      = 200000
	defaultPersistDuration = 3 * 24 * time.Hour
	sizeLength             = 4
)

// New is ctor for Queue
func New(conf Conf) (q *Queue, err error) {
	if conf.Directory == "" {
		err = errEmptyDirectory
		return
	}
	if conf.PersistDuration < time.Hour {
		conf.PersistDuration = defaultPersistDuration
	}
	if conf.MaxFileSize <= 0 {
		conf.MaxFileSize = qfileDefaultSize
	}

	if conf.WriteBatch <= 0 {
		conf.WriteBatch = defaultWriteBatch
	}
	if conf.MaxMsgSize <= 0 {
		conf.MaxMsgSize = defaultMaxMsgSize
	}
	if conf.MaxPutting <= 0 {
		conf.MaxPutting = defaultMaxPutting
	}
	if conf.CustomDecoder != nil {
		conf.customDecoder = true
	}

	q = &Queue{
		closer:    closer.NewSignal(),
		conf:      conf,
		writeCh:   make(chan *writeRequest, conf.WriteBatch),
		writeReqs: make([]*writeRequest, 0, conf.WriteBatch),
		gcCh:      make(chan *gcRequest),
		wm:        wm.NewOffset(),
	}
	if conf.customDecoder {
		q.writeBuffs = make(net.Buffers, 0, conf.WriteBatch)
		// q.sizeBuffs = nil
	} else {
		q.writeBuffs = make(net.Buffers, 0, conf.WriteBatch*2)
		q.sizeBuffs = make([]byte, sizeLength*conf.WriteBatch)
	}
	q.meta = newQueueMeta(&q.conf)
	err = q.init()
	return
}

const (
	dirPerm = 0770
)

// NumFiles is proxy for meta
func (q *Queue) NumFiles() int {
	return q.meta.NumFiles()
}

func (q *Queue) writeBufferPool() *sync.Pool {
	if !q.conf.EnableWriteBuffer {
		return nil
	}
	if q.conf.WriteBufferPool != nil {
		return q.conf.WriteBufferPool
	} else if q.conf.writeBufferPool != nil {
		return q.conf.writeBufferPool
	} else {
		pool := &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, q.conf.MaxFileSize))
			},
		}
		q.conf.writeBufferPool = pool
		return pool
	}
}

// Stat is proxy for meta
func (q *Queue) Stat() QueueMeta {
	return q.meta.Stat()
}

// FileMeta is proxy for meta
func (q *Queue) FileMeta(idx int) FileMeta {
	return q.meta.FileMeta(idx)
}

// init the queue
func (q *Queue) init() (err error) {

	// 确保各种目录存在
	err = os.MkdirAll(filepath.Join(q.conf.Directory, qfSubDir), dirPerm)
	if err != nil {
		return
	}

	// 初始化元数据
	err = q.meta.Init()
	if err != nil {
		return
	}

	// 加载qfile
	stat := q.Stat()
	nFiles := int(stat.FileCount)
	q.minValidIndex = int(stat.MinValidIndex)
	q.files = make([]*qfile, 0, nFiles-q.minValidIndex)
	var qf *qfile
	for i := q.minValidIndex; i < nFiles; i++ {
		qf, err = openQfile(q, i, i == nFiles-1)
		if err != nil {
			return
		}
		if i < (nFiles - 1) {
			err = qf.Shrink()
			if err != nil {
				return
			}
		}
		q.files = append(q.files, qf)
	}

	// enough data, ready to go!

	if len(q.files) == 0 {
		err = q.createQfile()
		if err != nil {
			logger.Instance().Error("Init createQfile", zap.Error(err))
			return
		}
	}

	util.GoFunc(q.closer.WaitGroupRef(), q.handleWriteAndGC)
	util.GoFunc(q.closer.WaitGroupRef(), q.handleCommit)

	return nil
}

func (q *Queue) maxValidIndex() int {
	return q.minValidIndex + len(q.files) - 1
}

func (q *Queue) nextIndex() int {
	return q.minValidIndex + len(q.files)
}

func (q *Queue) createQfile() (err error) {
	var qf *qfile
	if len(q.files) == 0 {
		qf, err = createQfile(q, 0, 0)
		if err != nil {
			return
		}
	} else {
		qf = q.files[len(q.files)-1]
		commitOffset := qf.DoneWrite()
		q.wm.Done(commitOffset)
		qf, err = createQfile(q, q.nextIndex(), qf.WrotePosition())
		if err != nil {
			return
		}
	}
	q.flock.Lock()
	q.files = append(q.files, qf)
	q.flock.Unlock()
	return
}

type writeResult struct {
	offset int64
}

type writeRequest struct {
	data   []byte
	result chan writeResult
}

type gcResult struct {
	n   int
	err error
}

type gcRequest struct {
	result chan gcResult
}

var wreqPool = sync.Pool{New: func() interface{} {
	return &writeRequest{result: make(chan writeResult, 1)}
}}

// dedicated G so that write is serial
func (q *Queue) handleWriteAndGC() {
	var (
		wReq           *writeRequest
		gcReq          *gcRequest
		qf             *qfile
		err            error
		wroteN, totalN int64
		gcN            int
	)

	startFM := q.meta.FileMeta(q.maxValidIndex())
	startWrotePosition := startFM.EndOffset

	var (
		updateWriteBufsFunc func(i int, data []byte)
		actualSizeLength    int64
	)
	if q.conf.customDecoder {
		updateWriteBufsFunc = func(i int, data []byte) {
			q.writeBuffs = append(q.writeBuffs, data)
		}
	} else {
		updateWriteBufsFunc = func(i int, data []byte) {
			q.updateSizeBuf(i, len(data))
			q.writeBuffs = append(q.writeBuffs, q.getSizeBuf(i))
			q.writeBuffs = append(q.writeBuffs, wReq.data)
		}
		actualSizeLength = sizeLength
	}

	handleWriteFunc := func() {
		// enough data, ready to go!
		qf = q.files[len(q.files)-1]

		writeBuffs := q.writeBuffs

		util.TryUntilSuccess(func() bool {
			wroteN, err = qf.writeBuffers(&q.writeBuffs)
			totalN += wroteN
			if err == mapped.ErrWriteBeyond {
				// 写超了，需要新开文件
				err = q.createQfile()
				if err != nil {
					logger.Instance().Error("handleWriteAndGC createQfile", zap.Error(err))
				} else {
					qf = q.files[len(q.files)-1]
					wroteN, err = qf.writeBuffers(&q.writeBuffs)
					totalN += wroteN
				}
			}
			if err != nil {
				logger.Instance().Error("handleWriteAndGC WriteTo", zap.Error(err))
				return false
			}
			return true
		}, time.Second)

		q.meta.UpdateFileStat(q.maxValidIndex(), len(q.writeReqs), startWrotePosition+totalN, NowNano())
		if !q.conf.EnableWriteBuffer {
			q.wm.Done(startWrotePosition + totalN)
		}

		q.writeBuffs = writeBuffs

		// 全部写入成功
		for _, req := range q.writeReqs {
			req.result <- writeResult{offset: startWrotePosition}
			startWrotePosition += actualSizeLength + int64(len(req.data))
		}
		totalN = 0
	}

	for {
		select {
		case <-q.closer.HasBeenClosed():
			// drain writeCh before quit
		DrainStart:
			q.writeReqs = q.writeReqs[:0]
			q.writeBuffs = q.writeBuffs[:0]
		DrainLoop:
			for i := 0; i < q.conf.WriteBatch; i++ {
				select {
				case wReq = <-q.writeCh:
					q.writeReqs = append(q.writeReqs, wReq)
					updateWriteBufsFunc(i, wReq.data)
				default:
					break DrainLoop
				}
			}

			if len(q.writeReqs) > 0 {
				handleWriteFunc()

				if len(q.writeReqs) == q.conf.WriteBatch {
					goto DrainStart
				}
			}

			close(q.writeCh)

			q.writeReqs = q.writeReqs[:0]
			q.writeBuffs = q.writeBuffs[:0]

			var ok bool
		DrainLoopFinal:
			for i := 0; i < q.conf.WriteBatch; i++ {
				select {
				case wReq, ok = <-q.writeCh:
					if !ok {
						break DrainLoopFinal
					}
					q.writeReqs = append(q.writeReqs, wReq)
					updateWriteBufsFunc(i, wReq.data)
				}
			}

			if len(q.writeReqs) > 0 {
				handleWriteFunc()
			}
			return
		case gcReq = <-q.gcCh:

			gcN, err = q.gc()
			gcReq.result <- gcResult{n: gcN, err: err}

		case wReq = <-q.writeCh:
			q.writeReqs = q.writeReqs[:0]
			q.writeBuffs = q.writeBuffs[:0]

			q.writeReqs = append(q.writeReqs, wReq)
			updateWriteBufsFunc(0, wReq.data)

			// collect more data
		BatchLoop:
			for i := 0; i < q.conf.WriteBatch-1; i++ {
				select {
				case wReq = <-q.writeCh:
					q.writeReqs = append(q.writeReqs, wReq)
					updateWriteBufsFunc(i+1, wReq.data)
				default:
					break BatchLoop
				}
			}

			handleWriteFunc()
		}
	}
}

func (q *Queue) getSizeBuf(i int) []byte {
	return q.sizeBuffs[sizeLength*i : sizeLength*i+sizeLength]
}

func (q *Queue) updateSizeBuf(i int, size int) {
	binary.BigEndian.PutUint32(q.sizeBuffs[sizeLength*i:], uint32(size))
}

const (
	commitMinimumInterval = 1
)

func (q *Queue) handleCommit() {
	if !q.conf.EnableWriteBuffer {
		return
	}

	interval := commitMinimumInterval
	if q.conf.CommitInterval > commitMinimumInterval {
		interval = q.conf.CommitInterval
	}
	ticker := time.NewTicker(time.Second * time.Duration(interval))

	for {
		select {
		case <-ticker.C:
			q.flock.RLock()
			qf := q.files[len(q.files)-1]
			q.flock.RUnlock()
			commitOffset := qf.Commit()
			q.wm.Done(commitOffset)
		case <-q.closer.HasBeenClosed():
			return
		}
	}
}

// Put data to queue
func (q *Queue) Put(data []byte) (offset int64, err error) {

	if !q.conf.customDecoder && len(data) > q.conf.MaxMsgSize {
		err = errMsgTooLarge
		return
	}

	err = q.checkCloseState()
	if err != nil {
		return
	}

	putting := atomic.AddInt32(&q.putting, 1)
	defer atomic.AddInt32(&q.putting, -1)
	if int(putting) > q.conf.MaxPutting {
		err = errMaxPutting
		return
	}

	wreq := wreqPool.Get().(*writeRequest)
	wreq.data = data
	if len(wreq.result) > 0 {
		<-wreq.result
	}

	select {
	case q.writeCh <- wreq:
		result := <-wreq.result
		wreq.data = nil
		wreqPool.Put(wreq)
		offset = result.offset
		return
	case <-q.closer.HasBeenClosed():
		err = errAlreadyClosed
		return
	}

}

func (q *Queue) qfByIdx(idx int) *qfile {
	fileIndex := idx - q.minValidIndex
	if fileIndex < 0 {
		return nil
	}
	return q.files[fileIndex]
}

// ReadFrom for read from offset
func (q *Queue) Read(ctx context.Context, offset int64) (data []byte, err error) {
	err = q.checkCloseState()
	if err != nil {
		return
	}

	idx := q.meta.LocateFile(offset)
	if idx < 0 {
		err = errInvalidOffset
		return
	}

	q.flock.RLock()
	qf := q.qfByIdx(idx)
	if qf == nil {
		q.flock.RUnlock()
		err = errInvalidOffset
		return
	}

	rfc := qf.IncrRef()

	q.flock.RUnlock()

	// already deleted
	if rfc == 1 {
		err = errInvalidOffset
		return
	}

	defer qf.DecrRef()

	data, err = qf.Read(ctx, offset)

	return
}

// StreamRead for stream read
func (q *Queue) StreamRead(ctx context.Context, offset int64) (chRet <-chan []byte, err error) {
	err = q.checkCloseState()
	if err != nil {
		return
	}

	idx := q.meta.LocateFile(offset)
	if idx < 0 {
		err = errInvalidOffset
		return
	}

	q.flock.RLock()
	qf := q.qfByIdx(idx)
	if qf == nil {
		q.flock.RUnlock()
		err = errInvalidOffset
		return
	}

	rfc := qf.IncrRef()

	q.flock.RUnlock()

	// already deleted
	if rfc == 1 {
		err = errInvalidOffset
		return
	}

	defer qf.DecrRef()

	ch := make(chan []byte)
	chRet = ch
	util.GoFunc(q.closer.WaitGroupRef(), func() {
		// close the channel when done
		defer close(ch)

		streamCtx, streamCancel := context.WithCancel(ctx)
		defer streamCancel()

		var streamWG sync.WaitGroup
		util.GoFunc(&streamWG, func() {
			for {
				otherFile, _ := qf.StreamRead(streamCtx, offset, ch)
				if !otherFile {
					return
				}
				if idx < q.meta.NumFiles()-1 {
					offset = qf.WrotePosition()
					idx++
					qf.DecrRef()
					q.flock.RLock()
					qf = q.qfByIdx(idx)
					if qf == nil {
						logger.Instance().Fatal("qfByIdx nil", zap.Int("idx", idx))
					}
					rfc := qf.IncrRef()
					q.flock.RUnlock()
					if rfc == 1 {
						logger.Instance().Fatal("StreamRead rfc == 1", zap.Int("idx", idx))
					}

				} else {
					return
				}
			}
		})

		select {
		case <-q.closer.HasBeenClosed():
			streamCancel()
		case <-ctx.Done():
		}
		streamWG.Wait()

	})

	return
}

type streamOffsetResp struct {
	lastOffset int64
	otherFile  bool
	err        error
}

// StreamOffsetRead for continuous read by offset
// close offsetCh to signal the end of read
func (q *Queue) StreamOffsetRead(offsetCh <-chan int64) (chRet <-chan []byte, err error) {
	err = q.checkCloseState()
	if err != nil {
		return
	}

	ch := make(chan []byte)
	chRet = ch
	util.GoFunc(q.closer.WaitGroupRef(), func() {
		// close the channel when done
		defer close(ch)

		streamCtx, streamCancel := context.WithCancel(context.Background())
		defer streamCancel()

		var (
			streamWG sync.WaitGroup
			offset   int64
			ok       bool
			resp     *streamOffsetResp
		)

		respCh := make(chan *streamOffsetResp, 1)

		select {
		case offset, ok = <-offsetCh:
			if !ok {
				return
			}
			for {
				idx := q.meta.LocateFile(offset)
				if idx < 0 {
					return
				}
				q.flock.RLock()
				qf := q.qfByIdx(idx)
				if qf == nil {
					q.flock.RUnlock()
					return
				}

				rfc := qf.IncrRef()

				q.flock.RUnlock()

				// already deleted
				if rfc == 1 {
					return
				}

				util.GoFunc(&streamWG, func() {
					otherFile, lastOffset, err := qf.StreamOffsetRead(streamCtx, offset, offsetCh, ch)
					respCh <- &streamOffsetResp{otherFile: otherFile, lastOffset: lastOffset, err: err}
				})

				quit := false
				select {
				case <-q.closer.HasBeenClosed():
					streamCancel()
					streamWG.Wait()
					quit = true
				case resp = <-respCh:
					if resp.otherFile {
						offset = resp.lastOffset
					} else {
						quit = true
					}
				}

				qf.DecrRef()
				if quit {
					return
				}
			}

		case <-q.closer.HasBeenClosed():
			return
		}
	})

	return
}

var (
	errEmptyDirectory = errors.New("directory empty")
	errAlreadyClosed  = errors.New("already closed")
	errAlreadyClosing = errors.New("already closing")
	errMsgTooLarge    = errors.New("msg too large")
	errMaxPutting     = errors.New("too much putting")
	errInvalidOffset  = errors.New("invalid offset")
	errOffsetChClosed = errors.New("offsetCh closed")
)

const (
	open uint32 = iota
	closing
	closed
)

func (q *Queue) checkCloseState() (err error) {
	closeState := atomic.LoadUint32(&q.closeState)
	switch closeState {
	case open:
	case closing:
		err = errAlreadyClosing
	case closed:
		err = errAlreadyClosed
	default:
		err = fmt.Errorf("unknown close state:%d", closeState)
	}
	return
}

// Close the queue
func (q *Queue) Close() {

	q.once.Do(func() {
		atomic.StoreUint32(&q.closeState, closing)

		q.closer.SignalAndWait()

		util.TryUntilSuccess(func() bool {
			// try until success
			err := q.meta.Close()
			if err != nil {
				logger.Instance().Error("meta.Close", zap.Error(err))
				return false
			}

			return true
			// need human interfere

		}, time.Second)

		for _, file := range q.files {
			err := file.Close()
			if err != nil {
				logger.Instance().Error("file.Close", zap.Error(err))
			}
		}
		atomic.StoreUint32(&q.closeState, closed)
	})

	return
}

var (
	// ErrGCing when already gc
	ErrGCing = errors.New("already CGing")
)

// GC removes expired qfiles
func (q *Queue) GC() (n int, err error) {
	err = q.checkCloseState()
	if err != nil {
		return
	}

	swapped := atomic.CompareAndSwapUint32(&q.gcFlag, 0, 1)
	if !swapped {
		err = ErrGCing
		return
	}
	defer atomic.StoreUint32(&q.gcFlag, 0)

	gcReq := &gcRequest{result: make(chan gcResult, 1)}

	select {
	case q.gcCh <- gcReq:
		select {
		case <-q.closer.HasBeenClosed():
			err = errAlreadyClosed
			return
		case gcResult := <-gcReq.result:
			n, err = gcResult.n, gcResult.err
			return
		}
	case <-q.closer.HasBeenClosed():
		err = errAlreadyClosed
		return
	}
}

func (q *Queue) gc() (n int, err error) {
	stat := q.Stat()
	maxIdx := q.NumFiles() - 1
	idx := int(stat.MinValidIndex)

	for {
		if idx >= maxIdx {
			return
		}
		fileMeta := q.FileMeta(idx)

		fileEndTime := time.Unix(0, fileMeta.EndTime)
		if time.Now().Sub(fileEndTime) < q.conf.PersistDuration {
			return
		}

		// can GC

		q.flock.Lock()

		qf := q.qfByIdx(idx)
		q.minValidIndex = idx + 1
		q.files[0] = nil
		q.files = q.files[1:]

		q.flock.Unlock()

		q.meta.UpdateMinValidIndex(uint32(idx))
		qf.DecrRef()

		idx++
		n++

	}

}

// Delete the queue
func (q *Queue) Delete() error {
	q.Close()
	return os.RemoveAll(q.conf.Directory)
}
