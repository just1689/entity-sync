package main

import (
	"encoding/json"
	"github.com/just1689/entity-sync/eswizard"
	"github.com/just1689/entity-sync/shared"
)

func main() {
	//Nonsense
	setup()
	done := make(chan bool)

	config := eswizard.Config{
		ListenAddr: *listenLocal,
		NSQAddr:    *nsqAddr,
	}
	entitySync := eswizard.Setup(config)
	eswizard.HandleEntity(entitySync, entityType, func(entityKey shared.EntityKey, handler shared.ByteHandler) {
		item := fetch(entityKey)
		b, _ := json.Marshal(item)
		handler(b)
	})

	<-done
}
