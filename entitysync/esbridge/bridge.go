package esbridge

import (
	"encoding/json"
	"github.com/just1689/entity-sync/entitysync/shared"
	"github.com/sirupsen/logrus"
	"sync"
)

func BuildBridge(queuePublisherBuilder shared.EntityHandler, queueSubscriberBuilder shared.EntityByteHandler, dbPullDataAndPush shared.EntityKeySecretByteHandler) *Bridge {
	return &Bridge{
		queueFunctions: queueFunctions{
			queuePublisherBuilder:  queuePublisherBuilder,
			queueSubscriberBuilder: queueSubscriberBuilder,
			queuePublishers:        make(map[shared.EntityType]shared.ByteHandler),
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
	dbPullDataAndPush shared.EntityKeySecretByteHandler
}

type queueFunctions struct {
	queuePublishers        map[shared.EntityType]shared.ByteHandler
	queuePublisherBuilder  shared.EntityHandler
	queueSubscriberBuilder shared.EntityByteHandler
}

//RegisterEntityForSync informs the framework that it needs to publish changes for this entity type and receive them
func (b *Bridge) RegisterEntityForSync(entityType shared.EntityType) {
	b.createQueuePublishers(entityType)
	b.subscribe(entityType)
}

func (b *Bridge) ClientBuilder(ToWS shared.ByteHandler) (sub shared.EntityKeyHandler, unSub shared.EntityKeyHandler, dc chan bool) {
	c := client{
		ToWS:          ToWS,
		RemoteDC:      make(chan bool),
		Subscriptions: make(map[string]shared.EntityKey),
	}
	b.clients = append(b.clients, &c)
	b.blockOnDisconnect(&c)
	return c.Subscribe, c.UnSubscribe, c.RemoteDC
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

func (b *Bridge) createQueuePublishers(entity shared.EntityType) {
	b.queueFunctions.queuePublishers[entity] = b.queueFunctions.queuePublisherBuilder(entity)
}
func (b *Bridge) subscribe(entityType shared.EntityType) {
	b.queueFunctions.queueSubscriberBuilder(entityType, func(barr []byte) {
		key := shared.EntityKey{}
		if err := json.Unmarshal(barr, &key); err != nil {
			logrus.Errorln(err)
			return
		}
		b.onQueueIncoming(key)
	})
}

func (b *Bridge) onQueueIncoming(key shared.EntityKey) {
	b.m.Lock()
	defer b.m.Unlock()
	for _, c := range b.clients {
		if _, found := c.Subscriptions[key.Hash()]; found == false {
			continue
		}
		// TODO: Get the secret into the BRIDGE
		b.dbPullDataAndPush(key, "SECRET", c.ToWS)
	}
}

func (b *Bridge) blockOnDisconnect(c *client) {
	go func() {
		<-c.RemoteDC
		b.m.Lock()
		removeClient(b, c)
		b.m.Unlock()
	}()
}
