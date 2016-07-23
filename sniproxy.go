package sproxy

import (
	"fmt"
	"net"
	"github.com/phpor/sproxy/tls"
	"io"
	"os/signal"
	"syscall"
	"os"
	"sync"
)

func ServeSniProxy(l *addr)  {
	ln, err := net.Listen("tcp4", l.String())

	if err != nil {
		panic("error listening on tcp port "+ l.String() + ":" + err.Error())
	}

	defer ln.Close()

	s := make(chan os.Signal, 10)
	signal.Notify(s, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		<-s
		ln.Close() //如果这个在main中open，则可以只在main中处理信号了，就不需要在每个listenner中处理信号了
	}()
	var wg sync.WaitGroup //确保每个层级的goroutine都能等子goroutine退出后自己才退出，才能保证不会中断未处理完成的请求
	for {
		c, err := ln.Accept() //这里不能直接处理信号,而是要在其它协程中接收到信号后，直接把ln给关掉，这里立刻就会返回失败
		if err != nil {
			if Stats.Stopping {
				log.Err("Stopped listenner "+l.String())
				break
			}
			log.Err("Error accepting new connection:  " + err.Error())
			break
		}
		Stats.CurrentTaskNum++
		wg.Add(1)
		go func() {
			server(c)
			wg.Done()
		}()
	}
	wg.Wait()
}


func server(c net.Conn)  {
	defer func() {
		c.Close()
		Stats.CurrentTaskNum--
	}()
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
	backendAddr := conf.GetBackend(hostname)
	if backendAddr == nil {
		log.Warning(fmt.Sprintf("Access %s is not allow", hostname))
		return
	}
	hostip, err := nslookup(hostname)
	if err != nil {
		log.Warning("Nslookup fail: " + hostname)
		return
	}
	dst := fmt.Sprintf("%s:%d", hostip, backendAddr.port)
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
	_, err = io.Copy(conn, c)
	if err != nil {
		log.Warning(fmt.Sprintf("3: error: %s\n", err.Error()))
	}
}