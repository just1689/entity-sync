package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/just1689/entity-sync/shared"
	"github.com/sirupsen/logrus"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {

	flag.Parse()
	log.SetFlags(0)
	var err error

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws/entity-sync/"}
	log.Printf("connecting to %s", u.String())

	var c *websocket.Conn
	if c, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		log.Fatal("Could not dial ws:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	go readMessages(done, c)

	//Subscribe to entity for a particular ID
	var b []byte
	if b, err = json.Marshal(shared.MessageAction{
		Action:    shared.ActionSubscribe,
		EntityKey: shared.EntityKey{Entity: "items", ID: "100"},
	}); err != nil {
		logrus.Fatal(err)
	}
	logrus.Println("Sending", string(b))
	if err = c.WriteMessage(websocket.TextMessage, b); err != nil {
		logrus.Fatal(err)
	}

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func readMessages(done chan struct{}, c *websocket.Conn) {
	defer close(done)
	logrus.Println("Starting reader (blocking)")
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)
	}
}
