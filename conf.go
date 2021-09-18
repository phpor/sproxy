package sproxy

import (
	"bytes"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type HttpsConf struct {
	Addr    string
	Cert    string
	Key     string
	Backend string
}
type HttpConf struct {
	Addr    string
	Backend string
}

type config struct {
	timeout     map[string]int64
	dnsResolver []string
	listenner   map[string][]string
	whitelist   []string
	Https       map[string]HttpsConf
	Http        map[string]HttpConf
	confPath    string
}

func NewConfig(conf_file string) *config {
	b, err := ioutil.ReadFile(conf_file)
	if err != nil {
		panic("read file " + conf_file + " fail: " + err.Error())
	}
	decoder := yaml.NewDecoder(bytes.NewReader(b))
	conf_data := map[string]interface{}{}
	err = decoder.Decode(conf_data)

	if err != nil {
		panic("load config file " + conf_file + " fail: " + err.Error())
	}

	c := &config{
		timeout:     make(map[string]int64),
		dnsResolver: nil,
		listenner:   make(map[string][]string),
		whitelist:   nil,
	}
	c.confPath = conf_file
	c.parseListener(conf_data)
	c.parseTimeout(conf_data)
	c.parseDnsResolver(conf_data)
	c.parseWhitelist(conf_data)
	c.parseHttpsConf(conf_data)
	c.parseHttpConf(conf_data)
	return c
}

func (c *config) parseHttpConf(data map[string]interface{}) {
	res := map[string]HttpConf{}
	if _http, exists := data["http"]; exists {
		for group, item := range _http.(map[string]interface{}) {
			_item := item.(map[string]interface{})
			var h HttpConf
			h.Addr = _item["addr"].(string)
			if backend, exists := _item["backend"]; exists {
				h.Backend = backend.(string)
			}
			res[group] = h
		}
	}
	c.Http = res
}
func (c *config) parseHttpsConf(data map[string]interface{}) {
	res := map[string]HttpsConf{}
	confBaseDir := filepath.Dir(c.confPath)
	if _https, exists := data["https"]; exists {
		for group, item := range _https.(map[string]interface{}) {
			_item := item.(map[string]interface{})
			var h HttpsConf
			h.Addr = _item["addr"].(string)
			h.Cert = _item["cert"].(string)
			h.Key = _item["key"].(string)
			if !filepath.IsAbs(h.Cert) {
				h.Cert = confBaseDir + "/" + h.Cert
			}
			if !filepath.IsAbs(h.Key) {
				h.Key = confBaseDir + "/" + h.Key
			}
			if backend, exists := _item["backend"]; exists {
				h.Backend = backend.(string)
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
func (c *config) IsAccessAllow(host, port string) bool {
	for _, addr := range c.whitelist {
		if addr == host || addr == host+":"+port {
			return true
		}
	}
	return false
}
func (c *config) GetBackend(host string) string {
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
		for k := range item {
			c.listenner[k] = parseStrSlice(item, k)
		}
	}
}
func (c *config) parseDnsResolver(data map[string]interface{}) {
	c.dnsResolver = parseStrSlice(data, "dnsresolver")
}
func (c *config) parseTimeout(data map[string]interface{}) {
	if v, exists := data["timeout"]; exists {
		for k, _v := range v.(map[string]interface{}) {

			c.timeout[k] = int64(_v.(int))
		}
	}
}
func (c *config) parseWhitelist(data map[string]interface{}) {
	c.whitelist = parseStrSlice(data, "whitelist")
	whitelistfile := parseStr(data, "whitelistfile")
	if whitelistfile == "" {
		return
	}
	_, err := os.Stat(whitelistfile)
	if err != nil {
		panic("stat " + whitelistfile + " fail: " + err.Error())
	}
	content, err := ioutil.ReadFile(whitelistfile)
	if err != nil {
		panic("read file " + whitelistfile + " fail: " + err.Error())
	}
	arr := strings.Split(string(content), "\n")
	for _, line := range arr {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		c.whitelist = append(c.whitelist, line)
	}
}

func parseStr(data map[string]interface{}, alias string) string {
	v, exists := data[alias]
	if !exists {
		return ""
	}
	return v.(string)
}

func parseStrSlice(data map[string]interface{}, alias string) (res []string) {
	v, exists := data[alias]
	if !exists {
		return
	}
	if _, ok := v.([]interface{}); !ok {
		return
	}

	for _, v := range v.([]interface{}) {
		res = append(res, v.(string))
	}
	return
}
