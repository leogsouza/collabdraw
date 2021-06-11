package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/leogsouza/collabdraw/internal/message"
	"github.com/tidwall/gjson"
)

type Hub struct {
	Clients    []*Client
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make([]*Client, 0),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.Register:
			hub.OnConnect(client)
		case client := <-hub.Unregister:
			hub.OnDisconnect(client)
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(req *http.Request) bool { return true },
}

func (hub *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, "could not upgrade", http.StatusInternalServerError)
		return
	}
	client := NewClient(hub, socket)
	hub.Clients = append(hub.Clients, client)
	hub.Register <- client
	client.Run()
}

func (hub *Hub) Send(message interface{}, client *Client) {
	data, _ := json.Marshal(message)
	client.Outbound <- data
}

func (hub *Hub) Broadcast(message interface{}, ignore *Client) {
	data, _ := json.Marshal(message)
	for _, c := range hub.Clients {
		if c != ignore {
			c.Outbound <- data
		}
	}
}

func (hub *Hub) OnConnect(client *Client) {
	log.Println("client connected:", client.Socket.RemoteAddr())

	// Make list of all users
	users := []message.User{}

	for _, c := range hub.Clients {
		users = append(users, message.User{ID: c.ID, Color: c.Color})
	}

	// Notify user joined
	hub.Send(message.NewConnected(client.Color, users), client)
	hub.Broadcast(message.NewUserJoined(client.ID, client.Color), client)
}

func (hub *Hub) OnDisconnect(client *Client) {
	log.Println("client disconnected:", client.Socket.RemoteAddr())
	client.Close()

	i := -1
	for j, c := range hub.Clients {
		if c.ID == client.ID {
			i = j
			break
		}
	}

	// Remove client from list
	copy(hub.Clients[i:], hub.Clients[i+1:])
	hub.Clients[len(hub.Clients)-1] = nil
	hub.Clients = hub.Clients[:len(hub.Clients)-1]

	hub.Broadcast(message.NewUserLeft(client.ID), nil)
}

func (hub *Hub) OnMessage(data []byte, client *Client) {
	kind := gjson.GetBytes(data, "kind").Int()

	if kind == message.KindStroke {
		var msg message.Stroke
		if json.Unmarshal(data, &msg) != nil {
			return
		}

		msg.UserID = client.ID
		hub.Broadcast(msg, client)
	} else if kind == message.KindClear {
		var msg message.Clear
		if json.Unmarshal(data, &msg) != nil {
			return
		}

		msg.UserID = client.ID
		hub.Broadcast(msg, client)
	}
}
