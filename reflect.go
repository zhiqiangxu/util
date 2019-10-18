package util

import (
	"reflect"
)

// ReplaceFuncVar for replace funcVar with fn
func ReplaceFuncVar(funcVarPtr interface{}, fn func(in []reflect.Value) (out []reflect.Value)) {

	v, ok := funcVarPtr.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(funcVarPtr)
	}

	v = reflect.Indirect(v)

	if v.Kind() != reflect.Func {
		panic("funcVarPtr must point to a func")
	}

	v.Set(reflect.MakeFunc(v.Type(), fn))
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
	fv, ok := fun.(reflect.Value)
	if !ok {
		fv = reflect.ValueOf(fun)
	}

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
	fv, ok := fun.(reflect.Value)
	if !ok {
		fv = reflect.ValueOf(fun)
	}

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

// StructFields for filter fields in struct
func StructFields(s interface{}, filter func(name string, f reflect.Value) bool) (fields map[string]reflect.Value) {
	v, ok := s.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(s)
	}
	v = reflect.Indirect(v)

	fields = make(map[string]reflect.Value)
	t := v.Type()
	count := v.NumField()
	for i := 0; i < count; i++ {
		field := v.Field(i)
		name := t.Field(i).Name
		if filter == nil {
			fields[name] = field
		} else if filter(name, field) {
			fields[name] = field
		}
	}

	return
}

// ScanMethods for scan methods of s
func ScanMethods(s interface{}) (methods []reflect.Value) {
	v, ok := s.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(s)
	}

	count := v.NumMethod()
	for i := 0; i < count; i++ {
		methods = append(methods, v.Method(i))
	}

	return
}
