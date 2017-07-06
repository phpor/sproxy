package sproxy

import (
	"fmt"
	"net"
	"strings"
	"github.com/smallfish/ftp"
	"net/url"
	"strconv"
)

func serveFtpProxy(downstream net.Conn, firstLine string) error {

	defer func() {
		downstream.Close()
		Stats.CurrentTaskNum--
	}()
	arr := strings.Split(firstLine, " ")
	method := arr[0]
	location := arr[1]
	Url, err := url.Parse(location)
	if err != nil {
		return err
	}
	arrHostPort := strings.Split(Url.Host, ":")
	host := arrHostPort[0]
	port := 21
	if len(arrHostPort) > 1 {
		port, _ = strconv.Atoi(arrHostPort[1])
	}
	ftp := new(ftp.FTP)

	//acl
	ftp.Connect(host, port)
	//ftp.Login(Url.User.Username(), Url.User.Password())
	if ftp.Code != 230 {
		return fmt.Errorf(ftp.Message)
	}
	ftp.Request("TYPE I")
	if method == "POST" {
		// 文件的上传和下载没有实现
	}

	return nil
}

func eatHeader(stream net.Conn) error {
	for {
		line, err := readLine(stream)
		if err != nil {
			return err
		}
		if line[len(line)-1:][0] == '\n' {
			break
		}
	}
	return nil
}