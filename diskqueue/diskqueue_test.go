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

	n := 1000
	for i := 0; i < n; i++ {
		offset, err := q.Put(testData)
		assert.Assert(t, err == nil)

		readData, err := q.Read(nil, offset)
		// fmt.Println(string(readData))
		assert.Assert(t, err == nil && bytes.Equal(readData, testData), "%v:%v", err, i)
	}

	assert.Assert(t, q.NumFiles() == 1 && q.FileMeta(0).MsgCount == uint64(n))

	ch, err := q.StreamRead(context.Background(), 0)
	assert.Assert(t, err == nil)
	for i := 0; i < n; i++ {
		readData := <-ch
		assert.Assert(t, bytes.Equal(readData, testData))
	}

	n, err = q.GC()
	assert.Assert(t, err == nil && n == 0)

	err = q.Delete()
	assert.Assert(t, err == nil)

}

func TestFixedQueue(t *testing.T) {
	fixedSizeMsg := []byte("abcd")

	conf := Conf{Directory: "/tmp/dq", WriteMmap: true, CustomDecoder: func(ctx context.Context, r *QfileSizeReader) (data []byte, err error) {
		data = make([]byte, len(fixedSizeMsg))
		err = r.Read(ctx, data)
		return
	}}
	q, err := New(conf)
	assert.Assert(t, err == nil)
	defer func() {
		err = q.Delete()
		assert.Assert(t, err == nil)
	}()

	n := 1000
	for i := 0; i < n; i++ {
		offset, err := q.Put(fixedSizeMsg)
		assert.Assert(t, err == nil)

		readData, err := q.Read(nil, offset)
		// fmt.Println(string(readData))
		assert.Assert(t, err == nil && bytes.Equal(readData, fixedSizeMsg), "%v:%v:%v", err, i, offset)
	}

	assert.Assert(t, q.NumFiles() == 1 && q.FileMeta(0).MsgCount == uint64(n))

	ch, err := q.StreamRead(context.Background(), 0)
	assert.Assert(t, err == nil)
	for i := 0; i < n; i++ {
		readData := <-ch
		assert.Assert(t, bytes.Equal(readData, fixedSizeMsg))
	}

	n, err = q.GC()
	assert.Assert(t, err == nil && n == 0)

}
