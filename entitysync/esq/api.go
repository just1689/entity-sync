package esq

import (
	shared2 "github.com/just1689/entity-sync/entitysync/shared"
)

/*
	If you would like to use something other than NSQ, you can provide your own implementation of
	`shared.AddressableEntityHandler` and `shared.AddressableEntityByteHandler`.
	This must be provided to `esbridge.BuildBridge()`
*/

//BuildPublisher can be called to give a method that in turn can be called to create publishers
var BuildPublisher shared2.AddressableEntityHandler = func(addr string) shared2.EntityHandler {
	return func(entityType shared2.EntityType) shared2.ByteHandler {
		return getNSQProducer(addr, entityType)
	}
}

var BuildSubscriber shared2.AddressableEntityByteHandler = func(addr string) shared2.EntityByteHandler {
	return func(entityType shared2.EntityType, callback shared2.ByteHandler) {
		subscribeNSQ(addr, entityType, callback)
	}
}
