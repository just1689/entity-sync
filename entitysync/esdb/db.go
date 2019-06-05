package esdb

import (
	shared2 "github.com/just1689/entity-sync/entitysync/shared"
)

func NewDatabaseHub() *DatabaseHub {
	return &DatabaseHub{
		handlers: make(map[shared2.EntityType]shared2.EntityKeyByteHandler),
	}
}

type DatabaseHub struct {
	handlers map[shared2.EntityType]shared2.EntityKeyByteHandler
}

func (d *DatabaseHub) AddDataPullAndPushHandler(entityType shared2.EntityType, client shared2.EntityKeyByteHandler) {
	d.handlers[entityType] = client

}

func (d *DatabaseHub) PullDataAndPush(key shared2.EntityKey, push shared2.ByteHandler) {
	handlerUpdateClient, found := d.handlers[key.Entity]
	if found {
		handlerUpdateClient(key, push)
	}
}
