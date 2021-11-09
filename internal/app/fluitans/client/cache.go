package client

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/dgraph-io/ristretto"
)

type Cache struct {
	*ristretto.Cache
}

func (c *Cache) SetControllerByAddress(address string, controller Controller) error {
	key := fmt.Sprintf("controllers/%s", address)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(controller); err != nil {
		return err
	}

	controllerBytes := buf.Bytes()
	c.Cache.Set(
		key, controllerBytes,
		int64(float64(controller.NetworkCostWeight)*float64(len(controllerBytes))),
	)
	return nil
}

func (c *Cache) UnsetControllerByAddress(address string) {
	key := fmt.Sprintf("controllers/%s", address)
	c.Cache.Del(key)
}

func (c *Cache) GetControllerByAddress(address string) (*Controller, bool, error) {
	key := fmt.Sprintf("controllers/%s", address)
	controllerServer, hasKey := c.Cache.Get(key)
	if !hasKey {
		return nil, false, nil
	}

	switch controllerGob := controllerServer.(type) {
	default:
		return nil, true, fmt.Errorf(
			"invalid cache key %s has unexpected type %T", key, controllerServer,
		)
	case []byte:
		var controller Controller
		buf := bytes.NewBuffer(controllerGob)
		dec := gob.NewDecoder(buf)
		if err := dec.Decode(&controller); err != nil {
			return nil, true, err
		}

		return &controller, true, nil
	}
}
