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

	fv := reflect.ValueOf(&f)
	ReplaceFuncVar(fv, func([]reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(31)}
	})

	result = f()
	assert.Assert(t, result == 31)

	s := &struct {
		F func() int
		I int
	}{}

	ReplaceFuncVar(&s.F, func([]reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(40)}
	})

	result = s.F()
	assert.Assert(t, result == 40)

	sv := reflect.ValueOf(s)
	ReplaceFuncVar(reflect.Indirect(sv).Field(0), func([]reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(41)}
	})

	result = s.F()
	assert.Assert(t, result == 41)

	sfields := StructFields(s, func(_ string, field reflect.Value) bool {
		return field.Kind() == reflect.Int
	})
	assert.Assert(t, len(sfields) == 1)
	sfields["I"].Set(reflect.ValueOf(20))
	assert.Assert(t, s.I == 20)

	vf := Func2Value(f)
	ret := vf.Call(nil)
	assert.Assert(t, ret[0].Interface().(int) == 31)

	inTypes := FuncInputTypes(testTarget)
	assert.Assert(t, len(inTypes) == 2 && inTypes[0].Kind() == reflect.Int && inTypes[1].Kind() == reflect.String)

	outTypes := FuncOutputTypes(testTarget)
	assert.Assert(t, len(outTypes) == 1 && outTypes[0].Kind() == reflect.Slice)

	stringType := TypeByPointer((*string)(nil))
	assert.Assert(t, stringType == reflect.ValueOf("").Type())

	is := InstanceByType(stringType)
	_, ok := is.(string)
	assert.Assert(t, ok)
	isPtr := InstancePtrByType(stringType)
	_, ok = isPtr.(*string)
	assert.Assert(t, ok)

	var t2 TestType
	methods := ScanMethods(t2)
	_, ok = methods["M1"]
	assert.Assert(t, len(methods) == 1 && ok)

	methods = ScanMethods(&t2)
	_, ok = methods["M2"]
	assert.Assert(t, len(methods) == 3 && ok, "%v %v", len(methods), ok)

	{
		s := "abc"
		sv := reflect.ValueOf(s)
		sptr := InstancePtrByClone(sv)
		sp, ok := sptr.(*string)
		assert.Assert(t, ok)
		*sp = "def"
		assert.Assert(t, s == "abc")
	}

	var itf interface{}
	itf = t2
	_, ok = itf.(interface{ M2() })
	assert.Assert(t, !ok)
	itf = &itf
	_, ok = itf.(interface{ M2() })
	assert.Assert(t, !ok)
	itf = &t2
	_, ok = itf.(interface{ M2() })
	assert.Assert(t, ok)
}

func testTarget(int, string) []int {
	return nil
}

type TestType struct {
	Base
}

type Base struct {
}

func (t TestType) M1() {

}

func (t TestType) m1() {

}

func (t *TestType) M2() {
}

func (b *Base) OK() {

}
