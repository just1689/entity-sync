package esweb

import (
	"github.com/gorilla/websocket"
	shared2 "github.com/just1689/entity-sync/entitysync/shared"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func SetupMuxBridge(mux *http.ServeMux, bridgeClientBuilder shared2.ByteHandlingRemoteProxy) {
	itemHub := newHub(bridgeClientBuilder)
	go run()
	mux.HandleFunc("/ws/entity-sync/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(itemHub, w, r)
	})

}

// serveWs handles websocket requests from the peer.
func serveWs(hub *hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := &client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		bridgeProxy: bridgeProxy{
			entityKeyHandlers: make(map[shared2.Action]shared2.EntityKeyHandler),
		},
	}
	register <- c

	entityKeyHandlers[shared2.ActionSubscribe],
		entityKeyHandlers[shared2.ActionUnSubscribe],
		queueDCNotify = bridgeClientBuilder(
		func(barr []byte) {
			send <- barr
		})

	go writePump()
	go readPump()
}
