package lf

import (
	"testing"

	"reflect"

	"gotest.tools/assert"
)

func TestList(t *testing.T) {
	l := NewList(1024)
	isNew, err := l.Insert([]byte("a"), []byte("a"))
	assert.Assert(t, isNew && err == nil)

	v, exists := l.Get([]byte("a"))
	assert.Assert(t, exists && reflect.DeepEqual(v, []byte("a")))

	isNew, err = l.Insert([]byte("a"), []byte("b"))
	assert.Assert(t, !isNew && err == nil)

	v, exists = l.Get([]byte("a"))
	assert.Assert(t, exists && reflect.DeepEqual(v, []byte("b")))

	isNew, err = l.Insert([]byte("b"), []byte("b"))
	assert.Assert(t, isNew && err == nil)

	deleted := l.Delete([]byte("a"))
	assert.Assert(t, deleted)
	v, exists = l.Get([]byte("a"))
	assert.Assert(t, !exists && v == nil)
}
