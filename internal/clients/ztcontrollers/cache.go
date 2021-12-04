package ztcontrollers

import (
	"fmt"

	"github.com/sargassum-eco/fluitans/pkg/framework/clientcache"
)

type Cache struct {
	Cache clientcache.Cache
}

// /network/controllers/:address

func keyControllerByAddress(address string) string {
	return fmt.Sprintf("/controllers/[%s]", address)
}

func (c *Cache) SetControllerByAddress(address string, ztController Controller) error {
	key := keyControllerByAddress(address)
	return c.Cache.SetEntry(key, ztController, ztController.NetworkCostWeight, -1)
}

func (c *Cache) UnsetControllerByAddress(address string) {
	key := keyControllerByAddress(address)
	c.Cache.UnsetEntry(key)
}

func (c *Cache) GetControllerByAddress(address string) (*Controller, bool, error) {
	key := keyControllerByAddress(address)
	var value Controller
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, keyExists, err
	}

	return &value, true, nil
}
