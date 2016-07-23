package sproxy


import "log/syslog"

var log *syslog.Writer
var conf *config

func SetLogger(l *syslog.Writer)  {
	log = l
}
func SetConfig(config *config)  {
	conf = config
}

