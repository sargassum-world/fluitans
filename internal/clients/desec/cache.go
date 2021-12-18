package desec

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/framework/clientcache"
)

type Cache struct {
	Cache       clientcache.Cache
	CostWeight  float32
	TTL         time.Duration
	RecordTypes []string
}

// /dns/domains/:name

func keyDomainByName(name string) string {
	return fmt.Sprintf("/dns/domains/n:[%s]", name)
}

func (c *Cache) SetDomainByName(name string, domain desec.Domain) error {
	key := keyDomainByName(name)
	return c.Cache.SetEntry(key, domain, c.CostWeight, c.TTL)
}

func (c *Cache) SetNonexistentDomainByName(name string) {
	key := keyDomainByName(name)
	c.Cache.SetNonexistentEntry(key, c.CostWeight, c.TTL)
}

func (c *Cache) GetDomainByName(name string) (*desec.Domain, bool, error) {
	key := keyDomainByName(name)
	var value desec.Domain
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, keyExists, err
	}

	return &value, true, nil
}

// /dns/domains/:name/subnames

func keySubnames(domainName string) string {
	return fmt.Sprintf("/dns/domains/n:[%s]/subnames", domainName)
}

func (c *Cache) SetSubnames(domainName string, subnames []string) error {
	key := keySubnames(domainName)
	return c.Cache.SetEntry(key, subnames, c.CostWeight, c.TTL)
}

func (c *Cache) UnsetSubnames(domainName string) {
	key := keySubnames(domainName)
	c.Cache.UnsetEntry(key)
}

func (c *Cache) GetSubnames(domainName string) ([]string, error) {
	key := keySubnames(domainName)
	var value []string
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, err
	}

	return value, nil
}

func (c *Cache) HasSubname(domainName string, subname string) bool {
	cachedSubnames, err := c.GetSubnames(domainName)
	if err != nil {
		return false
	}

	for _, s := range cachedSubnames {
		if s == subname {
			return true
		}
	}
	return false
}

// /dns/domains/:domain/rrsets/:subname

func (c *Cache) SetRRsetsByName(domainName, subname string, rrsets []desec.RRset) error {
	cacheableRRsets := filterRRsets(rrsets, c.RecordTypes)
	for _, recordType := range c.RecordTypes {
		rrset, hasRRset := cacheableRRsets[recordType]
		if !hasRRset {
			c.SetNonexistentRRsetByNameAndType(domainName, subname, recordType)
			continue
		}

		err := c.SetRRsetByNameAndType(domainName, subname, recordType, rrset)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf(
				"couldn't set cache entry for the %s RRset for %s.%s", recordType, subname, domainName,
			))
		}
	}
	return nil
}

func (c *Cache) GetRRsetsByName(domainName, subname string) ([]desec.RRset, error) {
	rrsets := make([]desec.RRset, 0, len(c.RecordTypes))
	for _, recordType := range c.RecordTypes {
		key := keyRRsetByNameAndType(domainName, subname, recordType)
		var value desec.RRset
		keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf(
				"couldn't get cache entry for the %s RRset for %s.%s", recordType, subname, domainName,
			))
		}

		if !keyExists {
			return nil, nil // missing recordtype key should be considered a cache miss
		}

		if valueExists {
			rrsets = append(rrsets, value)
		}
	}

	return rrsets, nil
}

// /dns/domains/:domain/rrsets/:subname/:type

func keyRRsetByNameAndType(domainName, subname, rrsetType string) string {
	return fmt.Sprintf("/dns/domains/n:[%s]/rrsets/sn:[%s]/t:[%s]", domainName, subname, rrsetType)
}

func (c *Cache) SetRRsetByNameAndType(
	domainName, subname, rrsetType string, rrset desec.RRset,
) error {
	key := keyRRsetByNameAndType(domainName, subname, rrsetType)
	return c.Cache.SetEntry(key, rrset, c.CostWeight, c.TTL)
}

func (c *Cache) SetNonexistentRRsetByNameAndType(domainName, subname, rrsetType string) {
	key := keyRRsetByNameAndType(domainName, subname, rrsetType)
	c.Cache.SetNonexistentEntry(key, c.CostWeight, c.TTL)
}

func (c *Cache) GetRRsetByNameAndType(
	domainName, subname, rrsetType string,
) (*desec.RRset, bool, error) {
	key := keyRRsetByNameAndType(domainName, subname, rrsetType)
	var value desec.RRset
	keyExists, valueExists, err := c.Cache.GetEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, keyExists, err
	}

	return &value, true, nil
}
