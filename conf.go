package sproxy

import (
	"fmt"
	"github.com/astaxie/beego/config/yaml"
	"strings"
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
	whitelist []string
}

func NewConfig(conf_file string) *config {
	conf_data, err := yaml.ReadYmlReader(conf_file)
	if err != nil {
		panic("load config file " + conf_file + " fail: " + err.Error())
	}

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
func (c *config) GetWhitelist() []string {
	return c.whitelist
}
func (c *config) IsAccessAllow(hostport string)bool  {
	for _, addr := range c.whitelist {
		if addr == hostport {
			return true
		}
	}
	return false
}
func (c *config) GetBackend(host string) string  {
	for _, addr := range c.whitelist {
		if strings.Split(addr, ":")[0] == host {
			return addr
		}
	}
	return ""
}

func (c *config) parseListener(data map[string]interface{}) {
	if v, exists := data["listen"]; exists {
		item := v.(map[string]interface{})
		for k, _ := range item {
			c.listenner[k] = parseAddr(item, k)
		}
	}
}
func (c *config) parseDnsResolver(data map[string]interface{}){
	c.dnsResolver = parseStrSlice(data, "dnsresolver")
}
func (c *config) parseTimeout(data map[string]interface{}) {
	if v, exists := data["timeout"]; exists {
		for k, v := range v.(map[string]interface{}) {
			c.timeout[k] = v.(int64)
		}
	}
}
func (c *config) parseWhitelist(data map[string]interface{})  {
	c.whitelist = parseStrSlice(data, "whitelist")
}
func parseStrSlice(data map[string]interface{}, alias string) (res []string) {
	if v, exists := data[alias]; exists {
		for _, v := range v.([]interface{}) {
			res = append(res, v.(string))
		}
	}
	return
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
