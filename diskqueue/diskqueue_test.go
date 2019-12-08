package diskqueue

import (
	"bytes"
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestQueue(t *testing.T) {
	conf := Conf{Directory: "/tmp/dq", WriteMmap: true}
	q, err := New(conf)
	assert.Assert(t, err == nil)

	testData := []byte("abc")
	offset, err := q.Put(testData)
	assert.Assert(t, err == nil)

	readData, err := q.Read(offset)
	fmt.Println(string(readData))
	assert.Assert(t, err == nil && bytes.Equal(readData, testData), err)

	err = q.Close()
	assert.Assert(t, err == nil)

}
