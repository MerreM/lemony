package punchy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type ClientInter interface {
	ConnectToRoom(io.Reader, string)
	ConnectToServer(*net.UDPAddr)
}

type Peer struct {
	net.UDPAddr
	name string
}

type Client struct {
	inputChannel  chan string
	clientChannel chan InboundMessage
	errorChannel  chan error
	middleMan     *net.UDPAddr
	conn          *net.UDPConn
	rooms         map[string][]Peer
}

func (c *Client) errorHandler() {
	err := <-c.errorChannel
	panic(err)
}

func NewClient(hostname string, port *int) *Client {
	addressString := fmt.Sprintf(hostname+":%v", *port)
	s, err := net.ResolveUDPAddr("udp", addressString)
	if err != nil {
		panic(err)
	}
	cAddr, err := net.ResolveUDPAddr("udp", ":")
	if err != nil {
		panic(err)
	}
	c, err := net.ListenUDP("udp", cAddr)
	if err != nil {
		panic(err)
	}

	client := &Client{
		make(chan string),
		make(chan InboundMessage),
		make(chan error),
		s, c,
		make(map[string][]Peer),
	}
	go client.errorHandler()
	return client

}

func (c *Client) ConnectToRoom(inputStream chan string, roomName string) {
	// Continous Read & Writes.
	roomMessage := RoomMessage{roomName}
	raw, err := roomMessage.RawMessage()
	if err != nil {
		panic(err)
	}
	message := &Message{raw, CONNECT_TO_ROOM, false, uint16(len(raw.Data))}
	data, err := message.EncodeMessage()
	if err != nil {
		panic(err)
	}
	c.conn.WriteTo(data, c.middleMan)
	log.Info("Join room")
	if err != nil {
		panic(err)
	}
	room := c.rooms[roomName]
	if room != nil {
		room = make([]Peer, 10)
	}
	c.rooms[roomName] = room
	log.Infof("Listening on...%v", c.conn.LocalAddr())

	go c.ClientContiniousWrite(inputStream, roomName)
	panic(<-c.errorChannel)
}

func (c *Client) ConnectToMiddleMan() {

}

func (c *Client) StartUp(displayChan chan string) {
	go c.ClientContiniousRead()
	go c.Display(displayChan)
}

func (c *Client) Display(displayChan chan string) {
	for {
		message := <-c.clientChannel
		if message.Type() == ROOM_MESSAGE {
			var chatMessage ChatMessage
			err := chatMessage.DecodeMessage(message.RawData())
			if err != nil {
				log.Error(err)
			}
			log.Infof("Display coroutine decoding message")
			displayChan <- fmt.Sprintf("%v says \"%v\" to room %v", message.Sender(), chatMessage.Message, chatMessage.Room)
			log.Infof("Dropped to dispaly chan %v", message)
		}
	}
}

func (c *Client) ClientContiniousRead() {
	buf := make([]byte, MAX_UDP_DATAGRAM)
	for {
		n, sender, err := c.conn.ReadFromUDP(buf)
		var message Message
		message.RawMessage.Sender = sender
		message.DecodeMessage(sender, buf[:n])
		log.Infof("Got message from %v", sender)
		if n > 0 && err == nil {
			if message.Type() == PING {
				log.Infof("Pong recieved from %v", sender)
				c.Pong()
			} else if message.Type() == ROOM_MESSAGE {
				log.Infof("Room message %v", sender)
				c.clientChannel <- &message
			} else if message.Type() == ROOM_LIST {
				log.Infof("Room list from %v", sender)
				c.UpdateRoomList(message)
			}

		} else if err != nil {
			log.Infof("Error %v", err)
			c.errorChannel <- err
		}
	}
}

func (c *Client) Pong() {
	m := &Message{RawMessage{nil, make([]byte, 0)}, PONG, false, 0}
	data, err := m.EncodeMessage()
	if err != nil {
		panic(err)
	}
	c.conn.WriteToUDP(data, c.middleMan)
}

func (c *Client) ClientContiniousWrite(messageChan chan string, roomName string) {
	for {
		text := <-messageChan

		for _, client := range c.rooms[roomName] {
			roomMes := &ChatMessage{RoomMessage{roomName}, text}
			roomData, err := roomMes.EncodeMessage()
			if err != nil {
				panic(err)
			}
			sendMe := Message{RawMessage{nil, roomData}, ROOM_MESSAGE, false, uint16(len(roomData))}
			data, err := sendMe.EncodeMessage()
			if err != nil {
				panic(err)
			}
			n, err := c.conn.WriteToUDP(data, &client.UDPAddr)
			if n > 0 && err == nil {
				log.Infof("Sent to %v", client)
			} else if err != nil {
				log.Critical(err)
				c.errorChannel <- err
			}
		}
	}
}

func (c *Client) UpdateRoomList(message Message) {
	var rm RoomListMessage
	rm.DecodeMessage(message.Data)
	c.rooms[rm.Room] = make([]Peer, rm.Length)
	log.Infof("Updating room %s", rm.Room)
	for i := 0; i < len(rm.Addresses); i++ {
		c.rooms[rm.Room][i] = Peer{rm.Addresses[i], ""}
	}
	log.Info("Room ", c.rooms[rm.Room])

}

func (c *Client) MakeRoomMessage(roomName, message string) Message {
	var sending Message
	sending.EncryptedMsg = false
	sending.MsgType = ROOM_MESSAGE
	sendMe := &ChatMessage{RoomMessage{roomName}, message}
	writeToMe := bytes.NewBuffer(make([]byte, 0))
	err := binary.Write(writeToMe, binary.LittleEndian, sendMe)
	if err != nil {
		panic(err)
	}
	return sending
}
