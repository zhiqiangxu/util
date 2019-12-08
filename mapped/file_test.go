package mapped

import (
	"bytes"
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestFile(t *testing.T) {

	fileName := "/tmp/test_file"
	os.Remove(fileName)
	f, err := CreateFile(fileName, 64000, false, nil)
	assert.Assert(t, err == nil)

	mtime, err := f.LastModified()
	assert.Assert(t, err == nil)
	mtime2, err := f.LastModified()
	assert.Assert(t, err == nil && mtime2 == mtime)

	wbytes := []byte("123")
	n, err := f.Write(wbytes)
	assert.Assert(t, err == nil && n == len(wbytes))

	mtime2, err = f.LastModified()
	assert.Assert(t, err == nil && mtime2 != mtime)

	rbytes := make([]byte, len(wbytes))
	n, err = f.Read(0, rbytes)
	assert.Assert(t, err == nil && n == len(wbytes) && bytes.Equal(wbytes, rbytes))
	err = f.Close()
	assert.Assert(t, err == nil)

	f, err = CreateFile(fileName, 64000, false, nil)
	// 已经存在
	assert.Assert(t, err != nil)
	os.Remove(fileName)
	// 已删除
	f, err = OpenFile(fileName, 64000, os.O_RDWR, false, nil)
	assert.Assert(t, err != nil)
}
