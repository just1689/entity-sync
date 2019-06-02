package shared

import "fmt"

type ByteHandler func([]byte)
type EntityByteHandler func(entity Entity, handler ByteHandler)
type EntityHandler func(entity Entity) ByteHandler

type Entity string

type RowKey struct {
	Entity Entity `json:"entity"`
	ID     string `json:"id"`
}

func (r *RowKey) Hash() string {
	return fmt.Sprint(r.Entity, ".", r.ID)
}
