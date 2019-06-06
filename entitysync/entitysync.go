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
	result := EntitySync{
		//Tell the databaseHub how to fetch an entity with (and any other rows related to) rowKey
		DatabaseHub: esdb.NewDatabaseHub(),
	}

	// The bridge matches communication from ws to nsq and from nsq to ws.
	// It also calls on the db to resolve entityKey
	result.Bridge = esbridge.BuildBridge(
		esq.BuildPublisher(config.NSQAddr),
		esq.BuildSubscriber(config.NSQAddr),
		result.DatabaseHub.PullDataAndPush,
	)

	//Pass the mux and a client builder to the libraries handlers
	esweb.SetupMuxBridge(config.Mux, result.Bridge.ClientBuilder)

	return result
}

func (es *EntitySync) RegisterEntityAndDBHandler(entityType shared.EntityType, databaseFetchAndPush shared.EntityKeySecretByteHandler) {
	es.DatabaseHub.AddDataPullAndPushHandler(entityType, databaseFetchAndPush)
	es.Bridge.RegisterEntityForSync(entityType)
}
