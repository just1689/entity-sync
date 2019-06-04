package shared

import "fmt"

type EntityType string
type ByteHandler func([]byte)
type EntityHandler func(entity EntityType) ByteHandler
type EntityKeyHandler func(entityKey EntityKey)
type EntityByteHandler func(entity EntityType, handler ByteHandler)
type EntityKeyByteHandler func(entityKey EntityKey, handler ByteHandler)
type ByteHandlingRemoteProxy func(ByteHandler) (sub EntityKeyHandler, unSub EntityKeyHandler, dc chan bool)

func (e EntityType) GetQueueName() string {
	return string(e)
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
