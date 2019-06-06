package esweb

import (
	"github.com/just1689/entity-sync/entitysync/shared"
)

func newHub(bridgeClientBuilder shared.ByteHandlingRemoteProxy) *hub {
	return &hub{
		register:            make(chan *client),
		unregister:          make(chan *client),
		clients:             make(map[*client]bool),
		bridgeClientBuilder: bridgeClientBuilder,
	}
}

type hub struct {
	clients             map[*client]bool
	register            chan *client
	unregister          chan *client
	bridgeClientBuilder shared.ByteHandlingRemoteProxy
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
