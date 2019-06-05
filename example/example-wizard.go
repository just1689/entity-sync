package main

import (
	"encoding/json"
	"github.com/just1689/entity-sync/entitysync"
	"github.com/just1689/entity-sync/entitysync/shared"
)

func main() {
	//Nonsense
	setup()
	done := make(chan bool)

	config := entitysync.Config{
		ListenAddr: *listenLocal,
		NSQAddr:    *nsqAddr,
	}
	entitySync := entitysync.Setup(config)
	entitysync.HandleEntity(entitySync, entityType, func(entityKey shared.EntityKey, handler shared.ByteHandler) {
		item := fetch(entityKey)
		b, _ := json.Marshal(item)
		handler(b)
	})

	<-done
}
