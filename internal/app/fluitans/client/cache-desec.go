package client

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/desec"
)

// /dns/domains/:name

func keyDomainByName(name string) string {
	return fmt.Sprintf("/dns/domains/[%s]", name)
}

func (c *Cache) SetDomainByName(
	name string, domain desec.Domain, costWeight float32, ttl time.Duration,
) error {
	key := keyDomainByName(name)
	return c.setEntry(key, domain, costWeight, ttl)
}

func (c *Cache) SetNonexistentDomainByName(
	name string, costWeight float32, ttl time.Duration,
) {
	key := keyDomainByName(name)
	c.setNonexistentEntry(key, costWeight, ttl)
}

func (c *Cache) GetDomainByName(name string) (*desec.Domain, bool, error) {
	key := keyDomainByName(name)
	var value desec.Domain
	keyExists, valueExists, err := c.getEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, keyExists, err
	}

	return &value, true, nil
}

// /dns/domains/:name/subnames

func keySubnames(domainName string) string {
	return fmt.Sprintf("/dns/domains/[%s]/subnames", domainName)
}

func (c *Cache) SetSubnames(
	domainName string, subnames []string, costWeight float32, ttl time.Duration,
) error {
	key := keySubnames(domainName)
	return c.setEntry(key, subnames, costWeight, ttl)
}

func (c *Cache) GetSubnames(domainName string) ([]string, error) {
	key := keySubnames(domainName)
	var value []string
	keyExists, valueExists, err := c.getEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, err
	}

	return value, nil
}

// /dns/domains/:domain/rrsets/:subname

func (c *Cache) SetRRsetsByName(
	domainName, subname string,
	rrsets []desec.RRset, costWeight float32, ttl time.Duration,
) error {
	cacheableRRsets := filterRRsets(rrsets)
	for _, recordType := range RecordTypes {
		rrset, hasRRset := cacheableRRsets[recordType]
		if !hasRRset {
			c.SetNonexistentRRsetByNameAndType(domainName, subname, recordType, costWeight, ttl)
			continue
		}

		err := c.SetRRsetByNameAndType(domainName, subname, recordType, rrset, costWeight, ttl)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf(
				"couldn't set cache entry for the %s RRset for %s.%s",
				recordType, subname, domainName,
			))
		}
	}
	return nil
}

func (c *Cache) GetRRsetsByName(domainName, subname string) ([]desec.RRset, error) {
	rrsets := make([]desec.RRset, 0, len(RecordTypes))
	for _, recordType := range RecordTypes {
		key := keyRRsetByNameAndType(domainName, subname, recordType)
		var value desec.RRset
		keyExists, valueExists, err := c.getEntry(key, &value)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf(
				"couldn't get cache entry for the %s RRset for %s.%s",
				recordType, subname, domainName,
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
	return fmt.Sprintf("/dns/domains/[%s]/rrsets/[%s]/[%s]", domainName, subname, rrsetType)
}

func (c *Cache) SetRRsetByNameAndType(
	domainName, subname, rrsetType string,
	rrset desec.RRset, costWeight float32, ttl time.Duration,
) error {
	key := keyRRsetByNameAndType(domainName, subname, rrsetType)
	return c.setEntry(key, rrset, costWeight, ttl)
}

func (c *Cache) SetNonexistentRRsetByNameAndType(
	domainName, subname, rrsetType string, costWeight float32, ttl time.Duration,
) {
	key := keyRRsetByNameAndType(domainName, subname, rrsetType)
	c.setNonexistentEntry(key, costWeight, ttl)
}

func (c *Cache) GetRRsetByNameAndType(
	domainName, subname, rrsetType string,
) (*desec.RRset, bool, error) {
	key := keyRRsetByNameAndType(domainName, subname, rrsetType)
	var value desec.RRset
	keyExists, valueExists, err := c.getEntry(key, &value)
	if !keyExists || !valueExists || err != nil {
		return nil, keyExists, err
	}

	return &value, true, nil
}
