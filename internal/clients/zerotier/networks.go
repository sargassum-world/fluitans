package zerotier

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

func GetControllerAddress(networkID string) string {
	addressLength := 10
	return networkID[:addressLength]
}

// All Networks

func (c *Client) getNetworkIDsFromCache(
	controller ztcontrollers.Controller, cc *ztcontrollers.Client,
) []string {
	networkIDs, err := cc.Cache.GetNetworkIDsByServer(controller.Server)
	if err != nil {
		// Log the error but return as a cache miss so we can manually query the network IDs
		c.Logger.Error(errors.Wrapf(
			err, "couldn't get the cache entry for the network IDs controlled by %s", controller.Name,
		))
		return nil // treat an unparseable cache entry like a cache miss
	}

	return networkIDs // networkIDs may be nil, which indicates a cache miss
}

func (c *Client) getNetworkIDsFromZerotier(
	ctx context.Context, controller ztcontrollers.Controller, cc *ztcontrollers.Client,
) ([]string, error) {
	client, cerr := controller.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.GetControllerNetworksWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the results
	networkIDs := *res.JSON200
	if err := cc.Cache.SetNetworkIDsByServer(
		controller.Server, networkIDs, controller.NetworkCostWeight,
	); err != nil {
		return nil, err
	}
	if len(networkIDs) > 0 {
		// It's safe to assume that all networks under a controller have the same controller address,
		// so we only need to cache the controller whose address is part of the first network's ID
		err := cc.Cache.SetControllerByAddress(GetControllerAddress(networkIDs[0]), controller)
		if err != nil {
			return nil, err
		}
	}

	return networkIDs, nil
}

func (c *Client) GetNetworkIDs(
	ctx context.Context, controller ztcontrollers.Controller, cc *ztcontrollers.Client,
) ([]string, error) {
	if networkIDs := c.getNetworkIDsFromCache(controller, cc); networkIDs != nil {
		return networkIDs, nil
	}
	return c.getNetworkIDsFromZerotier(ctx, controller, cc)
}

func (c *Client) GetAllNetworkIDs(
	ctx context.Context, controllers []ztcontrollers.Controller, cc *ztcontrollers.Client,
) ([][]string, error) {
	eg, ctx := errgroup.WithContext(ctx)
	allNetworkIDs := make([][]string, len(controllers))
	for i, controller := range controllers {
		eg.Go(func(i int, controller ztcontrollers.Controller) func() error {
			return func() (err error) {
				allNetworkIDs[i], err = c.GetNetworkIDs(ctx, controller, cc)
				return err
			}
		}(i, controller))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return allNetworkIDs, nil
}

func (c *Client) GetNetworks(
	ctx context.Context, controller ztcontrollers.Controller, ids []string,
) (map[string]zerotier.ControllerNetwork, error) {
	eg, ctx := errgroup.WithContext(ctx)
	networks := make([]*zerotier.ControllerNetwork, len(ids))
	for i, id := range ids {
		eg.Go(func(i int, id string) func() error {
			return func() (err error) {
				networks[i], err = c.GetNetwork(ctx, controller, id)
				return err
			}
		}(i, id))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Transform the responses into a more usable shape
	keyedNetworks := make(map[string]zerotier.ControllerNetwork, len(ids))
	for i, id := range ids {
		if networks[i] != nil {
			keyedNetworks[id] = *networks[i]
		}
	}
	return keyedNetworks, nil
}

func (c *Client) GetAllNetworks(
	ctx context.Context, controllers []ztcontrollers.Controller, ids [][]string,
) ([]map[string]zerotier.ControllerNetwork, error) {
	if len(controllers) != len(ids) {
		return nil, errors.Errorf("lists of controllers and ids must have the same length")
	}

	eg, ctx := errgroup.WithContext(ctx)
	allNetworks := make([]map[string]zerotier.ControllerNetwork, len(controllers))
	for i, controller := range controllers {
		eg.Go(func(i int, controller ztcontrollers.Controller, someIDs []string) func() error {
			return func() (err error) {
				allNetworks[i], err = c.GetNetworks(ctx, controller, someIDs)
				return err
			}
		}(i, controller, ids[i]))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return allNetworks, nil
}

// Individual Network

func (c *Client) getNetworkFromCache(id string) (*zerotier.ControllerNetwork, bool) {
	network, cacheHit, err := c.Cache.GetNetworkByID(id)
	if err != nil {
		// Log the error but return as a cache miss so we can manually query the network
		c.Logger.Error(errors.Wrapf(err, "couldn't get the cache entry for the network with id %s", id))
		return nil, false // treat an unparseable cache entry like a cache miss
	}
	return network, cacheHit // cache hit with nil rrset indicates nonexistent RRset
}

func (c *Client) getNetworkFromZerotier(
	ctx context.Context, controller ztcontrollers.Controller, id string,
) (*zerotier.ControllerNetwork, error) {
	client, cerr := controller.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.GetControllerNetworkWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if res.HTTPResponse.StatusCode == http.StatusNotFound {
		c.Cache.SetNonexistentNetworkByID(id)
		return nil, nil
	}

	if err := c.Cache.SetNetworkByID(id, *res.JSON200); err != nil {
		return nil, err
	}

	return res.JSON200, nil
}

func (c *Client) GetNetwork(
	ctx context.Context, controller ztcontrollers.Controller, id string,
) (*zerotier.ControllerNetwork, error) {
	if network, cacheHit := c.getNetworkFromCache(id); cacheHit {
		return network, nil // nil network indicates nonexistent network
	}
	return c.getNetworkFromZerotier(ctx, controller, id)
}

func (c *Client) getNetworkMemberAddressesFromCache(networkID string) []string {
	addresses, err := c.Cache.GetNetworkMembersByID(networkID)
	if err != nil {
		// Log the error but return as a cache miss so we can manually query the member addresses
		c.Logger.Error(errors.Wrapf(
			err, "couldn't get the cache entry for the member addresses of network %s", networkID,
		))
		return nil // treat an unparseable cache entry like a cache miss
	}

	return addresses // addresses may be nil, which indicates a cache miss
}

func (c *Client) getNetworkMemberAddressesFromZerotier(
	ctx context.Context, controller ztcontrollers.Controller, id string,
) ([]string, error) {
	client, cerr := controller.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.GetControllerNetworkMembersWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}

	// Transform the response into a usable shape
	var memberRevisions map[string]int
	if err = json.Unmarshal(res.Body, &memberRevisions); err != nil {
		return nil, err
	}
	memberAddresses := make([]string, 0, len(memberRevisions))
	for address := range memberRevisions {
		memberAddresses = append(memberAddresses, address)
	}

	if err := c.Cache.SetNetworkMembersByID(id, memberAddresses); err != nil {
		return nil, err
	}

	return memberAddresses, nil
}

func (c *Client) GetNetworkMemberAddresses(
	ctx context.Context, controller ztcontrollers.Controller, id string,
) ([]string, error) {
	if addresses := c.getNetworkMemberAddressesFromCache(id); addresses != nil {
		return addresses, nil
	}
	return c.getNetworkMemberAddressesFromZerotier(ctx, controller, id)
}

func (c *Client) GetNetworkInfo(
	ctx context.Context, controller ztcontrollers.Controller, id string,
) (*zerotier.ControllerNetwork, []string, error) {
	eg, ctx := errgroup.WithContext(ctx)
	var network *zerotier.ControllerNetwork
	var addresses []string
	eg.Go(func() (err error) {
		network, err = c.GetNetwork(ctx, controller, id)
		return err
	})
	eg.Go(func() (err error) {
		addresses, err = c.GetNetworkMemberAddresses(ctx, controller, id)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	return network, addresses, nil
}

func makeDefaultRules() []map[string]interface{} {
	ipv4Type := 2048
	ipv4ARPType := 2054
	ipv6Type := 34525
	return []map[string]interface{}{
		{
			"type":      "MATCH_ETHERTYPE",
			"etherType": ipv4Type,
			"not":       true,
		},
		{
			"type":      "MATCH_ETHERTYPE",
			"etherType": ipv4ARPType,
			"not":       true,
		},
		{
			"type":      "MATCH_ETHERTYPE",
			"etherType": ipv6Type,
			"not":       true,
		},
		{
			"type": "ACTION_DROP",
		},
		{
			"type": "ACTION_ACCEPT",
		},
	}
}

func makeDefaultNetworkRequest() zerotier.GenerateControllerNetworkJSONRequestBody {
	private := true
	n6plane := true
	v6AssignMode := zerotier.V6AssignMode{
		N6plane: &n6plane,
		Rfc4193: nil,
		Zt:      nil,
	}
	rules := makeDefaultRules()

	body := zerotier.GenerateControllerNetworkJSONRequestBody{}
	body.Private = &private
	body.V6AssignMode = &v6AssignMode
	body.Rules = &rules
	return body
}

func (c *Client) CreateNetwork(
	ctx context.Context, controller ztcontrollers.Controller, cc *ztcontrollers.Client,
) (*zerotier.ControllerNetwork, error) {
	client, cerr := controller.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	address, err := cc.GetAddress(ctx, controller)
	if err != nil || address == "" {
		return nil, err
	}

	body := makeDefaultNetworkRequest()
	nRes, err := client.GenerateControllerNetworkWithResponse(
		ctx, address, body,
	)
	if err != nil {
		return nil, err
	}

	// TODO: this should only happen on a success HTTP status code
	cc.Cache.UnsetNetworkIDsByServer(controller.Server)
	if err = c.Cache.SetNetworkByID(*nRes.JSON200.Id, *nRes.JSON200); err != nil {
		return nil, err
	}
	return nRes.JSON200, nil
}

func (c *Client) UpdateNetwork(
	ctx context.Context, controller ztcontrollers.Controller, id string,
	network zerotier.SetControllerNetworkJSONRequestBody,
) error {
	client, err := controller.NewClient()
	if err != nil {
		return err
	}

	res, err := client.SetControllerNetworkWithResponse(ctx, id, network)
	if err != nil {
		return err
	}

	// TODO: this should only happen on a success HTTP status code
	return c.Cache.SetNetworkByID(id, *res.JSON200)
}

func (c *Client) DeleteNetwork(
	ctx context.Context, controller ztcontrollers.Controller, id string, cc *ztcontrollers.Client,
) error {
	client, err := controller.NewClient()
	if err != nil {
		return err
	}

	_, err = client.DeleteControllerNetworkWithResponse(ctx, id)
	if err != nil {
		return err
	}

	// TODO: this should only happen on a success HTTP status code
	cc.Cache.UnsetNetworkIDsByServer(controller.Server)
	c.Cache.SetNonexistentNetworkByID(id)
	return nil
}
