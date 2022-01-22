package ztcontrollers

import (
	"fmt"

	"github.com/sargassum-eco/fluitans/pkg/framework/clientcache"
)

type Cache struct {
	Cache clientcache.Cache
}

// /ztcontrollers/controllers/:server/networkIDs

func keyNetworkIDsByServer(server string) string {
	return fmt.Sprintf("/ztcontrollers/controllers/s:[%s]/networkIDs", server)
}

func (c *Cache) SetNetworkIDsByServer(
	server string, networkIDs []string, costWeight float32,
) error {
	key := keyNetworkIDsByServer(server)
	return c.Cache.SetEntry(key, networkIDs, costWeight, -1)
}

func (c *Cache) UnsetNetworkIDsByServer(server string) {
	key := keyNetworkIDsByServer(server)
	c.Cache.UnsetEntry(key)
}

func (c *Cache) GetNetworkIDsByServer(server string) ([]string, error) {
	key := keyNetworkIDsByServer(server)
	var value []string
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, err
	}

	return value, nil
}

// /ztcontrollers/controllers/:address

func keyControllerByAddress(address string) string {
	return fmt.Sprintf("/ztcontrollers/controllers/a:[%s]", address)
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

// /ztcontrollers/addresses/:server

func keyAddressByServer(server string) string {
	return fmt.Sprintf("/ztcontrollers/addresses/s:[%s]", server)
}

func (c *Cache) SetAddressByServer(server string, address string, networkCostWeight float32) error {
	key := keyAddressByServer(server)
	return c.Cache.SetEntry(key, address, networkCostWeight, -1)
}

func (c *Cache) UnsetAddressByServer(server string) {
	key := keyAddressByServer(server)
	c.Cache.UnsetEntry(key)
}

func (c *Cache) GetAddressByServer(server string) (string, bool, error) {
	key := keyAddressByServer(server)
	var value string
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return "", keyExists, err
	}

	return value, true, nil
}
