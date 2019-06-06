package main

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/just1689/entity-sync/entitysync/esbridge"
	"github.com/just1689/entity-sync/entitysync/esdb"
	"github.com/just1689/entity-sync/entitysync/esq"
	"github.com/just1689/entity-sync/entitysync/esweb"
	"github.com/just1689/entity-sync/entitysync/shared"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const entityType shared.EntityType = "items"
const entityID string = "100"

var listenLocal = flag.String("listen", ":8080", "listen addr: :8080")
var role = flag.String("role", "2", "Role is 1 or 2. 2 Mutates the row for example purposes")
var nsqAddr = flag.String("nsqd", "192.168.88.26:30000", "nsqd address")

func main() {
	//Nonsense
	setup()
	var err error

	//A standard mux and http listener
	mux := http.NewServeMux()

	//Tell the databaseHub how to fetch an entity with (and any other rows related to) rowKey
	databaseHub := esdb.NewDatabaseHub()
	databaseHub.AddDataPullAndPushHandler(entityType, func(rowKey shared.EntityKey, pusher shared.ByteHandler) {
		item := fetch(rowKey)
		b, _ := json.Marshal(item)
		pusher(b)
	})

	// The bridge matches communication from ws to nsq and from nsq to ws.
	// It also calls on the db to resolve entityKey
	bridge := esbridge.BuildBridge(
		esq.BuildPublisher(*nsqAddr),
		esq.BuildSubscriber(*nsqAddr),
		databaseHub.PullDataAndPush,
	)

	//Sync entity will ensure that you can PUB and SUB for this entity type
	bridge.RegisterEntityForSync(entityType)

	//Pass the mux and a client builder to the libraries handlers
	esweb.SetupMuxBridge(mux, bridge.ClientBuilder)

	if *role == "2" {
		startMutator(bridge)
	}

	logrus.Println("Starting serve on ", *listenLocal)
	var l net.Listener
	if l, err = net.Listen("tcp", *listenLocal); err != nil {
		logrus.Fatalln(err)
	}
	if err = http.Serve(l, mux); err != nil {
		panic(err)
	}

}

func setup() {
	flag.Parse()
	l := os.Getenv("listen")
	if l != "" {
		listenLocal = &l
	}
	n := os.Getenv("nsqd")
	if n != "" {
		nsqAddr = &n
	}

	t := os.Getenv("role")
	role = &t
	if *role != "1" && *role != "2" {
		panic(errors.New("role can be 1 or 2 and nothing else"))
	}
}

//Some type
type ItemV1 struct {
	ID         string
	Closed     bool
	ClosedDate time.Time
}

//This simulates the data changing at an interval. This could be replaced with your API etc.
func startMutator(b *esbridge.Bridge) {
	go func() {
		key := shared.EntityKey{
			ID:     entityID,
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

//Simulate a thread safe data store
func fetch(rowKey shared.EntityKey) ItemV1 {
	exampleMutex.Lock()
	defer exampleMutex.Unlock()
	exampleItem.ID = rowKey.ID
	return exampleItem
}
