package main

import "github.com/google/uuid"

type Message struct {
	senderID uuid.UUID
	content  []byte
}
