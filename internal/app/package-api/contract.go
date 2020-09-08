package packageapi

// Storage describes the storage
type Storage interface {
	Close(delete bool) error
	Keys() ([]string, error)
	Get(key string) ([]byte, error)
	GetMultipleBySuffix(suffix string) ([]string, [][]byte, error)
	Put(key string, val []byte) error
	Delete(key string) error
	Purge() error
}
