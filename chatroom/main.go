// UDPPuncher project main.go
package chatroom

import (
	"flag"
	"os"

	"github.com/MerreM/UISandbox/chatroom/punchy"
)

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
		client.StartUp()
		client.ConnectToRoom(os.Stdout, "Hello")
		return
	}
	flag.Usage()
}
