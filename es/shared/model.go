package shared

import (
	"encoding/json"
	"fmt"
)

type EntityType string
type ByteHandler func([]byte)
type EntityHandler func(entity EntityType) ByteHandler
type EntityKeyHandler func(entityKey EntityKey)
type SecretByteHandler func(secret string, b []byte)
type EntityByteHandler func(entity EntityType, handler ByteHandler)
type AddressableEntityHandler func(addr string) EntityHandler
type EntityKeySecretByteHandler func(entityKey EntityKey, secret string, handler ByteHandler)
type AddressableEntityByteHandler func(addr string) EntityByteHandler
type ByteSecretHandlingRemoteProxy func(ByteHandler, func() string) (sub EntityKeyHandler, unSub EntityKeyHandler, dc chan bool)

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
const ActionSecret Action = "secret"

type Message struct {
	Action Action          `json:"action"`
	Body   json.RawMessage `json:"body"`
}
