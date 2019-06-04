package esweb

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/just1689/entity-sync/shared"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"time"
)

func HandleEntity(mux *http.ServeMux, bridgeClientBuilder shared.ByteHandlingRemoteProxy) {
	itemHub := newHub(bridgeClientBuilder)
	go itemHub.run()
	mux.HandleFunc("/ws/entity-sync/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(itemHub, w, r)
	})

}

type hub struct {
	clients             map[*client]bool
	register            chan *client
	unregister          chan *client
	bridgeClientBuilder shared.ByteHandlingRemoteProxy
}

func newHub(bridgeClientBuilder shared.ByteHandlingRemoteProxy) *hub {
	return &hub{
		register:            make(chan *client),
		unregister:          make(chan *client),
		clients:             make(map[*client]bool),
		bridgeClientBuilder: bridgeClientBuilder,
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
				c.bridgeProxy.queueDCNotify <- true
				close(c.bridgeProxy.queueDCNotify)
			}
		}
	}
}

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

// client is a middleman between the websocket connection and the hub.
type client struct {
	hub         *hub
	conn        *websocket.Conn
	send        chan []byte
	bridgeProxy bridgeProxy
}

type bridgeProxy struct {
	entityKeyHandlers map[shared.Action]shared.EntityKeyHandler
	queueDCNotify     chan bool
}

func (c *client) handleReadMsg(message []byte) {
	m := shared.MessageAction{}
	err := json.Unmarshal(message, &m)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	if f, found := c.bridgeProxy.entityKeyHandlers[m.Action]; found {
		f(m.EntityKey)
	} else {
		logrus.Errorln("Unknown action", m.Action)
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	readPumpToClient(c)
}

func readPumpToClient(c *client) {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.handleReadMsg(message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	writePumpToClient(c, ticker)
}

func writePumpToClient(c *client, ticker *time.Ticker) {
	for {
		var err error
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err = c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logrus.Errorln(err)
				continue
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
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
			entityKeyHandlers: make(map[shared.Action]shared.EntityKeyHandler),
		},
	}
	c.hub.register <- c

	c.bridgeProxy.entityKeyHandlers[shared.ActionSubscribe],
		c.bridgeProxy.entityKeyHandlers[shared.ActionUnSubscribe],
		c.bridgeProxy.queueDCNotify = hub.bridgeClientBuilder(func(barr []byte) {
		c.send <- barr
	})

	go c.writePump()
	go c.readPump()
}
