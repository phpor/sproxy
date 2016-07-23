package sproxy

import (
	"fmt"
	"github.com/astaxie/beego/config/yaml"
)

type addr struct {
	host string
	port int64
}

func newAddr(host string, port int64) *addr {
	return &addr{host, port}
}
func (l *addr) String() string {
	return fmt.Sprintf("%s:%d", l.host, l.port);
}

type config struct {
	timeout map[string]int64
	dnsResolver []string
	listenner map[string][]*addr
	whitelist []*addr
}

func NewConfig(conf_file string) *config {
	conf_data, err := yaml.ReadYmlReader(conf_file)
	if err != nil {
		panic("load config file " + conf_file + " fail: " + err.Error())
	}

	fmt.Printf("%v\n\n", conf_data)
	c := &config{
		timeout: make(map[string]int64),
		dnsResolver: nil,
		listenner: make(map[string][]*addr),
		whitelist: nil,
	}
	c.parseListener(conf_data)
	c.parseTimeout(conf_data)
	c.parseDnsResolver(conf_data)
	c.parseWhitelist(conf_data)
	return c
}
func (c *config) GetListener(alias string) []*addr {
	return c.listenner[alias]
}
func (c *config) GetDnsResolver() []string {
	return c.dnsResolver
}
func (c *config) GetTimeout(alias string) int64 {
	return c.timeout[alias]
}
func (c *config) GetWhitelist() []*addr {
	return c.whitelist
}
func (c *config) GetBackend(host string) *addr  {
	for _, addr := range c.whitelist {
		if addr.host == host {
			return addr
		}
	}
	return nil
}

func (c *config) parseListener(data map[string]interface{}) {
	if v, exists := data["listen"]; exists {
		fmt.Printf("%v\n", v)
		item := v.(map[string]interface{})
		for k, _ := range item {
			c.listenner[k] = parseAddr(item, k)
		}
	}
}
func (c *config) parseDnsResolver(data map[string]interface{}){
	if v, exists := data["dnsresolver"]; exists {
		for _, v := range v.([]interface{}) {
			c.dnsResolver = append(c.dnsResolver, v.(string))
		}
	}
}
func (c *config) parseTimeout(data map[string]interface{}) {
	if v, exists := data["timeout"]; exists {
		for k, v := range v.(map[string]interface{}) {
			c.timeout[k] = v.(int64)
		}
	}
}
func (c *config) parseWhitelist(data map[string]interface{})  {
	c.whitelist = parseAddr(data, "whitelist")
}
func parseAddr(data map[string]interface{}, alias string) (res []*addr) {
	if v, exists := data[alias]; exists {
		for _, listener := range v.([]interface{}) {
			for host, port := range listener.(map[string]interface{}) {
				res = append(res, newAddr(host, port.(int64)))
			}
		}
	}
	return
}
