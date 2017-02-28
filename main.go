package main

import (
	"flag"

	"github.com/MerreM/lemony/chatroom/punchy"
	"github.com/MerreM/lemony/ui"
)

func main() {

	serverPort := flag.Int("l", 0, "Listen mode. Specify port")
	//	clientConnect := flag.Int("c", 0, "Send mode. Specify port")
	clientMode := flag.Bool("c", true, "Client mode.")
	flag.Parse()
	if serverPort != nil && *serverPort != 0 {
		server := punchy.NewServer(serverPort)
		server.Serve()
		return
	} else if clientMode != nil && *clientMode == true {
		//		client := punchy.NewClient("localhost", clientConnect)
		//		ui.InitUi(client)
		ui.InitMultiRoomUi()
		return
	}
	flag.Usage()
}
