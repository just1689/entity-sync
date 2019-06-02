package dq

import (
	"github.com/bitly/go-nsq"
	"github.com/just1689/entity-sync/shared"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"log"
	"sync"
)

func GetNSQProducer(nsqAddr string, topic string) shared.ByteHandler {
	// addr = "127.0.0.1:4150"
	config := nsq.NewConfig()
	w, _ := nsq.NewProducer(nsqAddr, config)
	producer := func(msg []byte) {
		err := w.Publish(topic, msg)
		if err != nil {
			logrus.Errorln(err)
		}

	}
	return producer

}

func SubscribeNSQ(nsqAddr string, topic string, f shared.ByteHandler) {

	wg := &sync.WaitGroup{}
	wg.Add(1)

	config := nsq.NewConfig()
	q, _ := nsq.NewConsumer(topic, uuid.NewV4().String(), config)
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		f(message.Body)
		//log.Printf("Got a message: %v", message)
		wg.Done()
		return nil
	}))
	err := q.ConnectToNSQD(nsqAddr)
	if err != nil {
		log.Panic("Could not connect")
	}
	wg.Wait()

}
