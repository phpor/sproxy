package sproxy

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"io"
)


func ServeHttpTunnelProxy(c net.Conn)  {
	defer func() {
		c.Close()
		Stats.CurrentTaskNum--
	}()
	rd := bufio.NewReader(c)
	headerLine, _, err := rd.ReadLine()
	if err != nil {
		log.Warning("read header line fail")
		return
	}
	arr := strings.Split(string(headerLine), " ")
	method := arr[0]
	if method != "CONNECT" {
		log.Warning("Only support CONNECT method")
		c.Write([]byte("HTTP/1.1 405 method not allowed \r\n\r\n"))
		return
	}
	target := arr[1]
	if ! conf.IsAccessAllow(target) {
		log.Warning("Not allow access: " + target + c.RemoteAddr().String())
		c.Write([]byte("HTTP/1.1 403 FORBIDEN\r\n\r\n"))
		return
	}
	for {
		line, hasRemain, err := rd.ReadLine()
		if err != nil {
			log.Warning("Read more header fail: "+c.RemoteAddr().String())
			return
		}
		if !hasRemain && len(line) == 0 {
			break
		}
	}
	c.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	arrTarget := strings.Split(target, ":")
	hostname := arrTarget[0]
	port := arrTarget[1]
	hostip, err := nslookup(hostname)
	if err != nil {
		log.Warning("Nslookup fail: " + hostname)
		return
	}
	dst := fmt.Sprintf("%s:%s", hostip, port)
	conn, err := net.Dial("tcp", dst)
	if err != nil {
		log.Warning(fmt.Sprintf("connect %s fail\n", dst))
		return
	}
	log.Debug(fmt.Sprintf("connected to %s\n", dst))
	defer conn.Close()

	go func() {
		_, err := io.Copy(conn, c) //copy to upstream
		if err != nil {
			//log.Warning(fmt.Sprintf("3: error: %s\n", err.Error()))
		}
	}()
	_ ,err = io.Copy(c, conn) //copy to downstream
	if err != nil {
		//这个错误基本是因为数据输出完了而关闭导致的, 这里写的很有可能会出问题
	}
}

