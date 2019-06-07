package esq

import (
	"github.com/just1689/entity-sync/es/shared"
)

/*
	If you would like to use something other than NSQ, you can provide your own implementation of
	`shared.AddressableEntityHandler` and `shared.AddressableEntityByteHandler`.
	This must be provided to `esbridge.BuildBridge()`
*/

//BuildPublisher can be called to give a method that in turn can be called to create publishers
var BuildPublisher shared.AddressableEntityHandler = func(addr string) shared.EntityHandler {
	return func(entityType shared.EntityType) shared.ByteHandler {
		return getNSQProducer(addr, entityType)
	}
}

var BuildSubscriber shared.AddressableEntityByteHandler = func(addr string) shared.EntityByteHandler {
	return func(entityType shared.EntityType, callback shared.ByteHandler) {
		subscribeNSQ(addr, entityType, callback)
	}
}
