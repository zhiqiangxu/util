package diskqueue

import (
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
	"github.com/zhiqiangxu/util/logger"
	"github.com/zhiqiangxu/util/mapped"
	"go.uber.org/zap"
)

type queueInterface interface {
	Put([]byte) (int64, error)
	Read(offset int64) ([]byte, error)
	StreamRead(offset int64) (chan []byte, error)
	EndStream(ch chan []byte) error
	Close()
	Delete() error
}

var _ queueInterface = (*Queue)(nil)

// Queue for diskqueue
type Queue struct {
	putting    int32
	closeState uint32
	wg         sync.WaitGroup
	meta       *queueMeta
	conf       Conf
	writeCh    chan *writeRequest
	writeReqs  []*writeRequest
	writeBuffs net.Buffers
	sizeBuffs  []byte
	doneCh     chan struct{}
	flock      sync.RWMutex
	files      []*qfile
	once       sync.Once
}

const (
	defaultWriteBatch         = 100
	defaultMaxMsgSize         = 512 * 1024 * 1024
	defaultMaxPutting         = 10000
	defaultByteArenaChunkSize = 100 * 1024 * 1024
	sizeLength                = 4
)

// New is ctor for Queue
func New(conf Conf) (q *Queue, err error) {
	if conf.Directory == "" {
		err = errEmptyDirectory
		return
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
	if conf.ByteArenaChunkSize <= 0 {
		conf.ByteArenaChunkSize = defaultByteArenaChunkSize
	}

	q = &Queue{
		conf:       conf,
		writeCh:    make(chan *writeRequest, conf.WriteBatch),
		writeReqs:  make([]*writeRequest, 0, conf.WriteBatch),
		writeBuffs: make(net.Buffers, 0, conf.WriteBatch*2),
		sizeBuffs:  make([]byte, sizeLength*conf.WriteBatch),
		doneCh:     make(chan struct{}),
	}
	q.meta = newQueueMeta(&q.conf)
	err = q.init()
	return
}

const (
	dirPerm = 0770
)

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
	nFiles := q.meta.NumFiles()
	q.files = make([]*qfile, 0, nFiles)
	var qf *qfile
	for i := 0; i < nFiles; i++ {
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

	util.GoFunc(&q.wg, q.handleWrite)
	util.GoFunc(&q.wg, q.handleCommit)

	return nil
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
		qf.DoneWrite()
		qf, err = createQfile(q, len(q.files), qf.WrotePosition())
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

var wreqPool = sync.Pool{New: func() interface{} {
	return &writeRequest{result: make(chan writeResult, 1)}
}}

// dedicated G so that write is serial
func (q *Queue) handleWrite() {
	var (
		wreq           *writeRequest
		qf             *qfile
		err            error
		wroteN, totalN int64
	)

	startFM := q.meta.FileMeta(len(q.files) - 1)
	startWrotePosition := startFM.EndOffset

	for {
		select {
		case <-q.doneCh:
			return
		case wreq = <-q.writeCh:
			q.writeReqs = q.writeReqs[:0]
			q.writeBuffs = q.writeBuffs[:0]
			q.writeReqs = append(q.writeReqs, wreq)
			q.updateSizeBuf(0, len(wreq.data))
			q.writeBuffs = append(q.writeBuffs, q.getSizeBuf(0))
			q.writeBuffs = append(q.writeBuffs, wreq.data)

			// collect more data
		BatchLoop:
			for i := 0; i < q.conf.WriteBatch-1; i++ {
				select {
				case wreq = <-q.writeCh:
					q.writeReqs = append(q.writeReqs, wreq)
					q.updateSizeBuf(i+1, len(wreq.data))
					q.writeBuffs = append(q.writeBuffs, q.getSizeBuf(i+1))
					q.writeBuffs = append(q.writeBuffs, wreq.data)
				default:
					break BatchLoop
				}
			}

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
						logger.Instance().Error("handleWrite createQfile", zap.Error(err))
					} else {
						qf = q.files[len(q.files)-1]
						wroteN, err = qf.writeBuffers(&q.writeBuffs)
						totalN += wroteN
					}
				}
				if err != nil {
					logger.Instance().Error("handleWrite WriteTo", zap.Error(err))
					return false
				}
				return true
			}, time.Second)

			q.meta.UpdateFileStat(len(q.files)-1, len(q.writeReqs), startWrotePosition+totalN, NowNano())

			q.writeBuffs = writeBuffs

			// 全部写入成功
			for _, req := range q.writeReqs {
				req.result <- writeResult{offset: startWrotePosition}
				startWrotePosition += int64(4) + int64(len(req.data))
			}
			totalN = 0

		}
	}
}

func (q *Queue) getSizeBuf(i int) []byte {
	return q.sizeBuffs[4*i : 4*i+sizeLength]
}

func (q *Queue) updateSizeBuf(i int, size int) {
	binary.BigEndian.PutUint32(q.sizeBuffs[4*i:], uint32(size))
}

func (q *Queue) handleCommit() {
	if !q.conf.EnableWriteBuffer {
		return
	}

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			q.flock.RLock()
			qf := q.files[len(q.files)-1]
			q.flock.RUnlock()
			qf.Commit()
		case <-q.doneCh:
			return
		}
	}
}

// Put data to queue
func (q *Queue) Put(data []byte) (offset int64, err error) {

	err = q.checkCloseState()
	if err != nil {
		return
	}

	if len(data) > q.conf.MaxMsgSize {
		err = errMsgTooLarge
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
	case <-q.doneCh:
		err = errAlreadyClosed
		return
	}

}

// ReadFrom for read from offset
func (q *Queue) Read(offset int64) (data []byte, err error) {
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
	qf := q.files[idx]
	q.flock.RUnlock()

	data, err = qf.Read(offset)

	return
}

// StreamRead for stream read
func (q *Queue) StreamRead(offset int64) (ch chan []byte, err error) {
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
	// qf := q.files[idx]
	q.flock.RUnlock()

	return
}

// EndStream for end stream
func (q *Queue) EndStream(ch chan []byte) error {
	return nil
}

var (
	errEmptyDirectory = errors.New("directory empty")
	errAlreadyClosed  = errors.New("already closed")
	errAlreadyClosing = errors.New("already closing")
	errMsgTooLarge    = errors.New("msg too large")
	errMaxPutting     = errors.New("too much putting")
	errInvalidOffset  = errors.New("invalid offset")
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

		close(q.doneCh)
		q.wg.Wait()
		atomic.StoreUint32(&q.closeState, closed)
	})

	return
}

// Delete the queue
func (q *Queue) Delete() error {
	q.Close()
	return os.RemoveAll(q.conf.Directory)
}
