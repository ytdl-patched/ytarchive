//go:build cgo

package main

// #cgo pkg-config: python3
// #cgo LDFLAGS: -lpython3 -lpython3.9
// #include <Python.h>
//
// char* uintptrPyFormatC(void);
// // Workaround missing variadic function support
// // https://github.com/golang/go/issues/975
// // https://docs.python.org/ja/3/c-api/arg.html#c.Py_BuildValue
// int PyArg_ParseTuple_s(PyObject* args, char* cz) {
//     return PyArg_ParseTuple(args, "s", cz);
// }
// int PyArg_ParseTuple_Xsss(PyObject* args, void* ptr, char* c1, char* c2, char* c3) {
//     return PyArg_ParseTuple(args, sprintf("%s sss", uintptrPyFormatC()), ptr, c1, c2, c3);
// }
// PyObject* uintPtrToPyInt(void* ptr) {
//     return Py_BuildValue(uintptrPyFormatC(), ptr);
// }
// int ui_size = sizeof(unsigned int);
// int ul_size = sizeof(unsigned long);
// int ull_size = sizeof(unsigned long long);
//
// PyObject* initializePy(PyObject*, PyObject*);
// PyObject* registerFormatPy(PyObject*, PyObject*);
// PyObject* loadCookiesPy(PyObject*, PyObject*);
// PyObject* runDownloaderPy(PyObject*, PyObject*);
// PyObject* interruptPy(PyObject*, PyObject*);
// PyObject* pollPy(PyObject*, PyObject*);
import "C"
import "unsafe"

func uintptrPyFormatC() *C.char {
	return C.CString(uintptrPyFormat())
}
func uintptrPyFormat() string {
	// https://docs.python.org/ja/3/c-api/arg.html#c.Py_BuildValue
	var a uintptr
	size := unsafe.Sizeof(a)
	ret := "I"
	if size == C.ull_size {
		ret = "K"
	}
	if size == C.ul_size {
		ret = "k"
	}
	if size == C.ui_size {
		ret = "I"
	}
	return ret
}

//export initializePy
func initializePy(self, args *C.PyObject) *C.PyObject {
	var videoId *C.char
	if C.PyArg_ParseTuple_s(args, &videoId) == 0 {
		C.PyErr_SetString(C.PyExc_TypeError, "Wrong parameters for initialize")
		return nil
	}
	return C.uintPtrToPyInt(initialize(videoId))
}

//export registerFormatPy
func registerFormatPy(self, args *C.PyObject) *C.PyObject {
	var ptr uintptr
	var fmtUrlC *C.char
	var manifestUrlC *C.char
	var filepathC *C.char
	if C.PyArg_ParseTuple_Xsss(args, &ptr, &fmtUrlC, &manifestUrlC, &filepathC) == 0 {
		C.PyErr_SetString(C.PyExc_TypeError, "Wrong parameters for registerFormat")
		return nil
	}
	registerFormat(ptr, fmtUrlC, manifestUrlC, filepathC)
	// no return value
	return nil
}
//export loadCookiesPy
func loadCookiesPy(self, args *C.PyObject) *C.PyObject {

}
//export runDownloaderPy
func runDownloaderPy(self, args *C.PyObject) *C.PyObject {

}
//export interruptPy
func interruptPy(self, args *C.PyObject) *C.PyObject {

}
//export pollPy
func pollPy(self, args *C.PyObject) *C.PyObject {

}

// func main() {}
