package esweb

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/just1689/entity-sync/es/shared"
	"github.com/sirupsen/logrus"
	"log"
	"time"
)

// client is a middleman between the websocket connection and the hub.
type client struct {
	hub         *hub
	conn        *websocket.Conn
	send        chan []byte
	bridgeProxy bridgeProxy
	secret      string
}

type bridgeProxy struct {
	entityKeyHandlers map[shared.Action]shared.EntityKeyHandler
	queueDCNotify     chan bool
}

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

func (c *client) handleReadMsg(message []byte) {
	m := shared.Message{}
	if err := json.Unmarshal(message, &m); err != nil {
		logrus.Errorln(err)
		return
	}
	if m.Action == shared.ActionSecret {
		secret := ""
		err := json.Unmarshal(m.Body, &secret)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		c.secret = secret
		return
	}

	if f, found := c.bridgeProxy.entityKeyHandlers[m.Action]; found {
		entityKey := shared.EntityKey{}
		err := json.Unmarshal(m.Body, &entityKey)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		f(entityKey)
	} else {
		logrus.Errorln("Unknown action", m.Action)
	}
}

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
