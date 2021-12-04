package client

import (
	"fmt"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
)

// /network/controllers/:address

func keyControllerByAddress(address string) string {
	return fmt.Sprintf("/network/controllers/[%s]", address)
}

func (c *Cache) SetControllerByAddress(
	address string, controller models.Controller,
) error {
	key := keyControllerByAddress(address)
	return c.setEntry(key, controller, controller.NetworkCostWeight, -1)
}

func (c *Cache) UnsetControllerByAddress(address string) {
	key := keyControllerByAddress(address)
	c.unsetEntry(key)
}

func (c *Cache) GetControllerByAddress(
	address string,
) (*models.Controller, bool, error) {
	key := keyControllerByAddress(address)
	var value models.Controller
	keyExists, valueExists, err := c.getEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, keyExists, err
	}

	return &value, true, nil
}
