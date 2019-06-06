package entitysync

import (
	"github.com/just1689/entity-sync/entitysync/esbridge"
	"github.com/just1689/entity-sync/entitysync/esdb"
	"github.com/just1689/entity-sync/entitysync/esq"
	"github.com/just1689/entity-sync/entitysync/esweb"
	"github.com/just1689/entity-sync/entitysync/shared"
	"net/http"
)

type Config struct {
	NSQAddr string
	Mux     *http.ServeMux
}

type EntitySync struct {
	Bridge      *esbridge.Bridge
	DatabaseHub *esdb.DatabaseHub
}

func Setup(config Config) EntitySync {

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
	esweb.SetupMuxBridge(config.Mux, bridge.ClientBuilder)

	return EntitySync{
		Bridge:      bridge,
		DatabaseHub: databaseHub,
	}
}

func HandleEntity(es EntitySync, entityType shared.EntityType, databaseFetchAndPush shared.EntityKeyByteHandler) {
	es.DatabaseHub.AddDataPullAndPushHandler(entityType, databaseFetchAndPush)
	es.Bridge.SyncEntityType(entityType)
}
