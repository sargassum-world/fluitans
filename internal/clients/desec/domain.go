package desec

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/desec"
)

// Domain

func (c *Client) getDomainFromCache() (*desec.Domain, error) {
	domainName := c.Config.DomainName
	domain, cacheHit, err := c.Cache.GetDomainByName(domainName)
	if err != nil {
		// Log the error and proceed to manually query all controllers
		c.Logger.Error(errors.Wrap(err, fmt.Sprintf(
			"couldn't get the cache entry for the domain with name %s", domainName,
		)))
		return nil, nil // treat an unparseable cache entry like a cache miss
	}

	if !cacheHit {
		return nil, nil
	}

	if domain == nil {
		// The cache has a record (within TTL) that the domain does not exist, so
		// there's no need to issue a request to the deSEC API only for it to tell
		// us that the domain still does not exist. So we treat it like a cache hit.
		return nil, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("couldn't find domain %s", domainName),
		)
	}

	return domain, nil
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
	domain, err := c.getDomainFromCache()
	if err != nil {
		return nil, err
	}

	if domain != nil {
		return domain, nil
	}

	// We had a cache miss, so we need to query the desec API
	if err = c.tryAddLimitedRead(); err != nil {
		return nil, err
	}

	return c.getDomainFromDesec(ctx)
}
