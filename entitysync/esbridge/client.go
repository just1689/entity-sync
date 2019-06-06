package esbridge

import "github.com/just1689/entity-sync/entitysync/shared"

type client struct {
	Subscriptions map[string]shared.EntityKey
	ToWS          shared.ByteHandler
	RemoteDC      chan bool
}

func (c *client) Subscribe(key shared.EntityKey) {
	c.Subscriptions[key.Hash()] = key
}

func (c *client) UnSubscribe(key shared.EntityKey) {
	delete(c.Subscriptions, key.Hash())
}

func removeClient(b *Bridge, c *client) {
	for i, row := range b.clients {
		if row == c {
			b.clients[i] = b.clients[len(b.clients)-1]
			b.clients = b.clients[:len(b.clients)-1]
			break
		}
	}
}
