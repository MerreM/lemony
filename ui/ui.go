// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/MerreM/lemony/chatroom/punchy"
	"github.com/jroimartin/gocui"
	"github.com/op/go-logging"
)

var debugView = false

var (
	viewArr = []string{"input-box", "send-button"}
	active  = 0
)

type ChatboxManager struct {
	chatroomClient *punchy.Client
	room           string
	input          chan string
	output         chan string
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(viewArr)
	name := viewArr[nextIndex]

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}

	active = nextIndex
	return nil
}

func (manager *ChatboxManager) processMessage(g *gocui.Gui, v *gocui.View) error {
	g.Execute(func(g *gocui.Gui) error {
		inputBox, err := g.View("input-box")
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(inputBox)
		data_str := string(data)
		data_str = strings.TrimSuffix(data_str, "\n")
		data_str = strings.TrimSuffix(data_str, " ")
		if strings.HasPrefix(data_str, "/") {
			log.Info("Handle Command")
		} else {
			log.Info("Send message to room")
			manager.output <- data_str
			manager.input <- fmt.Sprintf("You said \"%s\" to room ??", data_str)
		}
		inputBox.Clear()
		inputBox.Rewind()
		inputBox.SetCursor(0, 0)
		return nil
	})
	return nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func (manager *ChatboxManager) debugBoxLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("console-box", 0, 0, maxX-1, (maxY/10)*8); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		uibackend := logging.NewLogBackend(v, "UI", 0)
		var format = logging.MustStringFormatter(
			`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
		)
		formattedUiBackend := logging.NewBackendFormatter(uibackend, format)
		logging.SetBackend(formattedUiBackend)
		v.Title = "Console Room"
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = true

	}
	return nil
}

func (manager *ChatboxManager) chatBoxLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("chat-box", 0, 0, maxX-1, (maxY/10)*8); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Chat Room"
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = true
	}
	return nil
}

func (manager *ChatboxManager) inputLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("input-box", 0, ((maxY/10)*8)+1, ((maxX/10)*9)-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Input"
		v.Editable = true
		if _, err = setCurrentViewOnTop(g, "input-box"); err != nil {
			return err
		}
		if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
			log.Critical(err)
		}
	}
	if v, err := g.SetView("send-button", ((maxX/10)*9)+1, ((maxY/10)*8)+1, maxX-1, maxY-1); err != nil {
		v.Editable = false
		fmt.Fprintln(v, "Send Message")
		if err != gocui.ErrUnknownView {
			return err
		}
		if err := g.SetKeybinding("send-button", gocui.KeyEnter, gocui.ModNone, manager.processMessage); err != nil {
			log.Critical(err)
		}
	}
	return nil
}

func (manager *ChatboxManager) Layout(g *gocui.Gui) error {
	err := manager.debugBoxLayout(g)
	if err != nil {
		return err
	}

	err = manager.chatBoxLayout(g)
	if err != nil {
		return err
	}
	err = manager.inputLayout(g)
	if err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func initChatRoomManager(chatroomClient *punchy.Client) *ChatboxManager {
	input := make(chan string)
	output := make(chan string)

	return &ChatboxManager{chatroomClient, "", input, output}
}

func (manager *ChatboxManager) updateChatMessages(g *gocui.Gui) {
	for {
		message, more := <-manager.input
		if more {
			g.Execute(func(g *gocui.Gui) error {
				v, err := g.View("chat-box")
				v.Write([]byte(message + "\n"))
				if err != nil {
					log.Error(err)
					return err
				}
				return nil

			})
		} else {
			return
		}
	}
}

func toggleDebug(g *gocui.Gui, v *gocui.View) error {
	if !debugView {
		_, err := g.SetViewOnTop("console-box")
		if err != nil {
			return err
		}
		debugView = true
	} else {
		_, err := g.SetViewOnTop("chat-box")
		if err != nil {
			return err
		}
		debugView = false
	}
	return nil
}

func InitUi(chatroomClient *punchy.Client) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Critical(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	manager := initChatRoomManager(chatroomClient)
	log.Info("Startup")
	chatroomClient.StartUp(manager.input)
	log.Info("Connecting")
	go chatroomClient.ConnectToRoom(manager.output, "Hello")
	log.Info("Manager setting")
	g.SetManager(manager)

	go manager.updateChatMessages(g)

	log.Info("Manager set")

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Critical(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, toggleDebug); err != nil {
		log.Critical(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Critical(err)
	}
}
