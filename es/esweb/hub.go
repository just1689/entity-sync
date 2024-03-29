package esweb

import (
	"github.com/just1689/entity-sync/es/shared"
)

func newHub(bridgeClientBuilder shared.ByteSecretHandlingRemoteProxy, msgPassThrough shared.SecretByteHandler) *hub {
	return &hub{
		register:            make(chan *client),
		unregister:          make(chan *client),
		clients:             make(map[*client]bool),
		bridgeClientBuilder: bridgeClientBuilder,
		MsgPassThrough:      msgPassThrough,
	}
}

type hub struct {
	clients             map[*client]bool
	register            chan *client
	unregister          chan *client
	bridgeClientBuilder shared.ByteSecretHandlingRemoteProxy
	MsgPassThrough      shared.SecretByteHandler
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
