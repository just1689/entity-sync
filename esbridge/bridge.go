package esbridge

import (
	"encoding/json"
	"github.com/just1689/entity-sync/shared"
	"github.com/sirupsen/logrus"
	"sync"
)

func BuildBridge(queuePublisherBuilder shared.EntityHandler, queueSubscriberBuilder shared.EntityByteHandler, dbPullDataAndPush shared.EntityKeyByteHandler) *Bridge {
	return &Bridge{
		queueFunctions: QueueFunctions{
			queuePublisherBuilder:  queuePublisherBuilder,
			queueSubscriberBuilder: queueSubscriberBuilder,
			queuePublishers:        make(map[shared.EntityType]shared.ByteHandler),
		},
		clients:           make([]*Client, 0),
		dbPullDataAndPush: dbPullDataAndPush,
	}
}

type Bridge struct {
	m sync.Mutex

	queueFunctions QueueFunctions

	//Websocket clients
	clients []*Client

	//Database function
	dbPullDataAndPush shared.EntityKeyByteHandler
}

type QueueFunctions struct {
	queuePublishers        map[shared.EntityType]shared.ByteHandler
	queuePublisherBuilder  shared.EntityHandler
	queueSubscriberBuilder shared.EntityByteHandler
}

func (b *Bridge) removeClient(c *Client) {
	for i, client := range b.clients {
		if client == c {
			b.clients[i] = b.clients[len(b.clients)-1]
			b.clients = b.clients[:len(b.clients)-1]
			break
		}
	}
}

func (b *Bridge) CreateQueuePublishers(entity shared.EntityType) {
	b.queueFunctions.queuePublishers[entity] = b.queueFunctions.queuePublisherBuilder(entity)
}

//NotifyAll can be called to publish to all nodes (via NSQ) that a row of EntityType has changed
func (b *Bridge) NotifyAllOfChange(key shared.EntityKey) {
	pub, found := b.queueFunctions.queuePublishers[key.Entity]
	if !found {
		logrus.Fatalln("Could not notify all for entity", key.Entity)
	}
	barr, err := json.Marshal(key)
	if err != nil {
		logrus.Fatalln("Could not json EntityKey", err)
	}
	pub(barr)

}

func (b *Bridge) Subscribe(entityType shared.EntityType) {
	b.queueFunctions.queueSubscriberBuilder(entityType, func(barr []byte) {
		key := shared.EntityKey{}
		if err := json.Unmarshal(barr, &key); err != nil {
			logrus.Errorln(err)
			return
		}
		b.onNotify(key)
	})
}

func (b *Bridge) onNotify(key shared.EntityKey) {
	b.m.Lock()
	defer b.m.Unlock()
	logrus.Println("onNotify")
	for _, client := range b.clients {
		if _, found := client.Subscriptions[key.Hash()]; found == false {
			continue
		}
		b.dbPullDataAndPush(key, client.ToWS)
	}
}

func (b *Bridge) blockOnDisconnect(c *Client) {
	go func() {
		<-c.RemoteDC
		b.m.Lock()
		b.removeClient(c)
		b.m.Unlock()
	}()
}

func (b *Bridge) ClientBuilder(ToWS shared.ByteHandler) (sub shared.EntityKeyHandler, unSub shared.EntityKeyHandler, dc chan bool) {
	client := Client{
		ToWS:          ToWS,
		RemoteDC:      make(chan bool),
		Subscriptions: make(map[string]shared.EntityKey),
	}
	b.clients = append(b.clients, &client)
	b.blockOnDisconnect(&client)
	return client.Subscribe, client.UnSubscribe, client.RemoteDC
}
