package main

import (
	"flag"

	"os"

	"github.com/MerreM/lemony/chatroom/punchy"
	"github.com/MerreM/lemony/ui"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("ui")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func initLogging() {
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.INFO, "")

	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter)

	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")
}

func main() {

	serverPort := flag.Int("s", 0, "Listen mode. Specify port")
	clientConnect := flag.Int("c", 0, "Send mode. Specify port")
	flag.Parse()
	if serverPort != nil && *serverPort != 0 {
		server := punchy.NewServer(serverPort)
		server.Serve()
		return
	} else if clientConnect != nil && *clientConnect != 0 {
		client := punchy.NewClient("localhost", clientConnect)
		ui.InitUi(client)
		return
	}
	flag.Usage()
}
