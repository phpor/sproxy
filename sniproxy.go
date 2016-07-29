package sproxy

import (
	"fmt"
	"net"
	"github.com/phpor/sproxy/tls"
	"io"
)


func ServeSniProxy(downstream net.Conn) error {
	defer func() {
		downstream.Close()
		Stats.CurrentTaskNum--
	}()
	//set timeout to connection
	clientHelloMsg, err := tls.ReadClientHello(downstream)
	if err != nil {
		log.Err("Error read client hello:  " + err.Error())
		return err
	}
	if clientHelloMsg.ServerName == "" {
		log.Warning("Client has no sni: "+ downstream.LocalAddr().String() +" <= "+ downstream.RemoteAddr().String())
		return err
	}
	hostname := clientHelloMsg.ServerName
	upstream, err := createUpstream(hostname, downstream)
	if err != nil {
		return err
	}
	defer upstream.Close()

	_, err = io.WriteString(upstream, string(clientHelloMsg.RawData))
	if err != nil {
		log.Warning(fmt.Sprintf("Write client hello fail: %s", err.Error()))
		return err
	}

	ioCopy(downstream, upstream)
	return nil
}