package desec

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/desec"
)

// Domain

func (c *Client) getDomainFromCache() (*desec.Domain, bool) {
	domainName := c.Config.DomainName
	domain, cacheHit, err := c.Cache.GetDomainByName(domainName)
	if err != nil {
		// Log the error but return as a cache miss so we can manually query the domain
		c.Logger.Error(errors.Wrap(err, fmt.Sprintf(
			"couldn't get the cache entry for the domain with name %s", domainName,
		)))
		return nil, false // treat an unparseable cache entry like a cache miss
	}
	return domain, cacheHit // cache hit with nil domain indicates nonexistent domain
}

func (c *Client) getDomainFromDesec(ctx context.Context) (*desec.Domain, error) {
	client, cerr := c.Config.DNSServer.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	domainName := c.Config.DomainName
	res, err := client.RetrieveDomainWithResponse(ctx, domainName)
	if err != nil {
		return nil, err
	}

	if err = c.handleDesecClientError(*res.HTTPResponse, c.Logger); err != nil {
		return nil, err
	}

	domain := res.JSON200
	if err = c.Cache.SetDomainByName(domainName, *domain); err != nil {
		return nil, err
	}

	return domain, nil
}

func (c *Client) GetDomain(ctx context.Context) (*desec.Domain, error) {
	if domain, cacheHit := c.getDomainFromCache(); cacheHit {
		return domain, nil // nil domain indicates nonexistent domain
	}
	if err := c.tryAddLimitedRead(); err != nil {
		return nil, err
	}
	return c.getDomainFromDesec(ctx)
}
