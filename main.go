package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func BasicRestHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Reporting!\n"))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func echo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("echo echo hit!")
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/report", BasicRestHandler)
	r.HandleFunc("/echo", echo)

	log.Println("Serving services on port 9006")
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":9006", r))
}
