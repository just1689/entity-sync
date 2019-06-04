package esnsq

import (
	"github.com/just1689/entity-sync/shared"
)

//BuildPublisher can be called to give a method that in turn can be called to create publishers
var BuildPublisher shared.AddressableEntityHandler = func(addr string) shared.EntityHandler {
	return func(entityType shared.EntityType) shared.ByteHandler {
		return GetNSQProducer(addr, entityType)
	}
}

var BuildSubscriber shared.AddressableEntityByteHandler = func(addr string) shared.EntityByteHandler {
	return func(entityType shared.EntityType, callback shared.ByteHandler) {
		SubscribeNSQ(addr, entityType, callback)
	}
}
