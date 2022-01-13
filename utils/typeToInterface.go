package utils

import "reflect"

// TypeToInterface converts a type to an interface.
func TypeToInterface(args interface{}) (out []interface{}, ok bool) {
	slice, success := takeArg(args, reflect.Slice)
	if !success {
		ok = false
		return
	}
	c := slice.Len()
	out = make([]interface{}, c)
	for i := 0; i < c; i++ {
		out[i] = slice.Index(i).Interface()
	}
	return
}

// takeArg takes an interface as a param and return its value (value type)
func takeArg(args interface{}, kind reflect.Kind) (value reflect.Value, ok bool) {
	value = reflect.ValueOf(args)
	if value.Kind() == kind {
		ok = true
	}
	return
}
