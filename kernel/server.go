package main

// server.go — TCP 服务器 (非 JNI 编译模式)
//go:build !cshared

import (
	"encoding/json"
	"net"
)

// Serve 启动 TCP 服务器监听指定地址
func Serve(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logError("Listen failed: %v", err)
		return
	}
	defer listener.Close()
	logInfo("Server listening on %s", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logError("Accept error: %v", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 65536)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		var req Request
		if err := json.Unmarshal(buf[:n], &req); err != nil {
			writeResp(conn, Response{Ok: false, Error: i18n("json_parse_error")})
			continue
		}
		resp := dispatch(req)
		data, _ := json.Marshal(resp)
		data = append(data, '\n')
		conn.Write(data)
	}
}

func writeResp(conn net.Conn, resp Response) {
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	conn.Write(data)
}
