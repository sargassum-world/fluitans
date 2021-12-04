package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/framework/log"
)

func filterRRsets(rrsets []desec.RRset) map[string]desec.RRset {
	all := make(map[string]desec.RRset)
	for _, rrset := range rrsets {
		key := strings.ToUpper(rrset.Type)
		all[key] = rrset
	}
	filtered := make(map[string]desec.RRset)
	for _, recordType := range RecordTypes {
		rrset, hasRRset := all[recordType]
		if hasRRset {
			filtered[recordType] = rrset
		}
	}
	return filtered
}

func FilterAndSortRRsets(rrsets []desec.RRset) []desec.RRset {
	filtered := filterRRsets(rrsets)
	sorted := make([]desec.RRset, 0, len(filtered))
	for _, recordType := range RecordTypes {
		rrset, hasRRset := filtered[recordType]
		if hasRRset {
			sorted = append(sorted, rrset)
		}
	}
	return sorted
}

// All RRsets

func getRRsetsFromCache(
	domainName string, cache *Cache, l log.Logger,
) map[string][]desec.RRset {
	subnames, err := cache.GetSubnames(domainName)
	if err != nil {
		// Log the error but return as a cache miss so we can manually query the RRsets
		l.Error(errors.Wrap(err, fmt.Sprintf(
			"couldn't get the cache entry for the RRsets for %s", domainName,
		)))
		return nil // treat an unparseable cache entry like a cache miss
	}

	if subnames == nil {
		return nil // this is the standard cache miss
	}

	rrsets := make(map[string][]desec.RRset)
	for _, subname := range subnames {
		subnameRRsets := getSubnameRRsetsFromCache(domainName, subname, cache, l)
		if subnameRRsets == nil {
			return nil // cache miss for any subname is cache miss for the overall query
		}
		if len(subnameRRsets) > 0 {
			rrsets[subname] = subnameRRsets
		}
	}
	return rrsets
}

func getRRsetsFromDesec(
	ctx context.Context, domain *DNSDomain, l log.Logger,
) (map[string][]desec.RRset, error) {
	client, cerr := domain.makeClientWithResponses()
	if cerr != nil {
		return nil, cerr
	}

	params := desec.ListRRsetsParams{}
	res, err := client.ListRRsetsWithResponse(ctx, domain.DomainName, &params)
	if err != nil {
		return nil, err
	}

	// TODO: handle pagination
	if err = domain.handleDesecClientError(*res.HTTPResponse, l); err != nil {
		return nil, err
	}

	mergedRRsets := *res.JSON200
	rrsets := make(map[string][]desec.RRset)
	for _, rrset := range mergedRRsets {
		subname := *rrset.Subname
		rrsets[subname] = append(rrsets[subname], rrset)
	}
	subnames := make([]string, 0, len(rrsets))
	for subname := range rrsets {
		subnames = append(subnames, subname)
	}
	if err = domain.Cache.SetSubnames(
		domain.DomainName, subnames,
		domain.Server.NetworkCostWeight, domain.APISettings.ReadCacheTTL,
	); err != nil {
		return nil, err
	}

	for subname, subnameRRsets := range rrsets {
		if err = domain.Cache.SetRRsetsByName(
			domain.DomainName, subname, subnameRRsets,
			domain.Server.NetworkCostWeight, domain.APISettings.ReadCacheTTL,
		); err != nil {
			return nil, err
		}
	}
	return rrsets, nil
}

func GetRRsets(
	ctx context.Context, domain *DNSDomain, l log.Logger,
) (map[string][]desec.RRset, error) {
	if rrsets := getRRsetsFromCache(domain.DomainName, domain.Cache, l); rrsets != nil {
		return rrsets, nil
	}

	// We had a cache miss, so we need to query the desec API
	if err := domain.tryAddLimitedRead(); err != nil {
		return nil, err
	}

	// fmt.Println("Performing a desec API read operation for GetRRsets...")
	return getRRsetsFromDesec(ctx, domain, l)
}

// Subname RRsets

func getSubnameRRsetsFromCache(
	domainName, subname string, cache *Cache, l log.Logger,
) []desec.RRset {
	rrsets, err := cache.GetRRsetsByName(domainName, subname)
	if err != nil {
		// Log the error but return as a cache miss so we can manually query the RRsets
		l.Error(errors.Wrap(err, fmt.Sprintf(
			"couldn't get the cache entry for one of the RRsets for %s.%s",
			subname, domainName,
		)))
		return nil // treat an unparseable cache entry like a cache miss
	}

	return rrsets // rrsets may be nil, indicating a cache miss
}

func getSubnameRRsetsFromDesec(
	ctx context.Context, domain *DNSDomain, subname string, l log.Logger,
) ([]desec.RRset, error) {
	client, cerr := domain.makeClientWithResponses()
	if cerr != nil {
		return nil, cerr
	}

	params := desec.ListRRsetsParams{Subname: &subname}
	res, err := client.ListRRsetsWithResponse(ctx, domain.DomainName, &params)
	if err != nil {
		return nil, err
	}

	// TODO: handle pagination
	if err = domain.handleDesecClientError(*res.HTTPResponse, l); err != nil {
		return nil, err
	}

	rrsets := FilterAndSortRRsets(*res.JSON200)
	if err = domain.Cache.SetRRsetsByName(
		domain.DomainName, subname, rrsets,
		domain.Server.NetworkCostWeight, domain.APISettings.ReadCacheTTL,
	); err != nil {
		return nil, err
	}

	return rrsets, nil
}

func GetSubnameRRsets(
	ctx context.Context, domain *DNSDomain, subname string, l log.Logger,
) ([]desec.RRset, error) {
	if rrsets := getSubnameRRsetsFromCache(
		domain.DomainName, subname, domain.Cache, l,
	); rrsets != nil {
		return rrsets, nil
	}

	// We had a cache miss, so we need to query the desec API
	if err := domain.tryAddLimitedRead(); err != nil {
		return nil, err
	}

	// fmt.Println("Performing a desec API read operation for GetSubnameRRsets...")
	return getSubnameRRsetsFromDesec(ctx, domain, subname, l)
}

// Individual RRset

func getRRsetFromCache(
	domainName, subname, recordType string, cache *Cache, l log.Logger,
) (*desec.RRset, bool) {
	rrset, cacheHit, err := cache.GetRRsetByNameAndType(domainName, subname, recordType)
	if err != nil {
		// Log the error but return as a cache miss so we can manually query the RRsets
		l.Error(errors.Wrap(err, fmt.Sprintf(
			"couldn't get the cache entry for the %s RRsets for %s.%s",
			recordType, subname, domainName,
		)))
		return nil, false // treat an unparseable cache entry like a cache miss
	}

	return rrset, cacheHit // cache hit with nil rrset indicates nonexistent RRset
}

func getRRsetFromDesec(
	ctx context.Context, domain *DNSDomain, subname, recordType string, l log.Logger,
) (*desec.RRset, error) {
	client, cerr := domain.makeClientWithResponses()
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.RetrieveRRsetWithResponse(
		ctx, domain.DomainName, subname, recordType,
	)
	if err != nil {
		return nil, err
	}

	if err = domain.handleDesecMissingRRsetError(
		*res.HTTPResponse, subname, recordType,
	); err != nil {
		domain.Cache.SetNonexistentRRsetByNameAndType(
			domain.DomainName, subname, recordType,
			domain.Server.NetworkCostWeight, domain.APISettings.ReadCacheTTL,
		)
		return nil, nil // treat this as a nonexistent RRset
	}

	if err = domain.handleDesecClientError(*res.HTTPResponse, l); err != nil {
		return nil, err
	}

	rrset := res.JSON200
	if err = domain.Cache.SetRRsetByNameAndType(
		domain.DomainName, subname, recordType, *rrset,
		domain.Server.NetworkCostWeight, domain.APISettings.ReadCacheTTL,
	); err != nil {
		return nil, err
	}

	return rrset, nil
}

func GetRRset(
	ctx context.Context, domain *DNSDomain, subname, recordType string, l log.Logger,
) (*desec.RRset, error) {
	rrset, cacheHit := getRRsetFromCache(
		domain.DomainName, subname, recordType, domain.Cache, l,
	)
	if cacheHit {
		return rrset, nil // nil rrset indicates nonexistent RRset
	}

	// We had a cache miss, so we need to query the desec API
	if err := domain.tryAddLimitedRead(); err != nil {
		return nil, err
	}

	// fmt.Println("Performing a desec API read operation for GetRRset...")
	return getRRsetFromDesec(ctx, domain, subname, recordType, l)
}

func CreateRRset(
	ctx context.Context, domain *DNSDomain,
	subname string, recordType string, ttl int64, records []string,
	l log.Logger,
) (*desec.RRset, error) {
	client, cerr := domain.makeClientWithResponses()
	if cerr != nil {
		return nil, cerr
	}

	// TODO: handle rate-limiting
	requestBody := desec.CreateRRsetsJSONRequestBody{
		Subname: &subname,
		Type:    recordType,
		Ttl:     int(ttl),
		Records: records,
	}
	res, err := client.CreateRRsetsWithResponse(ctx, domain.DomainName, requestBody)
	if err != nil {
		return nil, err
	}

	if err = domain.handleDesecMissingDomainError(*res.HTTPResponse); err != nil {
		return nil, err
	}

	if err = domain.handleDesecClientError(*res.HTTPResponse, l); err != nil {
		return nil, err
	}

	rrset := res.JSON201
	if err = domain.Cache.SetRRsetByNameAndType(
		domain.DomainName, subname, rrset.Type, *rrset,
		domain.Server.NetworkCostWeight, domain.APISettings.ReadCacheTTL,
	); err != nil {
		return nil, err
	}

	if !domain.Cache.HasSubname(domain.DomainName, subname) {
		domain.Cache.UnsetSubnames(domain.DomainName)
	}

	return rrset, nil
}

func DeleteRRset(
	ctx context.Context, domain *DNSDomain, subname string, recordType string,
	l log.Logger,
) error {
	client, cerr := domain.makeClientWithResponses()
	if cerr != nil {
		return cerr
	}

	// TODO: handle rate-limiting
	res, err := client.DestroyRRsetWithResponse(
		ctx, domain.DomainName, subname, recordType,
	)
	if err != nil {
		return err
	}

	if err = domain.handleDesecMissingDomainError(*res.HTTPResponse); err != nil {
		return err
	}

	if err = domain.handleDesecClientError(*res.HTTPResponse, l); err != nil {
		return err
	}

	domain.Cache.SetNonexistentRRsetByNameAndType(
		domain.DomainName, subname, recordType,
		domain.Server.NetworkCostWeight, domain.APISettings.ReadCacheTTL,
	)
	return nil
}
