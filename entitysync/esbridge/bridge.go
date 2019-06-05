package esbridge

import (
	"encoding/json"
	shared2 "github.com/just1689/entity-sync/entitysync/shared"
	"github.com/sirupsen/logrus"
	"sync"
)

func BuildBridge(queuePublisherBuilder shared2.EntityHandler, queueSubscriberBuilder shared2.EntityByteHandler, dbPullDataAndPush shared2.EntityKeyByteHandler) *Bridge {
	return &Bridge{
		queueFunctions: queueFunctions{
			queuePublisherBuilder:  queuePublisherBuilder,
			queueSubscriberBuilder: queueSubscriberBuilder,
			queuePublishers:        make(map[shared2.EntityType]shared2.ByteHandler),
		},
		clients:           make([]*client, 0),
		dbPullDataAndPush: dbPullDataAndPush,
	}
}

type Bridge struct {
	m sync.Mutex

	//Allows subscribing and pushing to remote queues
	queueFunctions queueFunctions

	//Websocket clients
	clients []*client

	//Database function
	dbPullDataAndPush shared2.EntityKeyByteHandler
}

type queueFunctions struct {
	queuePublishers        map[shared2.EntityType]shared2.ByteHandler
	queuePublisherBuilder  shared2.EntityHandler
	queueSubscriberBuilder shared2.EntityByteHandler
}

//SyncEntityType informs the framework that it needs to publish changes for this entity type and receive them
func (b *Bridge) SyncEntityType(entityType shared2.EntityType) {
	b.createQueuePublishers(entityType)
	b.subscribe(entityType)
}

func (b *Bridge) ClientBuilder(ToWS shared2.ByteHandler) (sub shared2.EntityKeyHandler, unSub shared2.EntityKeyHandler, dc chan bool) {
	c := client{
		ToWS:          ToWS,
		RemoteDC:      make(chan bool),
		Subscriptions: make(map[string]shared2.EntityKey),
	}
	b.clients = append(b.clients, &c)
	b.blockOnDisconnect(&c)
	return Subscribe, UnSubscribe, RemoteDC
}

//NotifyAll can be called to publish to all nodes (via NSQ) that a row of EntityType has changed
func (b *Bridge) NotifyAllOfChange(key shared2.EntityKey) {
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

func (b *Bridge) createQueuePublishers(entity shared2.EntityType) {
	b.queueFunctions.queuePublishers[entity] = b.queueFunctions.queuePublisherBuilder(entity)
}
func (b *Bridge) subscribe(entityType shared2.EntityType) {
	b.queueFunctions.queueSubscriberBuilder(entityType, func(barr []byte) {
		key := shared2.EntityKey{}
		if err := json.Unmarshal(barr, &key); err != nil {
			logrus.Errorln(err)
			return
		}
		b.onQueueIncoming(key)
	})
}

func (b *Bridge) onQueueIncoming(key shared2.EntityKey) {
	b.m.Lock()
	defer b.m.Unlock()
	for _, c := range b.clients {
		if _, found := Subscriptions[key.Hash()]; found == false {
			continue
		}
		b.dbPullDataAndPush(key, ToWS)
	}
}

func (b *Bridge) blockOnDisconnect(c *client) {
	go func() {
		<-RemoteDC
		b.m.Lock()
		removeClient(b, c)
		b.m.Unlock()
	}()
}
