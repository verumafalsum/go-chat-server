package main

import (
	"fmt"
	"log"
	"net/http"
)

var addr = ":8080"

func main() {
	fmt.Println("Init main.go")
	clientHub := createClientHub()
	go clientHub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(clientHub, w, r)
	})
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
