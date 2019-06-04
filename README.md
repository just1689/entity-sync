# Entity Sync

## Example

Connect the server to EntitySync. Wire the your mux to the bridge and provide a method that can resolve an `EntityKey`.
```go

//Create your own mux
mux := http.NewServeMux()
l, err := net.Listen("tcp", *listenLocal)
if err != nil {
    logrus.Fatalln(err)
}

//Build the bridge
// The bridge matches communication from ws to nsq and from nsq to ws. It also calls on the db to resolve entityKey
GlobalBridge = bridge.BuildBridge(dq.BuildPublisher(nsqAddr), dq.BuildSubscriber(nsqAddr), db.GlobalDatabaseHub.ProcessUpdateHandler)

//Create publisher for NSQ (Allows to call NotifyAllOfChange())
GlobalBridge.CreateQueuePublishers(entityType)

//Ensure the bridge will send NSQ messages for entityType to onNotify
GlobalBridge.Subscribe(entityType)

//Tell the databaseHub how to fetch an entity with (and any other rows related to) rowKey
db.GlobalDatabaseHub.AddUpdateHandler(entityType, func(rowKey shared.EntityKey, sender shared.ByteHandler) {
    item := ...
    b, _ := json.Marshall(item)
    sender(b)
})

//Give the mux and bridge to the web handler
web.HandleEntity(mux, GlobalBridge)

//Start the listener on that mux
logrus.Println("Starting serve on ", *listenLocal)
err = http.Serve(l, mux)
if err != nil {
    panic(err)
}

```

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

On the server: make some change to the item in question, then call:
`bridge.NotifyAllOfChange(key)` where key is a `KeyEntity`.

All connected clients over websockets will receive messages for the EntityKey/s to which they are subscribed.


# Roadmap

1. Add a secret to a client. Accept a secret from the ws and set in client state. Pass secret to the handler to ensure the user may request the KeyEntity they ask for.
2. Provide a method for incoming websocket requests that don't match any concern for this library.
