package zerotier

import (
	"fmt"

	"github.com/sargassum-eco/fluitans/pkg/framework/clientcache"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type Cache struct {
	Cache      clientcache.Cache
	CostWeight float32
}

// /zerotier/networks/:id

func keyNetworkByID(id string) string {
	return fmt.Sprintf("/zerotier/networks/id:[%s]", id)
}

func (c *Cache) SetNetworkByID(id string, network zerotier.ControllerNetwork) error {
	key := keyNetworkByID(id)
	return c.Cache.SetEntry(key, network, c.CostWeight, -1)
}

func (c *Cache) UnsetNetworkByID(id string) {
	key := keyNetworkByID(id)
	c.Cache.UnsetEntry(key)
}

func (c *Cache) SetNonexistentNetworkByID(id string) {
	key := keyNetworkByID(id)
	c.Cache.SetNonexistentEntry(key, c.CostWeight, -1)
}

func (c *Cache) GetNetworkByID(id string) (*zerotier.ControllerNetwork, bool, error) {
	key := keyNetworkByID(id)
	var value zerotier.ControllerNetwork
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, keyExists, err
	}

	return &value, true, nil
}

// /zerotier/networks/:id/members

func keyNetworkMembersByID(networkID string) string {
	return fmt.Sprintf("/zerotier/networks/id:[%s]/members", networkID)
}

func (c *Cache) SetNetworkMembersByID(networkID string, members []string) error {
	key := keyNetworkMembersByID(networkID)
	return c.Cache.SetEntry(key, members, c.CostWeight, -1)
}

func (c *Cache) UnsetNetworkMembersByID(networkID string) {
	key := keyNetworkMembersByID(networkID)
	c.Cache.UnsetEntry(key)
}

func (c *Cache) GetNetworkMembersByID(networkID string) ([]string, error) {
	key := keyNetworkMembersByID(networkID)
	value := make([]string, 0)
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, err
	}

	return value, nil
}
