package main

import "log"

// ClientHub maintains the set of active clients and broadcasts messages to the
// clients.
type ClientHub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages
	broadcast chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from the clients.
	unregister chan *Client
}

func createClientHub() *ClientHub {
	return &ClientHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *ClientHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			log.Println("Message received from: ", message.senderID)
			log.Println("Message: ", message.content)
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
