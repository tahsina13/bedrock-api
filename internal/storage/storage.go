package storage

// KVStorage represents a key-value storage system.
type KVStorage interface {
	// Set sets the value for a given key.
	Set(key string, value []byte) error
	// Get retrieves the value for a given key.
	Get(key string) ([]byte, error)
	// Delete removes the key-value pair for a given key.
	Delete(key string) error
	// List retrieves the values for a given prefix.
	List(prefix string) ([][]byte, error)
}
