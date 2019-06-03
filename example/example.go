package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/just1689/entity-sync/bridge"
	"github.com/just1689/entity-sync/db"
	"github.com/just1689/entity-sync/dq"
	"github.com/just1689/entity-sync/shared"
	"github.com/just1689/entity-sync/web"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const nsqAddr = "127.0.0.1:4150"
const entityType shared.EntityType = "items"

var GlobalBridge *bridge.Bridge

var Name string
var listenLocal = flag.String("listen", ":8080", "listen addr: :8080")

func main() {
	flag.Parse()

	mux := http.NewServeMux()
	l, err := net.Listen("tcp", *listenLocal)
	if err != nil {
		logrus.Fatalln(err)
	}

	//Build the bridge
	GlobalBridge = bridge.BuildBridge(dq.BuildPublisher(nsqAddr), dq.BuildSubscriber(nsqAddr))

	//Create publisher for NSQ (Allows to call NotifyAllOfChange())
	GlobalBridge.CreateQueuePublishers(entityType)

	//Ensure the bridge will send NSQ messages for entityType to onNotify
	GlobalBridge.Subscribe(entityType)

	//Tell the databaseHub how to fetch an entity with (and any other rows related to) rowKey
	db.GlobalDatabaseHub.AddUpdateHandler(entityType, func(rowKey shared.EntityKey, sender shared.ByteHandler) {
		item := fetch(rowKey)
		b, err := json.Marshal(item)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		sender(b)
	})

	web.HandleEntity(mux, GlobalBridge)

	resolveName(GlobalBridge)

	logrus.Println("Starting serve on ", *listenLocal)
	err = http.Serve(l, mux)
	if err != nil {
		panic(err)
	}

}

type ItemV1 struct {
	ID         string
	Closed     bool
	ClosedDate time.Time
}

func resolveName(b *bridge.Bridge) {
	role := os.Getenv("role")
	if role == "1" {
		Name = fmt.Sprint("Reader", rand.Intn(100))
	} else if role == "2" {
		Name = fmt.Sprint("Mutator", rand.Intn(100))
		startMutator(b)
	} else {
		panic(errors.New("role can be 1 or 2 and nothing else"))
	}
}

func startMutator(b *bridge.Bridge) {
	go func() {
		key := shared.EntityKey{
			ID:     "100",
			Entity: entityType,
		}
		for {
			exampleMutex.Lock()
			exampleItem.ClosedDate = time.Now()
			exampleMutex.Unlock()
			logrus.Println("Mutated", key.ID)
			logrus.Println("NotifyAllOfChange(", key.Hash(), ")")
			b.NotifyAllOfChange(key)
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

var exampleMutex sync.Mutex
var exampleItem = ItemV1{
	ID:         "",
	Closed:     true,
	ClosedDate: time.Now(),
}

func fetch(rowKey shared.EntityKey) ItemV1 {
	exampleMutex.Lock()
	defer exampleMutex.Unlock()
	exampleItem.ID = rowKey.ID
	return exampleItem
}
