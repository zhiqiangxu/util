package mapped

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestFile(t *testing.T) {
	fileName := "/tmp/test_file"
	os.Remove(fileName)
	f, err := CreateFile(fileName, 64000, false, nil)
	assert.Assert(t, err == nil)
	err = f.Close()
	assert.Assert(t, err == nil)
	f, err = OpenFile(fileName, 64000, os.O_RDWR, false, nil)
	assert.Assert(t, err == nil)
	err = f.Close()
	assert.Assert(t, err == nil)

	f, err = CreateFile(fileName, 64000, false, nil)
	assert.Assert(t, err != nil)
	os.Remove(fileName)
	f, err = OpenFile(fileName, 64000, os.O_RDWR, false, nil)
	assert.Assert(t, err != nil)
}
