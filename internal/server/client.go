package server

import (
	"github.com/gorilla/websocket"
	"github.com/leogsouza/collabdraw/internal/util"
	uuid "github.com/satori/go.uuid"
)

type Client struct {
	ID       string
	Hub      *Hub
	Color    string
	Socket   *websocket.Conn
	Outbound chan []byte
}

func NewClient(hub *Hub, socket *websocket.Conn) *Client {
	return &Client{
		ID:       uuid.NewV4().String(),
		Color:    util.GenerateColor(),
		Hub:      hub,
		Socket:   socket,
		Outbound: make(chan []byte),
	}
}

func (client *Client) Read() {
	defer func() {
		client.Hub.Unregister <- client
	}()

	for {
		_, data, err := client.Socket.ReadMessage()
		if err != nil {
			break
		}
		client.Hub.OnMessage(data, client)
	}
}

func (client *Client) Write() {
	for {
		select {
		case data, ok := <-client.Outbound:
			if !ok {
				client.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			client.Socket.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func (client Client) Run() {
	go client.Read()
	go client.Write()
}

func (client Client) Close() {
	client.Socket.Close()
	close(client.Outbound)
}
