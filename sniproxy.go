package sproxy

import (
	"fmt"
	"net"
	"github.com/phpor/sproxy/tls"
	"io"
)

func ServeSniProxy(l *addr)  {
	ln, err := net.Listen("tcp4", l.String())

	if err != nil {
		panic("error listening on tcp port "+ l.String() + ":" + err.Error())
	}

	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Err("Error accepting new connection:  " + err.Error())
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			//set timeout to connection
			clientHelloMsg, err := tls.ReadClientHello(c)
			if err != nil {
				log.Err("Error read client hello:  " + err.Error())
				return
			}
			if clientHelloMsg.ServerName == "" {
				log.Warning("Client has no sni: "+ c.LocalAddr().String() +" <= "+ c.RemoteAddr().String())
				return
			}
			hostname := clientHelloMsg.ServerName
			if ! conf.IsAccessAllow(hostname) {
				log.Warning(fmt.Sprintf("Access %s is not allow", hostname))
				return
			}
			dst := hostname + ":443"
			//dst := "220.181.112.244:443"
			conn, err := net.Dial("tcp", dst)
			if err != nil {
				log.Warning(fmt.Sprintf("connect %s fail\n", dst))
				return
			}
			log.Debug(fmt.Sprintf("connected to %s\n", dst))
			defer conn.Close()

			_, err = io.WriteString(conn, string(clientHelloMsg.RawData))
			if err != nil {
				log.Warning(fmt.Sprintf("Write client hello fail: %s", err.Error()))
			}

			go func() {
				_ ,err := io.Copy(c, conn)
				if err != nil {
					log.Warning(fmt.Sprintf("2: error: %s\n", err.Error()))
				}
			}()
			go func() {
				_, err := io.Copy(conn, c)
				if err != nil {
					log.Warning(fmt.Sprintf("3: error: %s\n", err.Error()))
				}
			}()
		}(c)
	}
}
