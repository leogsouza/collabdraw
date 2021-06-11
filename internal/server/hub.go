package server

import (
	"log"	"net/http"

	"github.com/gorilla/websocket"
)


type Hub struct {
	Clients    []Client
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
	client  := NewClient(hub, socket)
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