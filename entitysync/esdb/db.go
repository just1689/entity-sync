package esdb

import (
	"github.com/just1689/entity-sync/entitysync/shared"
)

func NewDatabaseHub() *DatabaseHub {
	return &DatabaseHub{
		handlers: make(map[shared.EntityType]shared.EntityKeyByteHandler),
	}
}

type DatabaseHub struct {
	handlers map[shared.EntityType]shared.EntityKeyByteHandler
}

func (d *DatabaseHub) AddDataPullAndPushHandler(entityType shared.EntityType, client shared.EntityKeyByteHandler) {
	d.handlers[entityType] = client

}

func (d *DatabaseHub) PullDataAndPush(key shared.EntityKey, push shared.ByteHandler) {
	handlerUpdateClient, found := d.handlers[key.Entity]
	if found {
		handlerUpdateClient(key, push)
	}
}
