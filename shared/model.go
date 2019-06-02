package shared

type ByteHandler func([]byte)

type RowKey struct {
	Entity string
	ID     string
}
