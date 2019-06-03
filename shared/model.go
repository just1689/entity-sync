package shared

import "fmt"

type ByteHandler func([]byte)
type EntityByteHandler func(entity EntityType, handler ByteHandler)
type EntityHandler func(entity EntityType) ByteHandler
type EntityKeyHandler func(entityKey EntityKey)

type EntityType string

func (e EntityType) GetQueueName() string {
	return fmt.Sprint(e, "")
}

type EntityKey struct {
	Entity EntityType `json:"entity"`
	ID     string     `json:"id"`
}

func (r *EntityKey) Hash() string {
	return fmt.Sprint(r.Entity, ".", r.ID)
}

type Action string

const ActionSubscribe Action = "subscribe"
const ActionUnSubscribe Action = "unsubscribe"

type MessageAction struct {
	Action    Action    `json:"action"`
	EntityKey EntityKey `json:"entityKey"`
}
