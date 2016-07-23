package sproxy

import "testing"

func TestNewConfig(t *testing.T) {
	conf := NewConfig("/Users/phpor/workspace/go/src/github.com/phpor/sproxy/conf/sproxy.yaml")
	t.Log(conf)
}
