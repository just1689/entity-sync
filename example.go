package main

import (
	"fmt"
	"github.com/just1689/entity-sync/bridge"
	"github.com/just1689/entity-sync/dq"
	"github.com/just1689/entity-sync/web"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
)

const nsqAddr = "127.0.0.1:4150"
const topic = "items"

func main() {

	mux := http.NewServeMux()
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		logrus.Fatalln(err)
	}

	itemNSQProducer := dq.GetNSQProducer(nsqAddr, topic)
	bridge.GlobalBridge.AddQueuePublisher(topic, itemNSQProducer)

	bridge.GlobalBridge.AddQueueSubscriber(topic)

	HandleEntity(mux, topic)

	//io.HandleEntity(mux, "items", "id", func(val string) (item interface{}, err error) {
	//	// Pretend to get the entity
	//	item = ItemV1{
	//		ID:         val,
	//		Closed:     false,
	//		ClosedDate: time.Now(),
	//	}
	//	return
	//})

	err = http.Serve(l, mux)
	if err != nil {
		panic(err)
	}

}

func HandleEntity(mux *http.ServeMux, topic string) {
	itemHub := web.NewHub(topic)
	go itemHub.Run()

	mux.HandleFunc(fmt.Sprint("/ws/entity-sync/", topic), func(w http.ResponseWriter, r *http.Request) {
		ServeWs(itemHub, w, r)
	})

}
