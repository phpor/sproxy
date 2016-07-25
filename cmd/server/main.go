package main

import (
	"flag"
	"log/syslog"
	"github.com/phpor/sproxy"
	"os"
	"syscall"
	"os/signal"
	"sync"
)

func main() {
	conf_file := flag.String("c", "/Users/phpor/workspace/go/src/github.com/phpor/sproxy/conf/sproxy.yaml", "path to config file")
	flag.Parse()
	log, err := syslog.New(syslog.LOG_ERR|syslog.LOG_INFO|syslog.LOG_DEBUG|syslog.LOG_LOCAL0, "sproxy")
	if err != nil {
		panic("init logger fail")
	}
	sproxy.SetLogger(log)

	if err != nil {
		log.Err(err.Error())
	}
	conf := sproxy.NewConfig(*conf_file)
	sproxy.SetConfig(conf)

	var wg sync.WaitGroup
	for _, l := range conf.GetListener("http_proxy") {
		log.Debug("start to listen " + l.String())
		wg.Add(1)
		go func() {
			sproxy.ServeTcp(l, sproxy.ServeHttpProxy)
			wg.Done()
		}()
	}
	for _, l := range conf.GetListener("sni_proxy") {
		log.Debug("start to listen " + l.String())
		wg.Add(1)
		go func() {
			sproxy.ServeTcp(l, sproxy.ServeSniProxy)
			wg.Done()
		}()
	}
	for _, l := range conf.GetListener("http_tunnel") {
		log.Debug("start to listen " + l.String())
		wg.Add(1)
		go func() {
			sproxy.ServeTcp(l, sproxy.ServeHttpTunnelProxy)
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



