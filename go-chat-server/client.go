package main

import (
	"bytes"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	id     uuid.UUID
	socket *websocket.Conn
	send   chan []byte
	hub    *ClientHub
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (client *Client) readPump() {
	defer func() {
		client.hub.unregister <- client
		client.socket.Close()
	}()

	for {
		_, message, err := client.socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		client.hub.broadcast <- message
	}
}

func serveWs(clientHub *ClientHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	clientID := uuid.New()
	client := &Client{id: clientID, socket: conn, hub: clientHub, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.readPump()
}
