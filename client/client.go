package client

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer
	mu   sync.Mutex
}

// Connect 连接到 remote-cache 服务器
func Connect(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c := &Client{
		conn: conn,
		r:    bufio.NewReader(conn),
		w:    bufio.NewWriter(conn),
	}
	// 尝试读取欢迎行（可忽略错误）
	c.r.ReadString('\n')
	return c, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.conn.Close()
}

// Do 发送任意命令并返回服务器回复（不包含换行）
func (c *Client) Do(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.w.WriteString(cmd + "\n"); err != nil {
		return "", err
	}
	if err := c.w.Flush(); err != nil {
		return "", err
	}
	line, err := c.r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// Get 读取 key 的值，found 表示是否存在
func (c *Client) Get(key string) (value string, found bool, err error) {
	resp, err := c.Do("GET " + key)
	if err != nil {
		return "", false, err
	}
	if resp == "NOTFOUND" {
		return "", false, nil
	}
	if strings.HasPrefix(resp, "OK ") {
		return strings.TrimPrefix(resp, "OK "), true, nil
	}
	return "", false, fmt.Errorf("unexpected response: %s", resp)
}

// Set 写入 key=value
func (c *Client) Set(key, value string) error {
	resp, err := c.Do("SET " + key + " " + value)
	if err != nil {
		return err
	}
	if strings.HasPrefix(resp, "OK") {
		return nil
	}
	return fmt.Errorf("set failed: %s", resp)
}

// Del 删除 key
func (c *Client) Del(key string) error {
	resp, err := c.Do("DEL " + key)
	if err != nil {
		return err
	}
	if strings.HasPrefix(resp, "OK") {
		return nil
	}
	return fmt.Errorf("del failed: %s", resp)
}
