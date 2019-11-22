package util

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestMmap(t *testing.T) {
	fileName := "/tmp/test_util.txt"
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	assert.Assert(t, err == nil)
	defer func() {
		f.Close()
		os.Remove(fileName)
	}()

	size := 64000
	bytes, err := Mmap(f, false, int64(size))
	assert.Assert(t, err == nil && len(bytes) == size)
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
