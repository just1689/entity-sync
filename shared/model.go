package shared

import "fmt"

type ByteHandler func([]byte)
type EntityByteHandler func(entity EntityType, handler ByteHandler)
type EntityHandler func(entity EntityType) ByteHandler

type EntityType string

type EntityKey struct {
	Entity EntityType `json:"entity"`
	ID     string     `json:"id"`
}

func (r *EntityKey) Hash() string {
	return fmt.Sprint(r.Entity, ".", r.ID)
}
