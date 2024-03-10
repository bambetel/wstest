package main

import (
	"bufio"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var (
	TIMESWAIT    = 0
	TIMESWAITMAX = 5
	in           = bufio.NewReader(os.Stdin)
)

func getInput(input chan string) {
	result, err := in.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("input:", result)
	input <- result
}

func main() {
	SERVER := "localhost:8080"
	PATH := "ws"
	URL := url.URL{Scheme: "ws", Host: SERVER, Path: PATH}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	input := make(chan string, 1)
	go getInput(input)
	log.Printf("Connecting to URL %s\n", URL.String())
	conn, resp, err := websocket.DefaultDialer.Dial(URL.String(), nil)
	if err != nil {
		if err == websocket.ErrBadHandshake {
			log.Printf("handshake failed, status=%d ", resp.StatusCode)
		}
		log.Println("Error:", err)
		return
	}
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			messageType, message, err := conn.ReadMessage()
			log.Printf("READ type/err: %v, %v\n", messageType, err)

			if messageType == websocket.CloseMessage {
				log.Println("close message :))) ")
			}
			if c, k := err.(*websocket.CloseError); k {
				log.Println("close error:", err, c.Code)
				return
			} else if err != nil {
				log.Println("ReadMessage() error:", err)
				return
			}
			log.Printf("Received (%v): %s\n", messageType, message)
		}
	}()

	for {
		select {
		case <-time.After(4 * time.Second):
			log.Println("Please give me input!", TIMESWAIT)
			TIMESWAIT++
			if TIMESWAIT > TIMESWAITMAX {
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}
		case <-done:
			log.Println("<-done")
			return
		case t := <-input:
			log.Println("<-input")
			err := conn.WriteMessage(websocket.TextMessage, []byte(t))
			if err != nil {
				log.Println("Write error:", err)
				return
			}
			TIMESWAIT = 0
			go getInput(input)
		case <-interrupt:
			log.Println("Interrupt signal - quitting.")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ":)))"))
			if err != nil {
				log.Println("Write close error:", err)
				return
			}
			select {
			case <-done:
				log.Println("select done")
			case <-time.After(3 * time.Second):
				log.Println("select timeout")
			}
			return
		}
	}
}
