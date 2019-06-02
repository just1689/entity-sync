package bridge

import (
	"encoding/json"
	"github.com/just1689/entity-sync/db"
	"github.com/just1689/entity-sync/shared"
	"github.com/sirupsen/logrus"
	"sync"
)

func BuildBridge(queuePublisherBuilder shared.EntityHandler) *Bridge {
	return &Bridge{
		queuePublisherBuilder: queuePublisherBuilder,
		queuePublishers:       make(map[shared.Entity]shared.ByteHandler),
		clients:               make([]*Client, 0),
	}
}

type Bridge struct {
	m sync.Mutex

	//queuePublishers write to on a entity
	queuePublishers       map[shared.Entity]shared.ByteHandler
	queuePublisherBuilder shared.EntityHandler

	clients []*Client
}

func (b *Bridge) CreateQueuePublishers(entity shared.Entity) {
	b.queuePublishers[entity] = b.queuePublisherBuilder(entity)

}

//NotifyAll can be called to publish to all nodes that a row of type Entity has changed
func (b *Bridge) NotifyAllOfChange(key shared.RowKey) {
	pub, found := b.queuePublishers[key.Entity]
	if !found {
		logrus.Fatalln("Could not notify all for entity", key.Entity)
	}
	barr, err := json.Marshal(key)
	if err != nil {
		logrus.Fatalln("Could not json RowKey", err)
	}
	pub(barr)

}

func (b *Bridge) onNotify(key shared.RowKey) {
	b.m.Lock()
	defer b.m.Unlock()

	for _, client := range b.clients {
		rowKey, found := client.Subscriptions[key.Hash()]
		if !found {
			logrus.Infoln("Client did not subscribe to", rowKey.Hash())
			continue
		}
		logrus.Println("Found client to notify about", rowKey.Hash())
		db.GlobalDatabaseHub.ProcessUpdateHandler(key, client.SendToWS)
	}
}

func (b *Bridge) BlockOnDisconnect(entity shared.Entity, c *Client, index int) {
	go func() {
		<-c.RemoteDC
		b.m.Lock()
		delete(b.queuePublishers, entity)
		b.m.Unlock()
	}()
}

type Client struct {
	Subscriptions map[string]shared.RowKey
	ToWS          chan []byte
	RemoteDC      chan bool
}

func (c *Client) Subscribe(key shared.RowKey) {
	c.Subscriptions[key.Hash()] = key
}

func (c *Client) SendToWS(msg []byte) {
	c.ToWS <- msg
}
