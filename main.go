package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

func main() {
	// 简单内存缓存，供 socket 接入使用
	var cache sync.Map

	// 启动 TCP 服务
	addr := ":12345"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("监听 %s 失败: %v", addr, err)
	}
	log.Printf("TCP 服务已启动，监听 %s", addr)

	// 后台示例：向 cache 写入一些初始数据
	go func() {
		cache.Store("user_1", "Alice")
		cache.Store("user_2", "Bob")
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept 错误: %v", err)
			continue
		}
		go handleConn(conn, &cache)
	}
}

// 处理单个连接的简单协议：
// 支持命令：GET key | SET key value | DEL key | ECHO text | HELP | QUIT
func handleConn(conn net.Conn, cache *sync.Map) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	log.Printf("连接来自 %s", addr)
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	write := func(s string) {
		w.WriteString(s)
		if !strings.HasSuffix(s, "\r\n") {
			w.WriteString("\r\n")
		}
		w.Flush()
	}

	write("Welcome to simple-cache TCP server. Type HELP for commands.")

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			log.Printf("%s 连接关闭: %v", addr, err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		cmd := strings.ToUpper(parts[0])
		switch cmd {
		case "GET":
			if len(parts) < 2 {
				write("ERR missing key")
				continue
			}
			key := parts[1]
			if v, ok := cache.Load(key); ok {
				write(fmt.Sprintf("OK %v", v))
			} else {
				write("NOTFOUND")
			}
		case "SET":
			if len(parts) < 3 {
				write("ERR usage: SET key value")
				continue
			}
			key := parts[1]
			val := parts[2]
			cache.Store(key, val)
			write("OK")
		case "DEL":
			if len(parts) < 2 {
				write("ERR missing key")
				continue
			}
			key := parts[1]
			cache.Delete(key)
			write("OK")
		case "ECHO":
			if len(parts) < 2 {
				write("")
			} else {
				// ECHO 后面可能包含空格，取原始行
				// parts[1] 只包含第一个空格后的内容，若要完整，取 line[5:]
				if len(line) >= 5 {
					write(line[5:])
				} else {
					write("")
				}
			}
		case "HELP":
			write("Commands: GET key | SET key value | DEL key | ECHO text | HELP | QUIT")
		case "QUIT":
			write("BYE")
			return
		default:
			write("ERR unknown command")
		}
	}
}
