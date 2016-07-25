package sproxy

import (
	"fmt"
	"net"
	"github.com/phpor/sproxy/tls"
	"io"
	"strings"
)


func ServeSniProxy(c net.Conn)  {
	defer func() {
		c.Close()
		Stats.CurrentTaskNum--
	}()
	//set timeout to connection
	clientHelloMsg, err := tls.ReadClientHello(c)
	i := 0
	i++ ;fmt.Println(i)
	if err != nil {
		log.Err("Error read client hello:  " + err.Error())
		return
	}
	i++ ;fmt.Println(i)
	if clientHelloMsg.ServerName == "" {
		log.Warning("Client has no sni: "+ c.LocalAddr().String() +" <= "+ c.RemoteAddr().String())
		return
	}
	i++ ;fmt.Println(i)
	hostname := clientHelloMsg.ServerName
	backendAddr := conf.GetBackend(hostname)
	if backendAddr == "" {
		log.Warning(fmt.Sprintf("Access %s is not allow", hostname))
		return
	}
	port := strings.Split(backendAddr, ":")[1]
	i++ ;fmt.Println(i)
	hostip, err := nslookup(hostname)
	if err != nil {
		log.Warning("Nslookup fail: " + hostname)
		return
	}
	i++ ;fmt.Println(i)
	dst := fmt.Sprintf("%s:%s", hostip, port)
	conn, err := net.Dial("tcp", dst)
	if err != nil {
		log.Warning(fmt.Sprintf("connect %s fail\n", dst))
		return
	}
	i++ ;fmt.Println(i)
	log.Debug(fmt.Sprintf("connected to %s\n", dst))
	defer conn.Close()

	i++ ;fmt.Println(i)
	_, err = io.WriteString(conn, string(clientHelloMsg.RawData))
	if err != nil {
		log.Warning(fmt.Sprintf("Write client hello fail: %s", err.Error()))
	}

	i++ ;fmt.Println(i)
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