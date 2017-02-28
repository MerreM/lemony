package ui

import (
	"fmt"

	"github.com/MerreM/lemony/chatroom/punchy"
	"github.com/jroimartin/gocui"
	"github.com/op/go-logging"
)

const (
	WELCOME_VIEW = "welcome-view"
	CONSOLE_VIEW = "console-box"
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

	//	err = manager.chatBoxLayout(g)
	//	if err != nil {
	//		return err
	//	}
	//	err = manager.inputLayout(g)
	//	if err != nil {
	//		return err
	//	}

	return nil
}

func (manager *MultiRoomChatboxManager) welcomeScreenLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("welcome-box:commands", 0, 0, ((maxX/10)*2)-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "[Commands]"
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = true
	}
	if v, err := g.SetView("welcome-box:welcome", ((maxX / 10) * 2), 0, maxX-1, ((maxY/10)*8)-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "[Welcome]"
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = true
		fmt.Fprintf(v, "\n\n  Hello and welcome to the Lemony client!\n")

		fmt.Fprintf(v, "\n\n  Available commands are listed to the right, and entered below")

		fmt.Fprintf(v, "\n\n  Press Ctrl + D to see the console.")
	}
	if v, err := g.SetView("welcome-box:input", ((maxX / 10) * 2), ((maxY / 10) * 8), maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if _, err = setCurrentViewOnTop(g, "welcome-box:input"); err != nil {
			return err
		}
		v.Title = "[Input]"
		v.Editable = true
		v.Wrap = true
		v.Autoscroll = true
	}
	return nil
}

func (manager *MultiRoomChatboxManager) toggleDebug(g *gocui.Gui, v *gocui.View) error {
	if !manager.DebugActive {
		_, err := g.SetViewOnTop(CONSOLE_VIEW)
		g.Cursor = false
		if err != nil {
			return err
		}
		manager.DebugActive = true
	} else if manager.ActiveRoom == WELCOME_VIEW {
		if _, err := g.SetViewOnTop("welcome-box:input"); err != nil {
			return err
		}
		if _, err := g.SetViewOnTop("welcome-box:welcome"); err != nil {
			return err
		}
		if _, err := g.SetViewOnTop("welcome-box:commands"); err != nil {
			return err
		}
		manager.DebugActive = false
		g.Cursor = true
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
		logging.SetBackend(formattedUiBackend)
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
		make([]RoomView, 20),
		WELCOME_VIEW,
		make(chan string),
		false}

}

type RoomView struct {
	gocui.View
	RoomName              string
	UserList              []punchy.Peer
	messageSendChannel    chan string
	messageReceiveChannel chan string
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
	g.Mouse = true

	manager := initMultiRoomMananger()
	g.SetManager(manager)

	log.Info("Manager set")

	// Global Keys - Debug/Quit/Etc

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Critical(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, manager.toggleDebug); err != nil {
		log.Critical(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Critical(err)
	}
}
