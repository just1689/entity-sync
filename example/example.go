package main

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/just1689/entity-sync/entitysync"
	"github.com/just1689/entity-sync/entitysync/esbridge"
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

	// Provide a configuration
	config := entitysync.Config{
		Mux:     http.NewServeMux(),
		NSQAddr: *nsqAddr,
	}
	//Setup entitySync with that configuration
	es := entitysync.Setup(config)

	//Register an entity and tell the library how to fetch and what to write to the client
	es.RegisterEntityAndDBHandler(entityType, func(entityKey shared.EntityKey, secret string, handler shared.ByteHandler) {
		item := fetch(entityKey)
		b, _ := json.Marshal(item)
		handler(b)
	})

	//Simulate data changes
	startMutator(es.Bridge)

	//Start a listener and provide the mux for routes / handling
	logrus.Println("Starting serve on ", *listenLocal)
	var l net.Listener
	var err error
	if l, err = net.Listen("tcp", *listenLocal); err != nil {
		panic(err)
	}
	if err = http.Serve(l, config.Mux); err != nil {
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
	if *role == "2" {
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
