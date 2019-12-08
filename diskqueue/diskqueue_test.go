package diskqueue

import (
	"bytes"
	"testing"

	"gotest.tools/assert"
)

func TestQueue(t *testing.T) {
	conf := Conf{Directory: "/tmp/dq", WriteMmap: true}
	q, err := New(conf)
	assert.Assert(t, err == nil)

	testData := []byte("abc")

	for i := 0; i < 1000; i++ {
		offset, err := q.Put(testData)
		assert.Assert(t, err == nil)

		readData, err := q.Read(offset)
		// fmt.Println(string(readData))
		assert.Assert(t, err == nil && bytes.Equal(readData, testData), err)
	}

	err = q.Close()
	assert.Assert(t, err == nil)

}
