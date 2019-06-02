package dq

import (
	"github.com/bitly/go-nsq"
	"github.com/just1689/entity-sync/shared"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

func GetNSQProducer(nsqAddr string, topic string) shared.ByteHandler {
	config := nsq.NewConfig()
	w, _ := nsq.NewProducer(nsqAddr, config)
	producer := func(msg []byte) {
		err := w.Publish(topic, msg)
		if err != nil {
			logrus.Panic("Could not publish to NSQ ", nsqAddr, "on topic", topic)
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
