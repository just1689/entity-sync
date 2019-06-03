package dq

import (
	"github.com/bitly/go-nsq"
	"github.com/just1689/entity-sync/shared"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

func BuildPublisher(nsqAddr string) shared.EntityHandler {
	return func(entityType shared.EntityType) shared.ByteHandler {
		return GetNSQProducer(nsqAddr, entityType)
	}
}

func GetNSQProducer(nsqAddr string, entity shared.EntityType) shared.ByteHandler {
	config := nsq.NewConfig()
	w, _ := nsq.NewProducer(nsqAddr, config)
	producer := func(msg []byte) {
		err := w.Publish(string(entity), msg)
		if err != nil {
			logrus.Panic("Could not publish to NSQ ", nsqAddr, "on topic", entity)
		}
	}
	return producer
}

func SubscribeNSQ(nsqAddr string, topic string, f shared.ByteHandler) {
	config := nsq.NewConfig()
	q, _ := nsq.NewConsumer(topic, uuid.NewV4().String(), config)
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		f(message.Body)
		return nil
	}))
	err := q.ConnectToNSQD(nsqAddr)
	if err != nil {
		logrus.Panic("Could not connect to NSQ for subscribe", nsqAddr)
	}

}
