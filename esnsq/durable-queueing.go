package esnsq

import (
	"fmt"
	"github.com/bitly/go-nsq"
	"github.com/just1689/entity-sync/shared"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

//BuildPublisher can be called to give a method that in turn can be called to create publishers
func BuildPublisher(nsqAddr string) shared.EntityHandler {
	return func(entityType shared.EntityType) shared.ByteHandler {
		return GetNSQProducer(nsqAddr, entityType)
	}
}

func GetNSQProducer(nsqAddr string, entityType shared.EntityType) shared.ByteHandler {
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

func BuildSubscriber(nsqAddr string) shared.EntityByteHandler {
	return func(entityType shared.EntityType, callback shared.ByteHandler) {
		SubscribeNSQ(nsqAddr, entityType, callback)
	}
}

func SubscribeNSQ(nsqAddr string, entityType shared.EntityType, f shared.ByteHandler) {
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
