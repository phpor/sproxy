package main

import (
	"flag"
	"github.com/phpor/sproxy"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	conf_file_path := os.Getenv("SPROXY_CONF_PATH")
	if conf_file_path == "" {
		conf_file_path = "/etc/sproxy/sproxy.yaml"
	}
	conf_file := flag.String("c", conf_file_path, "path to config file")
	flag.Parse()
	log, err := sproxy.NewLogger()
	if err != nil {
		panic(err)
	}
	sproxy.SetLogger(log)

	conf := sproxy.NewConfig(*conf_file)
	sproxy.SetConfig(conf)

	var wg sync.WaitGroup
	for _, l := range conf.GetListener("http_proxy") {
		log.Debug("start to listen " + l)
		wg.Add(1)
		go func() {
			err := sproxy.ServeTcp(l, sproxy.ServeHttp)
			if err != nil {
				log.Err(err.Error())
			}
			wg.Done()
		}()
	}
	for _, httpConf := range conf.Http {
		log.Debug("start to listen " + httpConf.Addr)
		wg.Add(1)
		go func() {
			err := sproxy.ServeTcp(httpConf.Addr, func(downstream net.Conn) error {
				return sproxy.ServeProxyInProxy(downstream, "http", httpConf)
			})
			if err != nil {
				log.Err(err.Error())
			}
			wg.Done()
		}()
	}
	for _, httpsConf := range conf.Https {
		log.Debug("start to listen " + httpsConf.Addr)
		wg.Add(1)
		go func() {
			err := sproxy.ServeTls(httpsConf.Addr, httpsConf.Cert, httpsConf.Key, func(downstream net.Conn) error {
				return sproxy.ServeProxyInProxy(downstream, "https", httpsConf)
			})
			if err != nil {
				log.Err(err.Error())
			}
			wg.Done()
		}()
	}
	for _, l := range conf.GetListener("sni_proxy") {
		log.Debug("start to listen " + l)
		wg.Add(1)
		go func() {
			err := sproxy.ServeTcp(l, sproxy.ServeSniProxy)
			if err != nil {
				log.Err(err.Error())
			}
			wg.Done()
		}()
	}

	log.Debug("Started")
	c := make(chan os.Signal, 10)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
	<-c
	sproxy.Stats.Stopping = true
	wg.Wait()
	log.Info("Stopped")
}
