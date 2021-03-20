package internal

import (
	golog "github.com/op/go-logging"
	"os"
)

var log = initLogger()

func initLogger() *golog.Logger {
	format := golog.MustStringFormatter(`%{time:15:04:05.000} [%{level:.4s}] ▶ %{color:reset}%{message}`)
	logger := golog.MustGetLogger("testdrive")
	backend := golog.NewLogBackend(os.Stderr, "", 0)
	formattedBackend := golog.NewBackendFormatter(backend, format)
	leveledFormattedBackend := golog.AddModuleLevel(formattedBackend)
	leveledFormattedBackend.SetLevel(golog.DEBUG, "")
	logger.SetBackend(leveledFormattedBackend)
	return logger
}
