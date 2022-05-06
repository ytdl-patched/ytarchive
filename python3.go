//go:build cgo

package main

// #cgo pkg-config: python3
// #cgo LDFLAGS: -lpython3 -lpython3.9
// #include <Python.h>
//
// int PyArg_ParseTuple_LL(PyObject * args, long long * a, long long * b);
import "C"

//export sum
func sum(self, args *C.PyObject) *C.PyObject {
	var a, b C.longlong
	if C.PyArg_ParseTuple_LL(args, &a, &b) == 0 {
		return nil
	}
	return C.PyLong_FromLongLong(a + b)
}

// func main() {}
