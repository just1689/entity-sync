package bridge

import (
	"github.com/just1689/entity-sync/db"
	"github.com/just1689/entity-sync/shared"
	"github.com/sirupsen/logrus"
	"sync"
)

var GlobalBridge = Bridge{
	Publishers:  make(map[string]shared.ByteHandler),
	Subscribers: make(map[string][]*Client),
}

type Bridge struct {
	m           sync.Mutex
	Publishers  map[string]shared.ByteHandler
	Subscribers map[string][]*Client
}

func (b *Bridge) Notify(key shared.RowKey) {
	b.m.Lock()
	defer b.m.Unlock()
	clients, found := b.Subscribers[key.Entity]
	if !found {
		logrus.Println("No one to notify on ", key.Entity)
		return
	}

	for _, client := range clients {
		subscribedID, f := client.Subscriptions[key.Entity]
		if !f {
			logrus.Errorln("Strange, found client in list of topics for topic", key.Entity, "but not in actual client")
			continue
		}

		if subscribedID == key.ID {
			db.GlobalDatabaseHub.ProcessUpdateHandler(key, client.SendToWS)
		}

	}

}

func (b *Bridge) AddQueuePublisher(topic string, f shared.ByteHandler) {
	b.m.Lock()
	defer b.m.Unlock()
	b.Publishers[topic] = f

}

func (b *Bridge) AddQueueSubscriber(topic string) shared.ByteHandler {
	return func(msg []byte) {
		b.m.Lock()
		defer b.m.Unlock()
		clients, ok := b.Subscribers[topic]
		if !ok {
			logrus.Errorln("No clients subscribed to topic", topic)
			return
		}
		for _, client := range clients {
			client.ToWS <- msg
		}
	}

}

func (b *Bridge) SubscribeClient(topic string, c *Client) {
	b.m.Lock()
	defer b.m.Unlock()
	_, found := b.Subscribers[topic]
	if !found {
		b.Subscribers[topic] = make([]*Client, 1)
		b.Subscribers[topic][0] = c
	} else {
		b.Subscribers[topic] = append(b.Subscribers[topic], c)
	}
	b.BlockOnDisconnect(topic, c, len(b.Subscribers[topic])-1)
}

func (b *Bridge) BlockOnDisconnect(topic string, c *Client, index int) {
	go func() {
		<-c.RemoteDC
		b.m.Lock()
		b.Subscribers[topic] = append(b.Subscribers[topic][:index], b.Subscribers[topic][index+1:]...)
		b.m.Unlock()
	}()
}

type Client struct {
	Subscriptions map[string]string
	ToWS          chan []byte
	RemoteDC      chan bool
}

func (c *Client) SendToWS(msg []byte) {
	c.ToWS <- msg
}
