package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/framework/log"
)

// Domain

func getDomainFromCache(
	domainName string, cache *Cache, l log.Logger,
) (*desec.Domain, error) {
	domain, cacheHit, err := cache.GetDomainByName(domainName)
	if err != nil {
		// Log the error and proceed to manually query all controllers
		l.Error(errors.Wrap(err, fmt.Sprintf(
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

func getDomainFromDesec(
	ctx context.Context, domain *DNSDomain, l log.Logger,
) (*desec.Domain, error) {
	client, cerr := domain.makeClientWithResponses()
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.RetrieveDomainWithResponse(ctx, domain.DomainName)
	if err != nil {
		return nil, err
	}

	if err = domain.handleDesecClientError(*res.HTTPResponse, l); err != nil {
		return nil, err
	}

	desecDomain := res.JSON200
	if err = domain.Cache.SetDomainByName(
		domain.DomainName, *desecDomain,
		domain.Server.NetworkCostWeight, domain.APISettings.ReadCacheTTL,
	); err != nil {
		return nil, err
	}

	return desecDomain, nil
}

func GetDomain(
	ctx context.Context, domain *DNSDomain, l log.Logger,
) (*desec.Domain, error) {
	desecDomain, err := getDomainFromCache(domain.DomainName, domain.Cache, l)
	if err != nil {
		return nil, err
	}

	if desecDomain != nil {
		return desecDomain, nil
	}

	// We had a cache miss, so we need to query the desec API
	if err = domain.tryAddLimitedRead(); err != nil {
		return nil, err
	}

	return getDomainFromDesec(ctx, domain, l)
}
