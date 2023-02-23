package desec

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/desec"
)

// Filtering & Sorting

func filterRRsets(rrsets []desec.RRset, recordTypes []string) map[string]desec.RRset {
	all := make(map[string]desec.RRset)
	for _, rrset := range rrsets {
		key := strings.ToUpper(rrset.Type)
		all[key] = rrset
	}
	filtered := make(map[string]desec.RRset)
	for _, recordType := range recordTypes {
		rrset, hasRRset := all[recordType]
		if hasRRset {
			filtered[recordType] = rrset
		}
	}
	return filtered
}

func FilterAndSortRRsets(rrsets []desec.RRset, recordTypes []string) []desec.RRset {
	filtered := filterRRsets(rrsets, recordTypes)
	sorted := make([]desec.RRset, 0, len(filtered))
	for _, recordType := range recordTypes {
		rrset, hasRRset := filtered[recordType]
		if hasRRset {
			sorted = append(sorted, rrset)
		}
	}
	return sorted
}

func GetReverseDomainNameFragments(domainName string) []string {
	fragments := strings.Split(domainName, ".")
	for i, j := 0, len(fragments)-1; i < j; i, j = i+1, j-1 {
		fragments[i], fragments[j] = fragments[j], fragments[i]
	}
	return fragments
}

func CompareSubnames(first, second string) bool {
	a := GetReverseDomainNameFragments(first)
	b := GetReverseDomainNameFragments(second)
	k := 0
	for k = 0; k < len(a) && k < len(b); k++ {
		if a[k] < b[k] {
			return true
		}

		if a[k] > b[k] {
			return false
		}
	}
	return len(a) < len(b)
}

func SortSubnameRRsets(
	rrsets map[string][]desec.RRset, filterRecordTypes []string,
) (subnames []string, sorted [][]desec.RRset) {
	subnames = make([]string, 0, len(rrsets))
	for subname := range rrsets {
		subnames = append(subnames, subname)
	}
	sort.Slice(subnames, func(i, j int) bool {
		return CompareSubnames(subnames[i], subnames[j])
	})
	sorted = make([][]desec.RRset, 0, len(subnames))
	for _, subname := range subnames {
		sorted = append(sorted, FilterAndSortRRsets(rrsets[subname], filterRecordTypes))
	}
	return subnames, sorted
}

// All RRsets

func (c *Client) getRRsetsFromCache() map[string][]desec.RRset {
	domainName := c.Config.DomainName
	subnames, err := c.Cache.GetSubnames(domainName)
	if err != nil && !errors.Is(err, context.Canceled) {
		// Log the error but return as a cache miss so we can manually query the RRsets
		c.Logger.Error(errors.Wrapf(
			err, "couldn't get the cache entry for the RRsets for %s", domainName,
		))
		return nil // treat an unparseable cache entry like a cache miss
	}

	if subnames == nil {
		return nil // this is the standard cache miss
	}

	rrsets := make(map[string][]desec.RRset)
	for _, subname := range subnames {
		subnameRRsets := c.getSubnameRRsetsFromCache(subname)
		if subnameRRsets == nil {
			return nil // cache miss for any subname is cache miss for the overall query
		}
		if len(subnameRRsets) > 0 {
			rrsets[subname] = subnameRRsets
		}
	}
	return rrsets
}

func (c *Client) getRRsetsFromDesec(ctx context.Context) (map[string][]desec.RRset, error) {
	client, cerr := c.Config.DNSServer.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	domainName := c.Config.DomainName
	params := desec.ListRRsetsParams{}
	res, err := client.ListRRsetsWithResponse(ctx, domainName, &params)
	if err != nil {
		return nil, err
	}
	// TODO: handle pagination
	if err = c.handleDesecClientError(*res.HTTPResponse, res.Body, c.Logger); err != nil {
		return nil, err
	}

	mergedRRsets := *res.JSON200
	rrsets := make(map[string][]desec.RRset)
	for _, rrset := range mergedRRsets {
		subname := rrset.Subname
		rrsets[subname] = append(rrsets[subname], rrset)
	}

	// Cache the results
	subnames := make([]string, 0, len(rrsets))
	for subname := range rrsets {
		subnames = append(subnames, subname)
	}
	if err = c.Cache.SetSubnames(domainName, subnames); err != nil {
		return nil, err
	}
	for subname, subnameRRsets := range rrsets {
		if err = c.Cache.SetRRsetsByName(domainName, subname, subnameRRsets); err != nil {
			return nil, err
		}
	}
	return rrsets, nil
}

func (c *Client) GetRRsets(ctx context.Context) (map[string][]desec.RRset, error) {
	if rrsets := c.getRRsetsFromCache(); rrsets != nil {
		return rrsets, nil
	}

	// We had a cache miss, so we need to query the desec API
	if err := c.tryAddLimitedRead(); err != nil {
		return nil, err
	}

	// fmt.Println("Performing a desec API read operation for GetRRsets...")
	return c.getRRsetsFromDesec(ctx)
}

func (c *Client) UpsertRRsets(ctx context.Context, rrsets ...desec.RRset) ([]desec.RRset, error) {
	if err := c.tryAddLimitedWrite(); err != nil {
		return nil, err
	}
	client, cerr := c.Config.DNSServer.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	// TODO: handle rate-limiting
	domainName := c.Config.DomainName
	res, err := client.PartialUpdateRRsetsWithResponse(ctx, domainName, rrsets)
	if err != nil {
		return nil, err
	}
	if err = c.handleDesecMissingDomainError(*res.HTTPResponse); err != nil {
		return nil, err
	}
	if err = c.handleDesecClientError(*res.HTTPResponse, res.Body, c.Logger); err != nil {
		return nil, err
	}

	returnedRRsets := *res.JSON200
	returnedKeys := make(map[RRsetKey]struct{})
	staleDomainNameCaches := make(map[string]struct{})
	for _, rrset := range returnedRRsets {
		key := NewRRsetKey(rrset)
		returnedKeys[key] = struct{}{}
		if err = c.Cache.SetRRsetByNameAndType(domainName, key.Subname, key.Type, rrset); err != nil {
			return nil, err
		}
		if !c.Cache.HasSubname(domainName, key.Subname) {
			staleDomainNameCaches[domainName] = struct{}{}
		}
	}
	for _, rrset := range rrsets {
		if !IsDeletionUpsertRRset(rrset) {
			continue
		}
		key := NewRRsetKey(rrset)
		if _, returned := returnedKeys[key]; !returned {
			c.Cache.SetNonexistentRRsetByNameAndType(domainName, key.Subname, key.Type)
		}
		if c.Cache.HasSubname(domainName, key.Subname) {
			staleDomainNameCaches[domainName] = struct{}{}
		}
	}

	for domainName := range staleDomainNameCaches {
		c.Cache.UnsetSubnames(domainName)
	}

	return returnedRRsets, nil
}

func (c *Client) DeleteRRsets(ctx context.Context, keys ...RRsetKey) error {
	rrsets := make([]desec.RRset, len(keys))
	for i, key := range keys {
		rrsets[i] = key.AsDeletionUpsertRRset()
	}
	returnedRRsets, err := c.UpsertRRsets(ctx, rrsets...)
	if len(returnedRRsets) != 0 {
		return errors.New("expected zero rrsets to be returned after a bulk delete operation")
	}
	return err
}

// Subname RRsets

func (c *Client) getSubnameRRsetsFromCache(subname string) []desec.RRset {
	domainName := c.Config.DomainName
	rrsets, err := c.Cache.GetRRsetsByName(domainName, subname)
	if err != nil && !errors.Is(err, context.Canceled) {
		// Log the error but return as a cache miss so we can manually query the RRsets
		c.Logger.Error(errors.Wrapf(
			err, "couldn't get the cache entry for one of the RRsets for %s.%s", subname, domainName,
		))
		return nil // treat an unparseable cache entry like a cache miss
	}

	return rrsets // rrsets may be nil, indicating a cache miss
}

func (c *Client) getSubnameRRsetsFromDesec(
	ctx context.Context, subname string,
) ([]desec.RRset, error) {
	client, cerr := c.Config.DNSServer.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	domainName := c.Config.DomainName
	params := desec.ListRRsetsParams{Subname: &subname}
	res, err := client.ListRRsetsWithResponse(ctx, domainName, &params)
	if err != nil {
		return nil, err
	}

	// TODO: handle pagination
	if err = c.handleDesecClientError(*res.HTTPResponse, res.Body, c.Logger); err != nil {
		return nil, err
	}

	rrsets := FilterAndSortRRsets(*res.JSON200, c.Cache.RecordTypes)
	if err = c.Cache.SetRRsetsByName(domainName, subname, rrsets); err != nil {
		return nil, err
	}

	return rrsets, nil
}

func (c *Client) GetSubnameRRsets(ctx context.Context, subname string) ([]desec.RRset, error) {
	if rrsets := c.getSubnameRRsetsFromCache(subname); rrsets != nil {
		return rrsets, nil
	}

	// We had a cache miss, so we need to query the desec API
	if err := c.tryAddLimitedRead(); err != nil {
		return nil, err
	}

	// fmt.Println("Performing a desec API read operation for GetSubnameRRsets...")
	return c.getSubnameRRsetsFromDesec(ctx, subname)
}

// Individual RRset

func (c *Client) getRRsetFromCache(subname, recordType string) (*desec.RRset, bool) {
	domainName := c.Config.DomainName
	rrset, cacheHit, err := c.Cache.GetRRsetByNameAndType(domainName, subname, recordType)
	if err != nil && !errors.Is(err, context.Canceled) {
		// Log the error but return as a cache miss so we can manually query the RRsets
		c.Logger.Error(errors.Wrapf(
			err, "couldn't get the cache entry for the %s RRsets for %s.%s",
			recordType, subname, domainName,
		))
		return nil, false // treat an unparseable cache entry like a cache miss
	}

	return rrset, cacheHit // cache hit with nil rrset indicates nonexistent RRset
}

func (c *Client) getRRsetFromDesec(
	ctx context.Context, subname, recordType string,
) (*desec.RRset, error) {
	client, cerr := c.Config.DNSServer.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	domainName := c.Config.DomainName
	res, err := client.RetrieveRRsetWithResponse(ctx, domainName, subname, recordType)
	if err != nil {
		return nil, err
	}

	if err = c.handleDesecMissingRRsetError(*res.HTTPResponse, subname, recordType); err != nil {
		c.Cache.SetNonexistentRRsetByNameAndType(domainName, subname, recordType)
		return nil, nil // treat this as a nonexistent RRset
	}

	if err = c.handleDesecClientError(*res.HTTPResponse, res.Body, c.Logger); err != nil {
		return nil, err
	}

	rrset := res.JSON200
	if err = c.Cache.SetRRsetByNameAndType(domainName, subname, recordType, *rrset); err != nil {
		return nil, err
	}

	return rrset, nil
}

func (c *Client) GetRRset(ctx context.Context, subname, recordType string) (*desec.RRset, error) {
	if rrset, cacheHit := c.getRRsetFromCache(subname, recordType); cacheHit {
		return rrset, nil // nil rrset indicates nonexistent RRset
	}

	// We had a cache miss, so we need to query the desec API
	if err := c.tryAddLimitedRead(); err != nil {
		return nil, err
	}

	// fmt.Println("Performing a desec API read operation for GetRRset...")
	return c.getRRsetFromDesec(ctx, subname, recordType)
}

func (c *Client) CreateRRset(
	ctx context.Context, subname, recordType string, ttl int64, records []string,
) (desec.RRset, error) {
	if err := c.tryAddLimitedWrite(); err != nil {
		return desec.RRset{}, err
	}
	client, cerr := c.Config.DNSServer.NewClient()
	if cerr != nil {
		return desec.RRset{}, cerr
	}

	// TODO: handle rate-limiting
	domainName := c.Config.DomainName
	intTTL := int(ttl)
	requestBody := desec.RRset{
		Subname: subname,
		Type:    recordType,
		Ttl:     &intTTL,
		Records: records,
	}
	res, err := client.CreateRRsetsWithResponse(ctx, domainName, []desec.RRset{requestBody})
	if err != nil {
		return desec.RRset{}, err
	}

	if err = c.handleDesecMissingDomainError(*res.HTTPResponse); err != nil {
		return desec.RRset{}, err
	}

	if err = c.handleDesecClientError(*res.HTTPResponse, res.Body, c.Logger); err != nil {
		return desec.RRset{}, err
	}

	rrsets := res.JSON201
	if len(*rrsets) != 1 {
		return desec.RRset{}, errors.Errorf(
			"response for creating a single rrset unexpectedly contains %d rrsets", len(*rrsets),
		)
	}
	rrset := (*rrsets)[0]
	if err = c.Cache.SetRRsetByNameAndType(
		domainName, subname, rrset.Type, rrset,
	); err != nil {
		return desec.RRset{}, err
	}

	if !c.Cache.HasSubname(domainName, subname) {
		c.Cache.UnsetSubnames(domainName)
	}

	return rrset, nil
}

func (c *Client) UpdateRRset(
	ctx context.Context, subname, recordType string, ttl int64, records []string,
) (*desec.RRset, error) {
	if err := c.tryAddLimitedWrite(); err != nil {
		return nil, err
	}
	client, cerr := c.Config.DNSServer.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	// TODO: handle rate-limiting
	domainName := c.Config.DomainName
	intTTL := int(ttl)
	requestBody := desec.RRset{
		Subname: subname,
		Type:    recordType,
		Ttl:     &intTTL,
		Records: records,
	}
	res, err := client.UpdateRRsetWithResponse(ctx, domainName, subname, recordType, requestBody)
	if err != nil {
		return nil, err
	}
	if err = c.handleDesecMissingDomainError(*res.HTTPResponse); err != nil {
		return nil, err
	}
	if err = c.handleDesecClientError(*res.HTTPResponse, res.Body, c.Logger); err != nil {
		return nil, err
	}

	if res.StatusCode() == http.StatusNoContent {
		c.Cache.SetNonexistentRRsetByNameAndType(domainName, subname, recordType)
		return nil, nil
	}

	rrset := res.JSON200
	if err = c.Cache.SetRRsetByNameAndType(
		domainName, subname, rrset.Type, *rrset,
	); err != nil {
		return nil, err
	}

	if !c.Cache.HasSubname(domainName, subname) {
		c.Cache.UnsetSubnames(domainName)
	}

	return rrset, nil
}

func (c *Client) DeleteRRset(ctx context.Context, subname, recordType string) error {
	if err := c.tryAddLimitedWrite(); err != nil {
		return err
	}
	client, cerr := c.Config.DNSServer.NewClient()
	if cerr != nil {
		return cerr
	}

	// TODO: handle rate-limiting
	domainName := c.Config.DomainName
	res, err := client.DestroyRRsetWithResponse(ctx, domainName, subname, recordType)
	if err != nil {
		return err
	}

	if err = c.handleDesecMissingDomainError(*res.HTTPResponse); err != nil {
		return err
	}
	if err = c.handleDesecClientError(*res.HTTPResponse, res.Body, c.Logger); err != nil {
		return err
	}

	c.Cache.SetNonexistentRRsetByNameAndType(
		domainName, subname, recordType,
	)
	if c.Cache.HasSubname(domainName, subname) {
		c.Cache.UnsetSubnames(domainName)
	}
	return nil
}
