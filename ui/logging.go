package ui

import (
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("ui")

func init() {
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	formatter := logging.NewBackendFormatter(backend1, format)

	logging.SetBackend(formatter)
}
