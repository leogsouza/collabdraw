package server

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
