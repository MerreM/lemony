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

func GetFileBackend() logging.Backend {
	f, err := os.OpenFile("output.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	file_backend := logging.NewLogBackend(f, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	formatted_file := logging.NewBackendFormatter(file_backend, format)
	return formatted_file
}
