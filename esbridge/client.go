package esbridge

import "github.com/just1689/entity-sync/shared"

type Client struct {
	Subscriptions map[string]shared.EntityKey
	ToWS          shared.ByteHandler
	RemoteDC      chan bool
}

func (c *Client) Subscribe(key shared.EntityKey) {
	c.Subscriptions[key.Hash()] = key
}

func (c *Client) UnSubscribe(key shared.EntityKey) {
	delete(c.Subscriptions, key.Hash())
}

func removeClient(b *Bridge, c *Client) {
	for i, client := range b.clients {
		if client == c {
			b.clients[i] = b.clients[len(b.clients)-1]
			b.clients = b.clients[:len(b.clients)-1]
			break
		}
	}
}
