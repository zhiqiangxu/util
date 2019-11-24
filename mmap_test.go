package util

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestMmap(t *testing.T) {
	fileName := "/tmp/test_util.txt"
	os.Remove(fileName)
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	assert.Assert(t, err == nil)
	defer func() {
		f.Close()
		os.Remove(fileName)
	}()

	size := 64000
	n, err := f.Write(make([]byte, size))
	assert.Assert(t, n == size && err == nil)
	bytes, err := Mmap(f, true, int64(size))
	assert.Assert(t, err == nil && len(bytes) == size)
	bytes[0] = 0
	err = Madvise(bytes, true)
	assert.Assert(t, err == nil)
	err = Madvise(bytes, false)
	assert.Assert(t, err == nil)
	err = Munmap(bytes)
	assert.Assert(t, err == nil)
	err = Madvise(bytes, false)
	assert.Assert(t, err != nil)
	err = Munmap(bytes)
	assert.Assert(t, err != nil)

}
