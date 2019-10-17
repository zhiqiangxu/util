package util

import (
	"reflect"
	"testing"

	"gotest.tools/assert"
)

func TestReplaceFuncVar(t *testing.T) {
	var f func() int
	ReplaceFuncVar(&f, func([]reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(30)}
	})

	result := f()
	assert.Assert(t, result == 30)

	var s struct {
		F func() int
		I int
	}

	ReplaceFuncVar(&s.F, func([]reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(40)}
	})

	result = s.F()
	assert.Assert(t, result == 40)
}
