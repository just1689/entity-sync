package esdb

import (
	"github.com/just1689/entity-sync/es/shared"
)

func NewDatabaseHub() *DatabaseHub {
	return &DatabaseHub{
		handlers: make(map[shared.EntityType]shared.EntityKeySecretByteHandler),
	}
}

type DatabaseHub struct {
	handlers map[shared.EntityType]shared.EntityKeySecretByteHandler
}

func (d *DatabaseHub) AddDataPullAndPushHandler(entityType shared.EntityType, client shared.EntityKeySecretByteHandler) {
	d.handlers[entityType] = client

}

func (d *DatabaseHub) PullDataAndPush(key shared.EntityKey, secret string, push shared.ByteHandler) {
	handlerUpdateClient, found := d.handlers[key.Entity]
	if found {
		handlerUpdateClient(key, secret, push)
	}
}
