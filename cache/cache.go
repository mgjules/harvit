package cache

import (
	"encoding/json"

	"github.com/dgraph-io/ristretto"
)

const _defaultBufferItems = 64

// Cache is a simple wrapper around ristretto.Cache.
type Cache struct {
	*ristretto.Cache
}

// New creates a new Cache.
func New(maxKeys, maxCost int64) (*Cache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: maxKeys,
		MaxCost:     maxCost,
		BufferItems: _defaultBufferItems,
		Cost: func(value interface{}) int64 {
			test, err := json.Marshal(value)
			if err != nil {
				return 1
			}

			return int64(len(test))
		},
	})
	if err != nil {
		return nil, err
	}

	return &Cache{cache}, nil
}
