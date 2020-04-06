package mmr

import (
	"errors"
	"io"
	"os"
	"unsafe"
)

type fileHashStore struct {
	fileName string
	file     *os.File
}

// NewFileHashStore returns a HashStore implement in file
func NewFileHashStore(name string, treeSize uint64) (hs HashStore, err error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return
	}
	store := &fileHashStore{
		fileName: name,
		file:     f,
	}

	hashCount := getStoredHashCount(treeSize)
	size := int64(hashCount) * int64(unsafe.Sizeof(HashType{}))

	err = store.checkConsistence(size)
	if err != nil {
		return
	}

	_, err = store.file.Seek(size, io.SeekStart)
	if err != nil {
		return
	}

	hs = store
	return
}

var (
	errStoredHashLessThanExpected = errors.New("stored hashes are less than expected")
)

func (self *fileHashStore) checkConsistence(fileSize int64) (err error) {

	stat, err := self.file.Stat()
	if err != nil {
		return
	}
	if stat.Size() < fileSize {
		err = errStoredHashLessThanExpected
		return
	}

	return
}

func (self *fileHashStore) Append(hash []HashType) error {
	buf := make([]byte, 0, len(hash)*int(unsafe.Sizeof(HashType{})))
	for _, h := range hash {
		buf = append(buf, h[:]...)
	}
	_, err := self.file.Write(buf)
	return err
}

func (self *fileHashStore) Flush() error {
	return self.file.Sync()
}

func (self *fileHashStore) Close() {
	self.file.Close()
}

func (self *fileHashStore) GetHash(pos uint64) (h HashType, err error) {
	h = unknownHash
	_, err = self.file.ReadAt(h[:], int64(pos)*int64(unsafe.Sizeof(HashType{})))
	if err != nil {
		return
	}

	return
}

func getStoredHashCount(treeSize uint64) int64 {
	subtreesize := getMoutainSizes(treeSize)
	sum := int64(0)
	for _, v := range subtreesize {
		sum += int64(v)
	}

	return sum
}
