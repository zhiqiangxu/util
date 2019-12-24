package diskqueue

import (
	"bytes"
	"context"
	"testing"

	"gotest.tools/assert"
)

func TestQueue(t *testing.T) {
	conf := Conf{Directory: "/tmp/dq", WriteMmap: true}
	q, err := New(conf)
	assert.Assert(t, err == nil)

	testData := []byte("abcd")

	for i := 0; i < 1000; i++ {
		offset, err := q.Put(testData)
		assert.Assert(t, err == nil)

		readData, err := q.Read(offset)
		// fmt.Println(string(readData))
		assert.Assert(t, err == nil && bytes.Equal(readData, testData), err)
	}

	ch, err := q.StreamRead(context.Background(), 0)
	assert.Assert(t, err == nil)
	for i := 0; i < 1000; i++ {
		readData := <-ch
		assert.Assert(t, bytes.Equal(readData, testData))
	}

	err = q.Delete()
	assert.Assert(t, err == nil)

}
