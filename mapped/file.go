package mapped

import (
	"bytes"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/zhiqiangxu/util"
)

type fileInterface interface {
	Flags() int
	Resize(newSize int64) (err error)
	Write(data []byte) (n int, err error)
	Read(offset int64, data []byte) (n int, err error)
	Commit() (err error)
	Sync() (err error)
	LastModified() (t time.Time, err error)
	Close() (err error)
	MappedBytes() []byte
}

var _ fileInterface = (*File)(nil)

// File for mmaped file
// Write/Resize should be called sequentially
// Commit/Write is concurrent safe
type File struct {

	// 有些字段仅在可写时有意义，trade some memory for better locality
	cwmu           sync.Mutex
	wrotePosition  int64
	commitPosition int64 // 仅在有写缓冲的情况使用
	writeBuffer    *bytes.Buffer
	pool           *sync.Pool

	mu       sync.RWMutex
	fileSize int64
	fileName string
	fmap     []byte
	file     *os.File
	flags    int
	wmm      bool
}

// OpenFile opens a mmaped file
func OpenFile(fileName string, fileSize int64, flags int, wmm bool, pool *sync.Pool) (f *File, err error) {
	f = &File{fileSize: fileSize, fileName: fileName, flags: flags, wmm: wmm}
	if flags&(os.O_RDWR|os.O_WRONLY) != 0 {
		// 写场景可配置缓冲池
		if pool != nil {
			f.pool = pool
			f.writeBuffer = pool.Get().(*bytes.Buffer)
		}
	} else {
		// 只读场景不需要缓冲池
		if pool != nil {
			err = errPoolForReadonly
			return
		}
	}

	// 新建，或打开已有文件，取决于flags
	err = f.init()
	return
}

// CreateFile creates a mmaped file
func CreateFile(fileName string, fileSize int64, wmm bool, pool *sync.Pool) (f *File, err error) {
	// 新建
	return OpenFile(fileName, fileSize, os.O_RDWR|os.O_CREATE|os.O_EXCL, wmm, pool)
}

// Flags for get file flags
func (f *File) Flags() int {
	return f.flags
}

var (
	errPoolForReadonly = errors.New("pool for readonly file")
	errWriteBeyond     = errors.New("write beyond")
	errReadBeyond      = errors.New("read beyond")
)

// init仅在构造函数中调用，所以不需要考虑并发
func (f *File) init() (err error) {

	f.file, err = os.OpenFile(f.fileName, f.flags, 0600)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			// 如果出错，及时释放资源
			f.file.Close()
			f.finalize()
		}
	}()

	stat, err := f.file.Stat()
	if err != nil {
		return
	}

	// 此时f.fileSize的意义：
	// 如果是新建，表示期望的大小
	// 如果是打开已有文件，表示从哪里继续写
	if f.flags&os.O_EXCL != 0 {
		// 新建
		err = f.file.Truncate(f.fileSize)
		if err != nil {
			return
		}
	} else {
		// 打开已有文件

		fileSize := stat.Size()
		offset := f.fileSize

		// offset > 实际大小，写超
		if offset > fileSize {
			err = errWriteBeyond
			return
		}

		// offset <= 实际大小，以实际大小为准
		f.fileSize = fileSize

		f.wrotePosition = offset
		if f.writeBuffer != nil {
			f.commitPosition = offset
		}
	}

	f.fmap, err = util.Mmap(f.file, f.wmm, f.fileSize)
	if err != nil {
		return
	}

	return
}

// MLock for the whole file
func (f *File) MLock() (err error) {
	err = util.MLock(f.fmap, len(f.fmap))
	return
}

// MUnlock for the whole file
// The memory lock on an address range is automatically removed if the address range is unmapped via munmap(2).
func (f *File) MUnlock() (err error) {
	err = util.MUnlock(f.fmap, len(f.fmap))
	return
}

// Resize will do truncate and remmap
func (f *File) Resize(newSize int64) (err error) {
	f.mu.Lock()
	defer f.mu.RUnlock()

	err = f.file.Truncate(newSize)
	if err != nil {
		return
	}

	if f.fmap != nil {
		err = util.MSync(f.fmap, int64(len(f.fmap)), syscall.MS_SYNC)
		if err != nil {
			return
		}

		err = util.Munmap(f.fmap)
		if err != nil {
			return
		}
	}

	f.fmap, err = util.Mmap(f.file, f.wmm, newSize)
	if err != nil {
		return
	}

	f.fileSize = newSize
	if f.wrotePosition >= newSize {
		f.wrotePosition = newSize - 1
	}
	return
}

func (f *File) getWrotePosition() int64 {
	return atomic.LoadInt64(&f.wrotePosition)
}

func (f *File) addAndGetWrotePosition(n int64) (new int64) {
	new = f.wrotePosition + n
	atomic.StoreInt64(&f.wrotePosition, new)
	return
}

func (f *File) getReadPosition() int64 {
	if f.writeBuffer != nil {
		return f.getCommitPosition()
	}

	return f.getWrotePosition()
}

func (f *File) getCommitPosition() int64 {
	return atomic.LoadInt64(&f.commitPosition)
}

func (f *File) addAndGetCommitPosition(n int64) (new int64) {
	new = f.commitPosition + n
	atomic.StoreInt64(&f.commitPosition, new)
	return
}

func (f *File) Write(data []byte) (n int, err error) {
	if f.wrotePosition+int64(len(data)) > f.fileSize {
		err = errWriteBeyond
		return
	}

	n, err = f.doWrite(data)
	return
}

func (f *File) doWrite(data []byte) (n int, err error) {

	// 写缓冲区
	if f.writeBuffer != nil {
		f.cwmu.Lock()
		n, err = f.writeBuffer.Write(data)
		f.cwmu.Unlock()
		f.addAndGetWrotePosition(int64(n))
		return
	}

	// 写共享内存
	if f.wmm {
		copy(f.fmap[f.wrotePosition:], data)
		n = len(data)
		f.addAndGetWrotePosition(int64(n))
		return
	}

	// 写文件
	n, err = f.file.Write(data)
	f.addAndGetWrotePosition(int64(n))
	return

}

// Commit buffer to os if any
func (f *File) Commit() (err error) {
	if f.writeBuffer == nil {
		return
	}

	f.cwmu.Lock()
	defer f.cwmu.Unlock()

	if f.writeBuffer.Len() == 0 {
		return
	}

	// 从缓冲区到共享内存或者文件

	if f.wmm {
		copy(f.fmap[f.commitPosition:], f.writeBuffer.Bytes())
		f.addAndGetCommitPosition(int64(f.writeBuffer.Len()))
		f.writeBuffer.Reset()
		return
	}

	n, err := f.writeBuffer.WriteTo(f.file)
	f.addAndGetCommitPosition(n)

	return
}

// Sync from os to disk
func (f *File) Sync() (err error) {
	if f.wmm {
		err = util.MSync(f.fmap, 0, len(f.fmap))
		return
	}

	err = f.file.Sync()
	return
}

// LastModified returns last modified time
func (f *File) LastModified() (t time.Time, err error) {
	stat, err := f.file.Stat()
	if err != nil {
		return
	}
	t = stat.ModTime()
	return
}

// Read bytes from offset
func (f *File) Read(offset int64, data []byte) (n int, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	readPosition := f.getReadPosition()
	if offset > readPosition {
		err = errReadBeyond
		return
	}

	readTo := offset + int64(len(data)) - 1
	if readTo > readPosition {
		readTo = readPosition
	}
	copy(data, f.fmap[offset:readTo+1])
	n = int(readTo - offset + 1)

	return
}

// Close the mapped file
func (f *File) Close() (err error) {
	err = f.file.Close()
	if err != nil {
		return
	}
	err = util.Munmap(f.fmap)
	if err != nil {
		return
	}
	f.fmap = nil

	f.finalize()
	return
}

// MappedBytes is valid until next Resize
func (f *File) MappedBytes() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.fmap
}

// return stuff to pools for write mode
func (f *File) finalize() {
	if f.writeBuffer == nil {
		return
	}

	f.writeBuffer.Reset()
	f.pool.Put(f.writeBuffer)
	f.writeBuffer = nil
	f.pool = nil

	return
}
