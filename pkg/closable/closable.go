package closable

type Closer interface {
	Close() error
}
