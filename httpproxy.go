package sproxy

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"io"
)

func ServeHttpProxy(downstream net.Conn)  {
	defer func() {
		downstream.Close()
		Stats.CurrentTaskNum--
	}()
	var headerLine []string
	rd := bufio.NewReader(downstream)
	var line []byte
	hostname := ""
	port := "80"
	for {
		_line, hasRemain, err := rd.ReadLine()
		if err != nil {
			log.Warning("Read error on: " + downstream.RemoteAddr().String())
			return
		}
		line = append(line, _line...)
		if len(line) == 0 {
			headerLine = append(headerLine, string(line))
			break
		}
		if ! hasRemain {
			if len(line) > 6 && string(line[:5]) == "Host:" {
				host := string(line[7:])
				arr := strings.Split(string(host), ":")
				hostname = arr[0]
				if len(arr) == 2 {
					port = arr[1]
				}
				hostname = strings.Split(string(line), ": ")[1]
			}
			headerLine = append(headerLine, string(line))
			line = line[:0]
		}
	}
	upstream, err := createUpstream(hostname+":"+port)
	if err != nil {
		if err == ErrAccessForbidden {
			downstream.Write([]byte("HTTP/1.1 403 FORBIDEN\r\n\r\n"))
		}
		return
	}
	defer upstream.Close()

	for _, line := range headerLine {
		_, err = io.WriteString(upstream, string(line) + "\r\n")
		if err != nil {
			log.Warning(fmt.Sprintf("Write client hello fail: %s", err.Error()))
		}
	}

	ioCopy(downstream, upstream)
}

