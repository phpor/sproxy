package sproxy

import (
	"net"
	"bufio"
	"strings"
)


func serveHttpTunnelProxy(downstream net.Conn, firstLine string) error {
	defer func() {
		downstream.Close()
		Stats.CurrentTaskNum--
	}()
	arr := strings.Split(firstLine, " ")
	target := arr[1]

	upstream, err := createUpstream(target, downstream)
	if err != nil {
		if err == ErrAccessForbidden {
			downstream.Write([]byte("HTTP/1.1 403 FORBIDEN\r\n\r\n"))
		}
		return err
	}
	defer upstream.Close()

	rd := bufio.NewReader(downstream)
	for {
		line, hasRemain, err := rd.ReadLine()
		if err != nil {
			log.Warning("Read more header fail: "+ downstream.RemoteAddr().String())
			return err
		}
		if !hasRemain && len(line) == 0 {
			break
		}
	}
	downstream.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	ioCopy(downstream, upstream)
	return nil
}

