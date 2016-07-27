package sproxy

import (
	"net"
	"bufio"
	"strings"
)


func ServeHttpTunnelProxy(downstream net.Conn)  {
	defer func() {
		downstream.Close()
		Stats.CurrentTaskNum--
	}()
	//解析首行，要求必须是connect方法
	rd := bufio.NewReader(downstream)
	headerLine, _, err := rd.ReadLine()
	if err != nil {
		log.Warning("read header line fail")
		return
	}
	arr := strings.Split(string(headerLine), " ")
	method := arr[0]
	if method != "CONNECT" {
		log.Warning("Only support CONNECT method")
		downstream.Write([]byte("HTTP/1.1 405 method not allowed \r\n\r\n"))
		return
	}
	target := arr[1]

	upstream, err := createUpstream(target)
	if err != nil {
		if err == ErrAccessForbidden {
			downstream.Write([]byte("HTTP/1.1 403 FORBIDEN\r\n\r\n"))
		}
		return
	}
	defer upstream.Close()

	for {
		line, hasRemain, err := rd.ReadLine()
		if err != nil {
			log.Warning("Read more header fail: "+ downstream.RemoteAddr().String())
			return
		}
		if !hasRemain && len(line) == 0 {
			break
		}
	}
	downstream.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	ioCopy(downstream, upstream)
}

