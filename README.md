# Entity Sync

<img src="https://goreportcard.com/badge/github.com/just1689/entity-sync">&nbsp;<a href="https://codebeat.co/projects/github-com-just1689-entity-sync-master"><img alt="codebeat badge" src="https://codebeat.co/badges/db75c6df-77e3-4f84-9464-ca1d2062566c" /></a>&nbsp;<a href="https://codeclimate.com/github/just1689/entity-sync/maintainability"><img src="https://api.codeclimate.com/v1/badges/4ccbe11fba6a8037fa76/maintainability" /></a>
<br />
Push entities to websocket clients onchange to keep clients in sync.

<img src="docs/diagram-v2.svg">


## Features

- Stateless server. Servers do not need to know about each other or which clients are connected to other servers. This allows the server to scale without synchronizing them.
- When you change something on the server side, provide the EntityKey to the bridge and all clients will be pushed the entity.
- Only one lookup per server on change.
- Multiple subscriptions. Each client can subscribe to multiple entities and multiples keys in each entity. 
- Multiple responses. You can send back several rows. This is great if updating the client means sending them rows from tables in foreign keys etc.
- Database / repository agnostic. This library can take a function that you implement to use whichever database, driver, client or interface you choose to implement. 

## Roadmap
- Add a secret to a client. Accept a secret from the ws and set in client state. Pass secret to the handler to ensure the user may request the KeyEntity they ask for.
- Provide a method for incoming websocket requests that don't match any concern for this library.


## Example

### Server setup
Connect the server to EntitySync. Wire the your mux to the bridge and provide a method that can resolve an `EntityKey`.
```go
//A standard mux and http listener
mux := http.NewServeMux()

//Tell the databaseHub how to fetch an entity with (and any other rows related to) rowKey
var databaseHub *esdb.DatabaseHub = esdb.NewDatabaseHub()
databaseHub.AddDataPullAndPushHandler(entityType, func(rowKey shared.EntityKey, pusher shared.ByteHandler) {
    item := fetch(rowKey)
    b, _ := json.Marshal(item)
    pusher(b)
})

// The bridge matches communication from ws to nsq and from nsq to ws.
// It also calls on the db to resolve entityKey
var bridge *esbridge.Bridge = esbridge.BuildBridge(
    esq.BuildPublisher(*nsqAddr),
    esq.BuildSubscriber(*nsqAddr),
    databaseHub.PullDataAndPush,
)

//Create publisher for NSQ (Allows to call NotifyAllOfChange())
bridge.CreateQueuePublishers(entityType)

//Ensure the bridge will send NSQ messages for entityType to onNotify
bridge.Subscribe(entityType)

//Pass the mux and a client builder to the libraries handlers
esweb.SetupMuxBridge(mux, bridge.ClientBuilder)


var l net.Listener
if l, err = net.Listen("tcp", *listenLocal); err != nil {
...
}
http.Serve(l, mux)
```

### Connect clients
Connect any number of clients:
1. Connect to the server over websocket ws://host:port/ws/entity-sync/
2. Send a subscription request
 
```json
{
    "action": "subscribe",
    "entityKey": {
        "id": "100",
        "entity": "items"
    }
}
```
### Mutate entity & notify

Make some change to the item in question where it is persisted and then call
`bridge.NotifyAllOfChange(entityKey)` where entityKey is a `shared.EntityKey`.

All connected clients over websockets will receive messages for the EntityKey/s to which they are subscribed.


