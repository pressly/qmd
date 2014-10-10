package qmd

import (
	"fmt"
	"os"

	stdlog "log"

	"github.com/op/go-logging"
	"github.com/pressly/gohttpware/errorlog"
)

var lg = logging.MustGetLogger("qmd")

type LoggingConfig struct {
	LogLevel    string   `toml:"log_level"`
	LogBackends []string `toml:"log_backends"`
	AirbrakeKey string   `toml:"airbrake_key"`
}

func (lc *LoggingConfig) Clean() {
	if lc.LogLevel == "" {
		lc.LogLevel = "INFO"
	}
	if len(lc.LogBackends) == 0 {
		lc.LogBackends = append(lc.LogBackends, "STDOUT")
	}
}

func SetupLogging(lc *LoggingConfig) error {
	// Setup logger
	logging.SetFormatter(logging.MustStringFormatter("%{level} %{message}"))

	var logBackends []logging.Backend
	for _, lb := range lc.LogBackends {
		// TODO: test for starting with / or ./ and treat it
		// as a file logger
		// TODO: case insensitive stdout / syslog
		switch lb {
		case "STDOUT":
			logBackend := logging.NewLogBackend(os.Stdout, "", stdlog.LstdFlags)
			logBackends = append(logBackends, logBackend)
		case "syslog":
			logBackend, err := logging.NewSyslogBackend("qmd")
			if err != nil {
				return err
			}
			logBackends = append(logBackends, logBackend)
		case "Airbrake":
			if lc.AirbrakeKey != "" {
				logBackend := errorlog.NewAirbrakeBackend(lc.AirbrakeKey)
				logBackends = append(logBackends, logBackend)
			} else {
				return fmt.Errorf("Couldn't find Airbrake API Key")
			}
		}
	}
	if len(logBackends) > 0 {
		logging.SetBackend(logBackends...)
	}

	logLevel, err := logging.LogLevel(lc.LogLevel)
	if err != nil {
		return err
	}
	logging.SetLevel(logLevel, "qmd")

	// Redirect standard logger
	stdlog.SetOutput(&logProxyWriter{})

	return nil
}

type logProxyWriter struct{}

func (l *logProxyWriter) Write(p []byte) (n int, err error) {
	lg.Info("%s", p)
	return len(p), nil
}
