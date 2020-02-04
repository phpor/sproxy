package sproxy

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func serveHttpProxy(downstream net.Conn, firstLine string) error {
	defer func() {
		downstream.Close()
		Stats.CurrentTaskNum--
	}()

	headerLine := []string{firstLine}
	hostname := ""
	port := "80"
	for {
		line, err := readLine(downstream)
		if err != nil {
			log.Warning("Read error on: " + downstream.RemoteAddr().String())
			return err
		}
		if len(line) == 0 {
			headerLine = append(headerLine, string(line))
			break
		}
		if len(line) > 6 && strings.ToUpper(string(line[:5])) == "HOST:" {
			host := string(line[7:])
			arr := strings.Split(string(host), ":")
			hostname = arr[0]
			if len(arr) == 2 {
				port = arr[1]
			}
			hostname = strings.Split(string(line), ": ")[1]
		}
		headerLine = append(headerLine, string(line))
	}
	if hostname == "" {
		msg := "http header Host not set"
		_ = log.Err(msg)
		return fmt.Errorf(msg)
	}
	upstream, err := createUpstream(hostname+":"+port, downstream)
	if err != nil {
		if err == ErrAccessForbidden {
			downstream.Write([]byte("HTTP/1.1 403 FORBIDEN\r\n\r\n"))
		}
		return err
	}
	defer upstream.Close()

	for _, line := range headerLine {
		_, err = io.WriteString(upstream, line+"\r\n")
		if err != nil {
			log.Warning(fmt.Sprintf("Write client hello fail: %s", err.Error()))
		}
	}

	ioCopy(downstream, upstream)
	return nil
}
