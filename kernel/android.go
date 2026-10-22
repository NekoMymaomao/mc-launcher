package main

// android.go — Android JNI 导出 (CGo 编译模式)
//go:build cshared

/*
#include <stdlib.h>
*/
import "C"

//export StartMCLauncher
func StartMCLauncher(addr *C.char) {
	go func() {
		a := C.GoString(addr)
		if a == "" {
			a = "127.0.0.1:9876"
		}
		Serve(a)
	}()
}
