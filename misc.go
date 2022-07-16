package sproxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/phpor/sproxy/proxy"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var ErrAccessForbidden = errors.New("Access deny")
var ErrReadDownStream = errors.New("Read Downstream fail")

var conf *config

func SetLogger(l Logger) {
	log = l
}
func SetConfig(config *config) {
	conf = config
}

func ServeTcp(l string, handler func(net.Conn) error) error {
	ln, err := net.Listen("tcp4", l)

	if err != nil {
		panic("error listening on tcp port " + l + ":" + err.Error())
	}
	return ServeConn(ln, handler)
}

func ServeTls(l string, cert string, key string, handler func(net.Conn) error) error {

	cer, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		panic("load cert or key fail: " + err.Error())
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, erl := tls.Listen("tcp", l, config)
	if erl != nil {
		panic("error listening on tcp port " + l + err.Error())
	}
	return ServeConn(ln, handler)
}
func ServeConn(ln net.Listener, handler func(net.Conn) error) error {

	defer ln.Close()

	s := make(chan os.Signal, 10)
	signal.Notify(s, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		<-s
		ln.Close() //如果这个在main中open，则可以只在main中处理信号了，就不需要在每个listenner中处理信号了
	}()
	var wg sync.WaitGroup       //确保每个层级的goroutine都能等子goroutine退出后自己才退出，才能保证不会中断未处理完成的请求
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		downstream, err := ln.Accept() //这里不能直接处理信号,而是要在其它协程中接收到信号后，直接把ln给关掉，这里立刻就会返回失败
		if err != nil {                // 如何判断这个错误其实是ln close导致的？
			if Stats.Stopping {
				log.Err("Stopped listenner " + ln.Addr().String())
				break
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Err(fmt.Sprintf("http: Accept error: %v; retrying in %v", err, tempDelay))
				time.Sleep(tempDelay)
				continue
			}

			log.Err("Error accepting new connection:  " + err.Error())
			return err
		}
		Stats.CurrentTaskNum++
		wg.Add(1)
		go func() {
			//设置超时
			s := time.Now()
			handler(NewTimeoutConn(downstream, time.Duration(conf.GetTimeout("client_read"))*time.Millisecond))
			e := time.Now()
			log.Err(fmt.Sprintf("%s  client %s time use %d ms", ln.Addr().String(), downstream.RemoteAddr().String(), e.Sub(s).Nanoseconds()/1000000))
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}
func readLine(conn net.Conn) ([]byte, error) {
	var buf = make([]byte, 1)
	var slice []byte
	var err error
	for {
		_, err = conn.Read(buf)
		if err != nil {
			break
		}
		if buf[0] == '\n' {
			return slice[:len(slice)-1], nil
		}
		slice = append(slice, buf...)
	}
	return nil, err
}
func ServeProxyInProxy(downstream net.Conn, proxyType string, config interface{}) error {
	backend := ""
	if proxyType == "https" {
		backend = config.(HttpsConf).Backend
	}
	if proxyType == "http" {
		backend = config.(HttpConf).Backend
	}
	if backend != "" {
		return httpProxyInHttpProxy(downstream, backend)
	}
	return ServeHttp(downstream)
}
func ServeHttp(downstream net.Conn) error {
	firstLine, err := readLine(downstream)
	if err != nil {
		log.Warning("read header line fail: " + err.Error())
		downstream.Close()
		return ErrReadDownStream
	}
	arr := strings.Split(string(firstLine), " ")
	method := arr[0]
	if len(arr) < 3 {
		log.Err("client Protocol bad")
		return fmt.Errorf("client protocol bad")
	}
	if method == "CONNECT" {
		return serveHttpTunnelProxy(downstream, string(firstLine))
	}

	return serveHttpProxy(downstream, string(firstLine))
}

func createUpstream(hostname string, downstream net.Conn) (net.Conn, error) {
	port := ""
	target := hostname
	arrTarget := strings.Split(target, ":")
	hostname = arrTarget[0]
	if len(arrTarget) == 2 {
		port = arrTarget[1]
	}

	if port == "" {
		port = strings.Split(downstream.LocalAddr().String(), ":")[1]
	}

	if conf.whitelist != nil { //当白名单为空时全部允许
		if !conf.IsAccessAllow(hostname, port) {
			log.Warning(ErrAccessForbidden.Error() + ": " + hostname)
			return nil, ErrAccessForbidden
		}
	}
	remote_dns := os.Getenv("REMOTE_DNS")
	dst := ""
	if remote_dns != "on" {
		hostip, err := nslookup(hostname)
		if err != nil {
			log.Warning("Nslookup fail: " + hostname)
			return nil, err
		}

		log.Debug(fmt.Sprintf("access %s(%s):%s", hostname, hostip, port))
		dst = fmt.Sprintf("%s:%s", hostip, port)
	} else {
		log.Debug(fmt.Sprintf("access %s:%s", hostname, port))
		dst = fmt.Sprintf("%s:%s", hostname, port)
	}
	//todo: 判断目标主机是本机的话就不给代理
	// Note: 目前这个proxy只能支持socket5代理，其它代理自动忽略，但是代码上允许扩展其它代理出来；设置方法; export all_proxy=socks5://127.0.0.1:1080
	upstream, err := proxy.FromEnvironment(NewTimeoutDailer(time.Duration(conf.GetTimeout("upstream_conn"))*time.Millisecond)).Dial("tcp", dst)

	/*
		upstream, err := net.Dial("tcp", dst)
	*/
	if err != nil {
		log.Warning(fmt.Sprintf("connect %s fail", dst))
		return nil, err
	}
	return NewTimeoutConn(upstream, time.Duration(conf.GetTimeout("upstream_read"))*time.Millisecond), nil
}
func ioCopy(downstream, upstream net.Conn) (int64, int64) {
	var len_up int64
	//io.Copy 中有两个判断分支，其实这里的conn确实有ReadFrom和WriteTo方法的，不需要走下面的循环
	go func() {
		len_up, _ = io.Copy(upstream, downstream) //copy to upstream
		// 下面之所以敢关闭上下游连接，基于一个假设：client端只有收到全部响应之后才会关闭连接（一般是这样的）
		downstream.Close() //这里的Close主要是为了让另外一个方向的copy立即退出， 重复Close不会导致错误
		upstream.Close()
	}()
	len_down, _ := io.Copy(downstream, upstream) //copy to downstream
	return len_down, len_up
}
