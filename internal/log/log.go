package log

import (
	golog "github.com/op/go-logging"
	"os"
)

var logger = initLogger()
var leveledFormattedBackend golog.LeveledBackend

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

func SetVerbose(verbose bool) {
	if verbose {
		leveledFormattedBackend.SetLevel(golog.DEBUG, "")
	} else {
		leveledFormattedBackend.SetLevel(golog.INFO, "")
	}
}

func initLogger() *golog.Logger {
	format := golog.MustStringFormatter(`%{time:15:04:05.000} [%{level:.4s}] ▶ %{color:reset}%{message}`)
	logger := golog.MustGetLogger("testdrive")
	backend := golog.NewLogBackend(os.Stderr, "", 0)
	formattedBackend := golog.NewBackendFormatter(backend, format)
	leveledFormattedBackend = golog.AddModuleLevel(formattedBackend)
	leveledFormattedBackend.SetLevel(golog.INFO, "")
	logger.SetBackend(leveledFormattedBackend)
	return logger
}
