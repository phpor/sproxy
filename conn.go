package sproxy

import (
	"net"
	"time"
	"github.com/phpor/sproxy/proxy"
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
	return c.Conn.Read(b)
}

func (c *TimeoutConn) Write(b []byte) (n int, err error) {
	c.SetWriteDeadline(time.Now().Add(c.timeout))
	return c.Conn.Write(b)
}


type TimeoutDailer struct {
	timeout time.Duration
}

func NewTimeoutDailer(timeout time.Duration) proxy.Dialer{
	return &TimeoutDailer{
		timeout,
	}
}

func (d *TimeoutDailer) Dial(network, address string) (net.Conn, error) {
	return net.DialTimeout(network, address, d.timeout)
}

