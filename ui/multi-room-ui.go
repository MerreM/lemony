package ui

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/MerreM/lemony/chatroom/punchy"
	"github.com/jroimartin/gocui"
	"github.com/op/go-logging"
)

const (
	WELCOME_VIEW = "welcome-box"
	CONSOLE_VIEW = "console-box"
	INPUT_VIEW   = "input"
	SEND_BUTTON  = "send-button"
)

type MultiRoomChatboxManager struct {
	ChatroomClient *punchy.Client
	ConnectedRooms []RoomView
	ActiveRoom     string
	DebugView      chan string
	DebugActive    bool
}

func (manager *MultiRoomChatboxManager) Layout(g *gocui.Gui) error {
	if err := manager.debugBoxLayout(g); err != nil {
		return err
	}
	if err := manager.welcomeScreenLayout(g); err != nil {
		return err
	}

	if g.CurrentView() == nil {
		g.SetCurrentView(INPUT_VIEW)
	}

	for _, room := range manager.ConnectedRooms {
		activeRoom := room.RoomName == manager.ActiveRoom
		err := room.Layout(g, activeRoom)
		if err != nil {
			return err
		}
	}

	return nil
}

func (manager *MultiRoomChatboxManager) welcomeScreenLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(WELCOME_VIEW+":commands", 0, 0, ((maxX/10)*2)-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "[Commands]"
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = true
	}
	if v, err := g.SetView(WELCOME_VIEW+":welcome", ((maxX / 10) * 2), 0, maxX-1, ((maxY/10)*8)-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "[Welcome]"
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = true
		fmt.Fprintf(v, "\n\n  Hello and welcome to the Lemony client!\n")

		fmt.Fprintf(v, "\n\n  Available commands are listed to the right, and entered below")
		fmt.Fprintf(v, "\n\n  Tab switches the active box. Select send to send a message or command.")
		fmt.Fprintf(v, "\n\n  Press Ctrl + D to see the console.")

	}
	if v, err := g.SetView(INPUT_VIEW, ((maxX / 10) * 2), ((maxY / 10) * 8), int((float32(maxX)/100.0)*95)-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "[Input]"
		v.Editable = true
		v.Wrap = true
		v.Autoscroll = true
	}
	if v, err := g.SetView(SEND_BUTTON, int((float32(maxX)/100.0)*95), ((maxY / 10) * 8), maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		newLineCount := ((maxY - 1) - ((maxY / 10) * 8)) / 2
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = false
		for i := 0; i < newLineCount; i++ {
			fmt.Fprintln(v)
		}
		fmt.Fprintf(v, "SEND")
	}
	return nil
}

func (manager *MultiRoomChatboxManager) toggleDebug(g *gocui.Gui, v *gocui.View) error {
	if !manager.DebugActive {
		_, err := g.SetViewOnTop(CONSOLE_VIEW)
		g.SetCurrentView(CONSOLE_VIEW)
		if err != nil {
			return err
		}
		manager.DebugActive = true
	} else if manager.ActiveRoom == WELCOME_VIEW {
		if _, err := g.SetViewOnTop(INPUT_VIEW); err != nil {
			return err
		}
		if _, err := g.SetViewOnTop(SEND_BUTTON); err != nil {
			return err
		}
		if _, err := g.SetViewOnTop(WELCOME_VIEW + ":welcome"); err != nil {
			return err
		}
		if _, err := g.SetViewOnTop(WELCOME_VIEW + ":commands"); err != nil {
			return err
		}
		g.SetCurrentView(INPUT_VIEW)
		manager.DebugActive = false
	} else {
		g.SetCurrentView(INPUT_VIEW)

	}
	return nil
}

func (manager *MultiRoomChatboxManager) debugBoxLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(CONSOLE_VIEW, 0, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		uibackend := logging.NewLogBackend(v, ">>> ", 0)
		var format = logging.MustStringFormatter(
			`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
		)
		formattedUiBackend := logging.NewBackendFormatter(uibackend, format)
		logging.SetBackend(formattedUiBackend, GetFileBackend())
		v.Title = "[Console]"
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = true
	}
	return nil
}

func initMultiRoomMananger() *MultiRoomChatboxManager {
	return &MultiRoomChatboxManager{
		nil,
		make([]RoomView, 0),
		WELCOME_VIEW,
		make(chan string),
		false}

}

func (manager *MultiRoomChatboxManager) switchInput(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case SEND_BUTTON:
		g.SetCurrentView(INPUT_VIEW)
		g.Cursor = true
		break
	case INPUT_VIEW:
		g.Cursor = false
		g.SetCurrentView(SEND_BUTTON)
		break
	}

	return nil
}

func (manager *MultiRoomChatboxManager) connectToRoom(g *gocui.Gui, roomName string) error {
	for _, room := range manager.ConnectedRooms {
		if room.RoomName == roomName {
			manager.ActiveRoom = roomName
			return nil
		}
	}
	log.Infof("Joining room %v", roomName)
	newRoom := RoomView{roomName, nil, make(chan string), make(chan string)}
	manager.ConnectedRooms = append(manager.ConnectedRooms, newRoom)
	go manager.ChatroomClient.ConnectToRoom(newRoom.messageReceiveChannel, roomName)
	log.Infof("Currently connected to {%v}", manager.ConnectedRooms)
	manager.ActiveRoom = roomName
	return nil
}

func (manager *MultiRoomChatboxManager) handleConnect(g *gocui.Gui, command string) error {
	log.Info("Handling /connect")
	if manager.ChatroomClient != nil {
		log.Warning("Already connected")
	}
	addressString := strings.Replace(command, " ", "", -1)
	log.Infof("Checking address input {%v}", addressString)
	s, err := net.ResolveUDPAddr("udp", addressString)
	if err != nil {
		return err
	}
	client := punchy.NewUIClient(s)
	manager.ChatroomClient = client
	log.Infof("Connecting to {%v}", s)
	return nil
}

func (manager *MultiRoomChatboxManager) handleCommand(g *gocui.Gui, command string) error {
	log.Infof("Command {%v} received.", command)
	if strings.HasPrefix(command, "/connect") {
		return manager.handleConnect(g, strings.Replace(command, "/connect", "", 1))
	} else if strings.HasPrefix(command, "/join") {
		return manager.connectToRoom(g, strings.Replace(command, "/join", "", 1))
	}
	return nil
}

func (manager *MultiRoomChatboxManager) sendMessage(g *gocui.Gui, send *gocui.View) error {
	v, err := g.View(INPUT_VIEW)
	if err != nil {
		return err
	}
	g.Execute(func(g *gocui.Gui) error {
		defer func() {
			v.Clear()
			v.Rewind()
			v.SetCursor(0, 0)
		}()
		data, err := ioutil.ReadAll(v)
		if err != nil {
			return err
		}
		data_str := string(data)
		data_str = strings.TrimSuffix(data_str, "\n")
		data_str = strings.TrimSuffix(data_str, " ")
		if strings.HasPrefix(data_str, "/") {
			log.Info("Handle Command")
			return manager.handleCommand(g, data_str)
		} else {
			log.Info("Send message!")
		}

		return nil
	})
	return nil
}

func InitMultiRoomUi() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Critical(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen
	g.Mouse = false

	manager := initMultiRoomMananger()
	g.SetManager(manager)
	g.SetCurrentView(INPUT_VIEW)

	log.Info("Manager set")

	// Global Keys - Debug/Quit/Etc

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Critical(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, manager.toggleDebug); err != nil {
		log.Critical(err)
	}

	// Input Keys - Tab Shift

	if err := g.SetKeybinding(INPUT_VIEW, gocui.KeyTab, gocui.ModNone, manager.switchInput); err != nil {
		log.Critical(err)
	}
	if err := g.SetKeybinding(SEND_BUTTON, gocui.KeyTab, gocui.ModNone, manager.switchInput); err != nil {
		log.Critical(err)
	}

	if err := g.SetKeybinding(SEND_BUTTON, gocui.KeyEnter, gocui.ModNone, manager.sendMessage); err != nil {
		log.Critical(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Critical(err)
	}
}
