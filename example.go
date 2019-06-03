package main

import (
	"github.com/just1689/entity-sync/bridge"
	"github.com/just1689/entity-sync/dq"
	"github.com/just1689/entity-sync/shared"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
)

const nsqAddr = "127.0.0.1:4150"
const entityType shared.EntityType = "items"

var GlobalBridge *bridge.Bridge

func main() {

	mux := http.NewServeMux()
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		logrus.Fatalln(err)
	}

	//Build the bridge
	GlobalBridge = bridge.BuildBridge(dq.BuildPublisher(nsqAddr))

	//Create publisher
	GlobalBridge.CreateQueuePublishers(entityType)

	//Ensure the bridge will send NSQ messages for entityType to onNotify
	GlobalBridge.Subscribe(entityType)

	HandleEntity(mux, topic)

	err = http.Serve(l, mux)
	if err != nil {
		panic(err)
	}

}
