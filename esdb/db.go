package esdb

import "github.com/just1689/entity-sync/shared"

func NewDatabaseHub() *DatabaseHub {
	return &DatabaseHub{
		handlers: make(map[shared.EntityType]HandleUpdateClient),
	}
}

type HandleUpdateClient func(rowKey shared.EntityKey, sender shared.ByteHandler)

type DatabaseHub struct {
	handlers map[shared.EntityType]HandleUpdateClient
}

func (d *DatabaseHub) AddDataPullAndPushHandler(entityType shared.EntityType, client HandleUpdateClient) {
	d.handlers[entityType] = client

}

func (d *DatabaseHub) PullDataAndPush(key shared.EntityKey, push shared.ByteHandler) {
	handlerUpdateClient, found := d.handlers[key.Entity]
	if found {
		handlerUpdateClient(key, push)
	}
}
