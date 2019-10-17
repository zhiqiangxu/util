package util

import (
	"reflect"
)

// ReplaceFuncVar for replace funcVar with fn
func ReplaceFuncVar(funcVarPtr interface{}, fn func(in []reflect.Value) (out []reflect.Value)) {
	v := reflect.ValueOf(funcVarPtr)
	if v.Kind() != reflect.Ptr {
		panic("funcVarPtr must be a pointer")
	}

	e := v.Elem()
	e.Set(reflect.MakeFunc(e.Type(), fn))
}

// Func2Value wraps a func with reflect.Value
func Func2Value(fun interface{}) reflect.Value {
	v := reflect.ValueOf(fun)
	if v.Kind() != reflect.Func {
		panic("fun must be a func")
	}
	return v
}

// FuncInputTypes for retrieve func input types
func FuncInputTypes(fun interface{}) (result []reflect.Type) {
	fv := reflect.ValueOf(fun)
	if fv.Kind() != reflect.Func {
		panic("fun must be a func")
	}

	tp := fv.Type()
	n := tp.NumIn()
	for i := 0; i < n; i++ {
		result = append(result, tp.In(i))
	}

	return
}

// FuncOutputTypes for retrieve func output types
func FuncOutputTypes(fun interface{}) (result []reflect.Type) {
	fv := reflect.ValueOf(fun)
	if fv.Kind() != reflect.Func {
		panic("fun must be a func")
	}

	tp := fv.Type()
	n := tp.NumOut()
	for i := 0; i < n; i++ {
		result = append(result, tp.Out(i))
	}

	return
}

// TypeByPointer for retrieve reflect.Type by a pointer value
func TypeByPointer(tp interface{}) reflect.Type {
	return reflect.TypeOf(tp).Elem()
}

// InstanceByType returns a instance of reflect.Type wrapped in interface{}
func InstanceByType(t reflect.Type) interface{} {
	return reflect.New(t).Elem().Interface()
}
