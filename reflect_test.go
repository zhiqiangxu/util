package util

import (
	"reflect"
	"testing"

	"gotest.tools/assert"
)

func TestReflect(t *testing.T) {
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

	vf := Func2Value(f)
	ret := vf.Call(nil)
	assert.Assert(t, ret[0].Interface().(int) == 30)

	inTypes := FuncInputTypes(testTarget)
	assert.Assert(t, len(inTypes) == 2 && inTypes[0].Kind() == reflect.Int && inTypes[1].Kind() == reflect.String)

	outTypes := FuncOutputTypes(testTarget)
	assert.Assert(t, len(outTypes) == 1 && outTypes[0].Kind() == reflect.Slice)

	stringType := TypeByTypePointer((*string)(nil))
	assert.Assert(t, stringType == reflect.ValueOf("").Type())
}

func testTarget(int, string) []int {
	return nil
}
