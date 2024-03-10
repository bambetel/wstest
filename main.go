package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	var upgrader = websocket.Upgrader{}
	tmpl := template.Must(template.ParseFiles("client.html"))
	// upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade failed: %+v\n", err)
			return
		}
		log.Println("Websocket Connected!")
		for {
			messageType, messageContent, err := conn.ReadMessage()
			timeReceived := time.Now()
			log.Printf("NEW: %v, %v, %v\n", messageType, messageContent, err)

			if c, ok := err.(*websocket.CloseError); ok {
				log.Printf("Close websocket: %v (%+v/%[2]T)\n", err, c.Code)
				return
			} else if err != nil {
				log.Println("Other error:", err)
				return
			}

			messageResponse := fmt.Sprintf("Message (%v): %s %s\n", messageType, messageContent, timeReceived)
			log.Println("RESP:", messageResponse)
			if err := conn.WriteMessage(messageType, []byte(messageResponse)); err != nil {
				log.Println("Error writing message:", err)
				return
			}
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("get home")
		tmpl.Execute(w, "ws://"+r.Host+"/ws")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
