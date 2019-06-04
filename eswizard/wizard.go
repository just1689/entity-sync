package eswizard

import (
	"github.com/just1689/entity-sync/esbridge"
	"github.com/just1689/entity-sync/esdb"
	"github.com/just1689/entity-sync/esq"
	"github.com/just1689/entity-sync/esweb"
	"github.com/just1689/entity-sync/shared"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
)

type Config struct {
	ListenAddr string
	NSQAddr    string
}

type EntitySync struct {
	Bridge      *esbridge.Bridge
	DatabaseHub *esdb.DatabaseHub
}

func Setup(config Config) EntitySync {
	var err error
	//A standard mux and http listener
	mux := http.NewServeMux()

	//Tell the databaseHub how to fetch an entity with (and any other rows related to) rowKey
	databaseHub := esdb.NewDatabaseHub()

	// The bridge matches communication from ws to nsq and from nsq to ws.
	// It also calls on the db to resolve entityKey
	bridge := esbridge.BuildBridge(
		esq.BuildPublisher(config.NSQAddr),
		esq.BuildSubscriber(config.NSQAddr),
		databaseHub.PullDataAndPush,
	)

	//Pass the mux and a client builder to the libraries handlers
	esweb.SetupMuxBridge(mux, bridge.ClientBuilder)

	logrus.Println("Starting serve on ", config.ListenAddr)
	var l net.Listener
	if l, err = net.Listen("tcp", config.ListenAddr); err != nil {
		logrus.Fatalln(err)
	}
	go func() {
		if err = http.Serve(l, mux); err != nil {
			panic(err)
		}
	}()

	return EntitySync{
		Bridge:      bridge,
		DatabaseHub: databaseHub,
	}
}

func HandleEntity(es EntitySync, entityType shared.EntityType, databaseFetchAndPush shared.EntityKeyByteHandler) {
	es.DatabaseHub.AddDataPullAndPushHandler(entityType, databaseFetchAndPush)
	es.Bridge.SyncEntityType(entityType)
}
