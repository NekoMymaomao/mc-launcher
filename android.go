package main

// android.go — Android APK CGo 
// : go build -buildmode=c-shared -o libmc_launcher.so android.go main.go *.go
package main

/*
#include <stdlib.h>
extern void mc_server_start(const char* addr);
*/
import "C"
import "net"
import "os"

//export StartMCLauncher
func StartMCLauncher(addr *C.char) {
	go func() {
		a := C.GoString(addr)
		if a == "" { a = "127.0.0.1:9876" }
		listener, err := net.Listen("tcp", a)
		if err != nil {
			os.Stderr.WriteString(": " + err.Error() + "\n")
			return
		}
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil { continue }
			go handleConn(conn)
		}
	}()
}
