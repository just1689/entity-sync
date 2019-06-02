package main

import (
	"github.com/just1689/entity-sync/bridge"
	"github.com/just1689/entity-sync/dq"
	"github.com/just1689/entity-sync/x/io"
	"net"
	"net/http"
	"time"
)

var nsqAddr = "127.0.0.1:4150"

func main() {

	mux := http.NewServeMux()
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	topic := "items"
	itemProducer := dq.GetNSQProducer(nsqAddr, topic)
	bridge.GlobalBridge.AddQueuePublisher(topic, itemProducer)

	bridge.GlobalBridge.AddQueueSubscriber(topic)

	io.HandleEntity(mux, "items", "id", func(val string) (item interface{}, err error) {
		// Pretend to get the entity
		item = ItemV1{
			ID:         val,
			Closed:     false,
			ClosedDate: time.Now(),
		}
		return
	})

	err = http.Serve(l, mux)
	if err != nil {
		panic(err)
	}

}
