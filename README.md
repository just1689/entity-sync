# Entity Sync

Example:

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