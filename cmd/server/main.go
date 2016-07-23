package main

import (
	"flag"
	"log/syslog"
	"github.com/phpor/sproxy"
	"time"
)

func main() {
	conf_file := flag.String("c", "/Users/phpor/workspace/go/src/github.com/phpor/sproxy/conf/sproxy.yaml", "path to config file")
	flag.Parse()
	log, err := syslog.New(syslog.LOG_DEBUG|syslog.LOG_LOCAL0, "sproxy")
	if err != nil {
		panic("init logger fail")
	}
	sproxy.SetLogger(log)

	if err != nil {
		log.Err(err.Error())
	}
	conf := sproxy.NewConfig(*conf_file)
	sproxy.SetConfig(conf)

	for _, l := range conf.GetListener("http_proxy") {
		log.Debug("start to listen " + l.String())
		go sproxy.ServeHttpProxy(l)
	}
	for _, l := range conf.GetListener("sni_proxy") {
		log.Debug("start to listen " + l.String())
		go sproxy.ServeSniProxy(l)
	}
	for _, l := range conf.GetListener("http_tunnel") {
		log.Debug("start to listen " + l.String())
		go sproxy.ServeHttpTunnelProxy(l)
	}
	for{
		time.Sleep(time.Second * 10)
	}
}



