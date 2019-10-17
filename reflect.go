package util

import "reflect"

// ReplaceFuncVar for replace funcVar with fn
func ReplaceFuncVar(funcVarPtr interface{}, fn func(in []reflect.Value) (out []reflect.Value)) {
	v := reflect.ValueOf(funcVarPtr)
	if v.Kind() != reflect.Ptr {
		panic("funcVarPtr must be a pointer")
	}

	e := v.Elem()
	e.Set(reflect.MakeFunc(e.Type(), fn))
}
