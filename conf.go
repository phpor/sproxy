package sproxy

import (
	"github.com/astaxie/beego/config/yaml"
	"strings"
	"path/filepath"
)


type https struct  {
	Addr string
	Cert string
	Key string
}

type config struct {
	timeout map[string]int64
	dnsResolver []string
	listenner map[string][]string
	whitelist []string
	Https map[string]https
	confPath string
}

func NewConfig(conf_file string) *config {
	conf_data, err := yaml.ReadYmlReader(conf_file)
	if err != nil {
		panic("load config file " + conf_file + " fail: " + err.Error())
	}

	c := &config{
		timeout: make(map[string]int64),
		dnsResolver: nil,
		listenner: make(map[string][]string),
		whitelist: nil,
	}
	c.confPath = conf_file
	c.parseListener(conf_data)
	c.parseTimeout(conf_data)
	c.parseDnsResolver(conf_data)
	c.parseWhitelist(conf_data)
	c.parseHttps(conf_data)
	return c
}

func (c * config) parseHttps(data map[string]interface{}) {
	res := map[string]https{}
	confBaseDir := filepath.Dir(c.confPath)
	if _https, exists := data["https"]; exists {
		for group, item := range _https.(map[string]interface{}) {
			_item := item.(map[string]interface{})
			var h https
			h.Addr = _item["addr"].(string)
			h.Cert = _item["cert"].(string)
			h.Key = _item["key"].(string)
			if !filepath.IsAbs(h.Cert) {
				h.Cert = confBaseDir + "/" + h.Cert
			}
			if !filepath.IsAbs(h.Key) {
				h.Key = confBaseDir + "/" + h.Key
			}
			res[group] = h
		}
	}
	c.Https = res
}
func (c *config) GetListener(alias string) []string {
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
			c.listenner[k] = parseStrSlice(item, k)
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
	v, exists := data[alias]
	if !exists {
		return
	}
	if _,ok := v.([]interface{}); !ok {
		return
	}

	for _, v := range v.([]interface{}) {
		res = append(res, v.(string))
	}
	return
}
