package esbridge

import (
	shared2 "github.com/just1689/entity-sync/entitysync/shared"
)

type client struct {
	Subscriptions map[string]shared2.EntityKey
	ToWS          shared2.ByteHandler
	RemoteDC      chan bool
}

func (c *client) Subscribe(key shared2.EntityKey) {
	c.Subscriptions[key.Hash()] = key
}

func (c *client) UnSubscribe(key shared2.EntityKey) {
	delete(c.Subscriptions, key.Hash())
}

func removeClient(b *Bridge, c *client) {
	for i, row := range clients {
		if row == c {
			clients[i] = clients[len(clients)-1]
			clients = clients[:len(clients)-1]
			break
		}
	}
}
