package sproxy

import (
	"fmt"
	"github.com/phpor/sproxy/tls"
	"io"
	"net"
)

func ServeSniProxy(downstream net.Conn) error {
	defer func() {
		_ = downstream.Close()
		Stats.CurrentTaskNum--
	}()
	//set timeout to connection
	clientHelloMsg, err := tls.ReadClientHello(downstream)
	if err != nil {
		_ = log.Err("Error read client hello:  " + err.Error())
		return err
	}
	if clientHelloMsg.ServerName == "" {
		_ = log.Warning("Client has no sni: " + downstream.LocalAddr().String() + " <= " + downstream.RemoteAddr().String())
		return err
	}
	hostname := clientHelloMsg.ServerName
	upstream, err := createUpstream(hostname, downstream)
	if err != nil {
		return err
	}
	defer func(upstream net.Conn) {
		_ = upstream.Close()
	}(upstream)

	_, err = io.WriteString(upstream, string(clientHelloMsg.RawData))
	if err != nil {
		_ = log.Warning(fmt.Sprintf("Write client hello fail: %s", err.Error()))
		return err
	}

	ioCopy(downstream, upstream)
	return nil
}
