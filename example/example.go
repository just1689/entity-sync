package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/just1689/entity-sync/esbridge"
	"github.com/just1689/entity-sync/esdb"
	"github.com/just1689/entity-sync/esq"
	"github.com/just1689/entity-sync/esweb"
	"github.com/just1689/entity-sync/shared"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const nsqAddr = "192.168.88.26:30000"
const entityType shared.EntityType = "items"

var Name string
var listenLocal = flag.String("listen", ":8080", "listen addr: :8080")

func main() {
	flag.Parse()
	checkListenLocal()
	var err error

	mux := http.NewServeMux()
	var l net.Listener
	if l, err = net.Listen("tcp", *listenLocal); err != nil {
		logrus.Fatalln(err)
	}

	//Tell the databaseHub how to fetch an entity with (and any other rows related to) rowKey
	var databaseHub *esdb.DatabaseHub = esdb.NewDatabaseHub()
	databaseHub.AddDataPullAndPushHandler(entityType, func(rowKey shared.EntityKey, pusher shared.ByteHandler) {
		item := fetch(rowKey)
		b, _ := json.Marshal(item)
		pusher(b)
	})

	// The bridge matches communication from ws to nsq and from nsq to ws.
	// It also calls on the db to resolve entityKey
	var GlobalBridge *esbridge.Bridge = esbridge.BuildBridge(
		esq.BuildPublisher(nsqAddr),
		esq.BuildSubscriber(nsqAddr),
		databaseHub.PullDataAndPush,
	)

	//Create publisher for NSQ (Allows to call NotifyAllOfChange())
	GlobalBridge.CreateQueuePublishers(entityType)

	//Ensure the bridge will send NSQ messages for entityType to onNotify
	GlobalBridge.Subscribe(entityType)

	//Pass the mux and a client builder to the libraries handlers
	esweb.SetupMuxBridge(mux, GlobalBridge.ClientBuilder)

	resolveName(GlobalBridge)

	logrus.Println("Starting serve on ", *listenLocal)
	if err = http.Serve(l, mux); err != nil {
		panic(err)
	}

}

func checkListenLocal() {
	l := os.Getenv("listen")
	if l != "" {
		listenLocal = &l
	}
}

type ItemV1 struct {
	ID         string
	Closed     bool
	ClosedDate time.Time
}

func resolveName(b *esbridge.Bridge) {
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

func startMutator(b *esbridge.Bridge) {
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
			time.Sleep(1000 * time.Millisecond)
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
