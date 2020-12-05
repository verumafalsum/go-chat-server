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
	send   chan *Message
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
		message := &Message{senderID: client.id, content: make([]byte, 256)}
		_, content, err := client.socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		content = bytes.TrimSpace(bytes.Replace(content, newline, space, -1))
		message.content = content
		client.hub.broadcast <- message
	}
}

func (client *Client) writePump() {
	defer func() {
		client.socket.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				// The hub closed the channel.
				client.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.socket.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message.content)

			// Add queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				message, _ := <-client.send
				w.Write(message.content)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func serveWs(clientHub *ClientHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	clientID := uuid.New()
	client := &Client{id: clientID, socket: conn, hub: clientHub, send: make(chan *Message)}
	client.hub.register <- client

	go client.readPump()
	go client.writePump()
}
