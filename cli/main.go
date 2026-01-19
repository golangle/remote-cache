package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	cacheclient "remotecache.golangle.net/client"
)

func main() {
	addr := flag.String("addr", "localhost:12345", "server address")
	oneshot := flag.Bool("c", false, "send single command and exit (use remaining args as command)")
	flag.Parse()

	if *oneshot {
		if flag.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "usage: client -c <COMMAND>")
			os.Exit(2)
		}
		cmd := strings.Join(flag.Args(), " ")
		sendOnce(*addr, cmd)
		return
	}

	cli, err := cacheclient.Connect(*addr)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer cli.Close()

	// 交互模式：读取 stdin 的命令并通过 client 发送
	stdin := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !stdin.Scan() {
			break
		}
		text := stdin.Text()
		if strings.TrimSpace(text) == "" {
			continue
		}
		resp, err := cli.Do(text)
		if err != nil {
			log.Printf("执行失败: %v", err)
			break
		}
		fmt.Println(resp)
		if strings.ToUpper(strings.TrimSpace(text)) == "QUIT" {
			break
		}
	}
}

func sendOnce(addr, cmd string) {
	cli, err := cacheclient.Connect(addr)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer cli.Close()

	resp, err := cli.Do(cmd)
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}
	fmt.Println(resp)
}
