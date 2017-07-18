package sproxy

import (
	"net"
	"time"
)

type stats struct {
	CurrentTaskNum int64 //当前正在执行的任务
	Stopping bool
}

var Stats *stats
func init()  {
	Stats = &stats{0, false}
}

type TimeoutConn struct {
	net.Conn
	timeout time.Duration
}

func NewTimeoutConn(conn net.Conn, timeout time.Duration) net.Conn{
	return &TimeoutConn{
		conn,
		timeout,
	}
}

func (c *TimeoutConn) Read(b []byte) (n int, err error) {
	c.SetReadDeadline(time.Now().Add(c.timeout))
	return c.Read(b)
}

func (c *TimeoutConn) Write(b []byte) (n int, err error) {
	c.SetWriteDeadline(time.Now().Add(c.timeout))
	return c.Write(b)
}