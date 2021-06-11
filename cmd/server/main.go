package main

import (
	"log"
	"net/http"

	"github.com/leogsouza/collabdraw/internal/server"
)

func main() {
	hub := server.NewHub()
	go hub.Run()
	http.HandleFunc("/ws", hub.HandleWebSocket)

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
