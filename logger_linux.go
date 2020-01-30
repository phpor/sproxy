package sproxy

import "log/syslog"

func NewLogger() (Logger, error) {
	log, err := syslog.New(syslog.LOG_ERR|syslog.LOG_INFO|syslog.LOG_DEBUG|syslog.LOG_LOCAL0, "sproxy")
	if err != nil {
		log, err = syslog.Dial("tcp", "localhost:514", syslog.LOG_ERR|syslog.LOG_INFO|syslog.LOG_DEBUG|syslog.LOG_LOCAL0, "sproxy")
	}
	return log, err
}
