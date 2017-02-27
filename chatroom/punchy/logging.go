package punchy

import (
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("punchy")

func init() {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	formatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(formatter)
}

func initServerLogging() {
	f, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	file_backend := logging.NewLogBackend(f, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	formatter := logging.NewBackendFormatter(backend, format)
	formatted_file_backend := logging.NewBackendFormatter(file_backend, format)

	logging.SetBackend(formatter, formatted_file_backend)
}
