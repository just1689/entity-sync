package db

import "github.com/just1689/entity-sync/shared"

var GlobalDatabaseHub = &DatabaseHub{
	handlers: make(map[string]HandleUpdateClient),
}

type HandleUpdateClient func(rowKey shared.RowKey, sender shared.ByteHandler)

type DatabaseHub struct {
	handlers map[string]HandleUpdateClient
}

func (d *DatabaseHub) AddUpdateHandler(entity string, client HandleUpdateClient) {
	d.handlers[entity] = client

}

func (d *DatabaseHub) ProcessUpdateHandler(key shared.RowKey, sender shared.ByteHandler) {
	handlerUpdateClient, found := d.handlers[key.Entity]
	if found {
		handlerUpdateClient(key, sender)
	}
}
