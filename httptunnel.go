package sproxy

import (
	"net"
	"bufio"
	"strings"
	"net/url"
	"fmt"
	"crypto/tls"
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
	downstream.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))

	ioCopy(downstream, upstream)
	return nil
}

func httpProxyInHttpProxy(downstream net.Conn, backend string) error {
	defer func() {
		downstream.Close()
		Stats.CurrentTaskNum--
	}()

	URL, err := url.Parse(backend)
	if err != nil {
		return err
	}
	host := URL.Host
	port := ""
	arr := strings.Split(host, ":")
	ip := arr[0]
	if len(arr) == 2 {
		port = arr[1]
	}
	var upstream net.Conn
	if URL.Scheme == "https" {
		if port == "" {
			port = "443"
		}
		ipPort := fmt.Sprintf("%s:%s", ip , port)
		upstream,err = tls.Dial("tcp", ipPort, nil)
	} else if URL.Scheme == "http" {
		if port == "" {
			port = "80"
		}
		ipPort := fmt.Sprintf("%s:%s", ip , port)
		upstream,err = net.Dial("tcp", ipPort)

	} else {
		return fmt.Errorf("only support http or https backend")
	}
	if err != nil {
		return err
	}
	defer func() {
		upstream.Close()
	}()

	ioCopy(downstream, upstream)
	return nil
}
