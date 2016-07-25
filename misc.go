package sproxy

import (
	"log/syslog"
	"net"
	"sync"
	"syscall"
	"os"
	"os/signal"
	"time"
	"fmt"
)

var log *syslog.Writer
var conf *config

func SetLogger(l *syslog.Writer)  {
	log = l
}
func SetConfig(config *config)  {
	conf = config
}

func ServeTcp(l *addr, handler func(net.Conn))  {
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
			//设置超时
			s := time.Now()
			handler(c)
			e := time.Now()
			log.Err(fmt.Sprintf("%s  client %s time use %d ms",l.String(), c.RemoteAddr().String(), e.Sub(s).Nanoseconds()/1000000))
			wg.Done()
		}()
	}
	wg.Wait()
}