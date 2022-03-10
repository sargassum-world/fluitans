package clientcache

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
)

type RistrettoCache struct {
	cache      *ristretto.Cache
	marshaller Marshaller
}

func NewRistrettoCache(cacheConfig ristretto.Config) (Cache, error) {
	cache, err := ristretto.NewCache(&cacheConfig)
	if err != nil {
		return nil, err
	}

	marshaller := NewMsgPackMarshaller()

	return &RistrettoCache{
		cache:      cache,
		marshaller: &marshaller,
	}, nil
}

func computeCacheCost(costWeight float32, bytes []byte) int64 {
	return int64(float64(costWeight) * float64(len(bytes)))
}

func (c *RistrettoCache) SetEntry(
	key string, value interface{}, costWeight float32, ttl time.Duration,
) error {
	marshaled, err := c.marshaller.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "couldn't marshal value for key %s", key)
	}

	if ttl < 0 {
		c.cache.Set(key, marshaled, computeCacheCost(costWeight, marshaled))
	} else {
		c.cache.SetWithTTL(key, marshaled, computeCacheCost(costWeight, marshaled), ttl)
	}
	return nil
}

func (c *RistrettoCache) UnsetEntry(key string) {
	c.cache.Del(key)
}

type nonexistentValue struct{}

func (c *RistrettoCache) SetNonexistentEntry(key string, cost float32, ttl time.Duration) {
	// Put a tombstone in the cache so the key has a cache hit indicating lack of a value
	if ttl < 0 {
		c.cache.Set(key, nonexistentValue{}, int64(cost))
	} else {
		c.cache.SetWithTTL(key, nonexistentValue{}, int64(cost), ttl)
	}
}

func (c *RistrettoCache) GetEntry(key string, value interface{}) (bool, bool, error) {
	entryRaw, hasKey := c.cache.Get(key)
	if !hasKey {
		// fmt.Printf("Cache miss: %s\n", key)
		return false, false, nil
	}

	switch marshaledBytes := entryRaw.(type) {
	default:
		return true, false, errors.Errorf("cache entry %s has unexpected type %T", key, entryRaw)
	case nonexistentValue:
		return true, false, nil
	case []byte:
		if err := c.marshaller.Unmarshal(marshaledBytes, value); err != nil {
			return true, true, errors.Wrapf(err, "couldn't unmarshal value for key %s", key)
		}

		return true, true, nil
	}
}
