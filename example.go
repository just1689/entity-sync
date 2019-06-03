package main

import (
	"flag"
	"fmt"
	"github.com/just1689/entity-sync/bridge"
	"github.com/just1689/entity-sync/dq"
	"github.com/just1689/entity-sync/shared"
	"github.com/just1689/entity-sync/web"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"net/http"
)

const nsqAddr = "127.0.0.1:4150"
const entityType shared.EntityType = "items"

var GlobalBridge *bridge.Bridge

var Name string
var role = flag.Int("role", 1, "1 for reader, 2 for changer")
var listenLocal = flag.String("listen", ":8080", "listen addr: :8080")

func main() {
	flag.Parse()

	resolveName()

	mux := http.NewServeMux()
	l, err := net.Listen("tcp", *listenLocal)
	if err != nil {
		logrus.Fatalln(err)
	}

	//Build the bridge
	GlobalBridge = bridge.BuildBridge(dq.BuildPublisher(nsqAddr))

	//Create publisher
	GlobalBridge.CreateQueuePublishers(entityType)

	//Ensure the bridge will send NSQ messages for entityType to onNotify
	GlobalBridge.Subscribe(entityType)

	web.HandleEntity(mux)

	err = http.Serve(l, mux)
	if err != nil {
		panic(err)
	}

}

func resolveName() {
	if *role == 1 {
		Name = fmt.Sprint("Reader", rand.Intn(100))
	} else if *role == 2 {
		Name = fmt.Sprint("Mutator", rand.Intn(100))
	}
}
