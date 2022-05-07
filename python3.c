#define Py_LIMITED_API
#include <Python.h>

PyObject * sum(PyObject *, PyObject *);

// Workaround missing variadic function support
// https://github.com/golang/go/issues/975

static PyMethodDef YtaMethods[] = {
    {"sum", sum, METH_VARARGS, "Add two numbers."},
    {NULL, NULL, 0, NULL}
};

static struct PyModuleDef ytamod = {
   PyModuleDef_HEAD_INIT, "ytarchive", NULL, -1, YtaMethods
};

PyMODINIT_FUNC
PyInit_ytarchive(void) {
    return PyModule_Create(&ytamod);
}