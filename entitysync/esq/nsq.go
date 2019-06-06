package esq

import (
	"fmt"
	"github.com/just1689/entity-sync/entitysync/shared"
	"github.com/nsqio/go-nsq"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

func getNSQProducer(nsqAddr string, entityType shared.EntityType) shared.ByteHandler {
	config := nsq.NewConfig()
	w, _ := nsq.NewProducer(nsqAddr, config)
	producer := func(msg []byte) {
		err := w.Publish(entityType.GetQueueName(), msg)
		if err != nil {
			logrus.Panic("Could not publish to NSQ ", nsqAddr, "on topic", entityType)
		}
	}
	return producer
}

func subscribeNSQ(nsqAddr string, entityType shared.EntityType, f shared.ByteHandler) {
	config := nsq.NewConfig()
	q, _ := nsq.NewConsumer(entityType.GetQueueName(), fmt.Sprint(uuid.NewV4().String(), "#ephemeral"), config)
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		f(message.Body)
		return nil
	}))
	err := q.ConnectToNSQD(nsqAddr)
	if err != nil {
		logrus.Panic("Could not connect to NSQ for subscribe", nsqAddr)
	}

}
