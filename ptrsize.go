//go:build ignore

package main

// #include <stdio.h>
//
// void test1(void) {
//     // printf("%d\n", sizeof(GoUintptr));
// }
//
// int int_size = sizeof(unsigned long long);
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	var a uintptr
	// a = 0
	fmt.Printf("#define GOLANG_UINTPTR_SIZE %d\n", unsafe.Sizeof(a))
	fmt.Printf("#define test2 %d\n", C.int_size)
}
