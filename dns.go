package sproxy

import (
	"github.com/phpor/godns"
	"fmt"
)

func nslookup(hostname string) (string, error) {

	option := &godns.LookupOptions{
		DNSServers: conf.GetDnsResolver(),
		OnlyIPv4: true,
		UseHosts: false,
	}
	addr, err := godns.LookupHost(hostname, option)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	//fmt.Printf("%v", addr)
	return addr[0], nil
}
