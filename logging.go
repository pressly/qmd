package qmd

import (
	stdlog "log"
	"os"

	"github.com/op/go-logging"
)

var (
	lg = logging.MustGetLogger("qmd")
)

type LoggingConf struct {
	LogLevel string `toml:"log_level"`
}

func (lc *LoggingConf) Setup() {
	if lc.LogLevel == "" {
		lc.LogLevel = "INFO"
	}

	logging.SetFormatter(logging.MustStringFormatter("%{level} %{message}"))

	// TODO: we can add more log backend support later.. see older qmd/logging.go
	logging.SetBackend(logging.NewLogBackend(os.Stdout, "", stdlog.LstdFlags))

	logLevel, err := logging.LogLevel(lc.LogLevel)
	if err != nil {
		stdlog.Fatal(err)
		return
	}
	logging.SetLevel(logLevel, "qmd")

	// Redirect standard logger
	stdlog.SetOutput(&logProxyWriter{})
	stdlog.SetFlags(0)
}

// Proxy writer for any packages using the standard log.Println() stuff
type logProxyWriter struct{}

func (l *logProxyWriter) Write(p []byte) (n int, err error) {
	lg.Info("%s", p)
	return len(p), nil
}
