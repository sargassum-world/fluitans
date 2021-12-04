package clientcache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
)

type RistrettoCache struct {
	cache *ristretto.Cache
}

func NewRistrettoCache(cacheConfig ristretto.Config) (Cache, error) {
	cache, err := ristretto.NewCache(&cacheConfig)
	if err != nil {
		return nil, err
	}

	return &RistrettoCache{cache: cache}, nil
}

func computeCacheCost(costWeight float32, bytes []byte) int64 {
	return int64(float64(costWeight) * float64(len(bytes)))
}

func (c *RistrettoCache) SetEntry(
	key string, value interface{}, costWeight float32, ttl time.Duration,
) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return errors.Wrap(err, fmt.Sprintf("couldn't gob-encode value %+v for key %s", value, key))
	}

	if ttl < 0 {
		gobbed := buf.Bytes()
		c.cache.Set(key, gobbed, computeCacheCost(costWeight, gobbed))
	} else {
		gobbed := buf.Bytes()
		c.cache.SetWithTTL(key, gobbed, computeCacheCost(costWeight, gobbed), ttl)
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
		return false, false, nil
	}

	switch entryGob := entryRaw.(type) {
	default:
		return true, false, fmt.Errorf("invalid cache entry %s has unexpected type %T", key, entryRaw)
	case nonexistentValue:
		return true, false, nil
	case []byte:
		buf := bytes.NewBuffer(entryGob)
		dec := gob.NewDecoder(buf)
		if err := dec.Decode(value); err != nil {
			return true, true, errors.Wrap(err, fmt.Sprintf(
				"couldn't gob-decode value %+v for key %s", buf, key,
			))
		}

		return true, true, nil
	}
}
