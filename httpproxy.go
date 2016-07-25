package sproxy

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"io"
)

func ServeHttpProxy(c net.Conn)  {
	defer func() {
		c.Close()
		Stats.CurrentTaskNum--
	}()
	var headerLine []string
	rd := bufio.NewReader(c)
	var line []byte
	hostname := ""
	port := "80"
	for {
		_line, hasRemain, err := rd.ReadLine()
		if err != nil {
			log.Warning("Read error on: " + c.RemoteAddr().String())
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
	backendAddr := conf.GetBackend(hostname)
	if backendAddr == "" {
		log.Warning(fmt.Sprintf("Access %s is not allow", hostname))
		return
	}
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
	for _, line := range headerLine {
		_, err = io.WriteString(conn, string(line) + "\r\n")
		if err != nil {
			log.Warning(fmt.Sprintf("Write client hello fail: %s", err.Error()))
		}
	}

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

