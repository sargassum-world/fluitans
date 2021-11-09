package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

// Controller

func checkCachedController(c echo.Context, address string, cache *Cache) (*Controller, error) {
	controller, cacheHit, err := cache.GetControllerByAddress(address)
	if err != nil {
		return nil, err
	}

	if !cacheHit {
		return nil, nil
	}

	client, cerr := zerotier.NewAuthClientWithResponses(
		controller.Server, controller.Authtoken,
	)
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.GetStatusWithResponse(c.Request().Context())
	if err != nil {
		// evict the cached controller if its authtoken is stale or it no longer responds
		cache.UnsetControllerByAddress(address)
		return nil, err
	}

	if address != *res.JSON200.Address { // cached controller's address is stale
		c.Logger().Warnf(
			"controller %s's address has changed from %s to %s",
			controller.Server, address, *res.JSON200.Address,
		)
		err = cache.SetControllerByAddress(*res.JSON200.Address, *controller)
		if err != nil {
			return nil, err
		}

		cache.UnsetControllerByAddress(address)
		return nil, nil
	}

	return controller, nil
}

func scanControllers(
	c echo.Context, controllers []Controller, cache *Cache,
) ([]string, error) {
	eg, ctx := errgroup.WithContext(c.Request().Context())
	addresses := make([]string, len(controllers))
	for i, controller := range controllers {
		eg.Go(func(i int) func() error {
			return func() error {
				client, cerr := zerotier.NewAuthClientWithResponses(
					controller.Server, controller.Authtoken,
				)
				if cerr != nil {
					return nil
				}

				res, err := client.GetStatusWithResponse(ctx)
				if err != nil {
					return err
				}

				addresses[i] = *res.JSON200.Address
				return nil
			}
		}(i))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	for i, v := range controllers {
		err := cache.SetControllerByAddress(addresses[i], v)
		if err != nil {
			return nil, err
		}
	}

	return addresses, nil
}

func FindControllerByAddress(
	c echo.Context, address string, cache *Cache,
) (*Controller, error) {
	controller, err := checkCachedController(c, address, cache)
	if err != nil {
		// Log the error and proceed to manually query all controllers
		c.Logger().Error(err)
	} else if controller != nil {
		return controller, nil
	}

	// Query the list of all known controllers
	controllers, err := GetControllers()
	if err != nil {
		return nil, err
	}

	c.Logger().Warnf(
		"rescanning controllers due to a stale/missing controller for %s in cache",
		address,
	)
	addresses, err := scanControllers(c, controllers, cache)
	if err != nil {
		return nil, err
	}

	for i, v := range controllers {
		if addresses[i] == address {
			return &v, nil
		}
	}

	return nil, echo.NewHTTPError(
		http.StatusNotFound,
		fmt.Sprintf("Controller not found with address %s", address),
	)
}

func GetController(
	c echo.Context, controller Controller, cache *Cache,
) (*zerotier.Status, *zerotier.ControllerStatus, []string, error) {
	client, cerr := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if cerr != nil {
		return nil, nil, nil, cerr
	}

	var status *zerotier.Status
	var controllerStatus *zerotier.ControllerStatus
	var networks []string
	eg, ctx := errgroup.WithContext(c.Request().Context())
	eg.Go(func() error {
		res, err := client.GetStatusWithResponse(ctx)
		if err != nil {
			return err
		}

		status = res.JSON200
		if err := cache.SetControllerByAddress(*status.Address, controller); err != nil {
			return err
		}

		return nil
	})
	eg.Go(func() error {
		res, err := client.GetControllerStatusWithResponse(ctx)
		if err != nil {
			return err
		}

		controllerStatus = res.JSON200
		return err
	})
	eg.Go(func() error {
		res, err := client.GetControllerNetworksWithResponse(ctx)
		if err != nil {
			return err
		}

		networks = *res.JSON200
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, nil, err
	}

	return status, controllerStatus, networks, nil
}

// Networks

func GetControllerAddress(networkID string) string {
	addressLength := 10
	return networkID[:addressLength]
}

func GetNetworkIDs(
	c echo.Context, controllers []Controller, cache *Cache,
) ([][]string, error) {
	eg, ctx := errgroup.WithContext(c.Request().Context())
	networkIDs := make([][]string, len(controllers))
	for i := range controllers {
		networkIDs[i] = []string{}
	}
	for i, controller := range controllers {
		eg.Go(func(i int, controller Controller) func() error {
			return func() error {
				client, cerr := zerotier.NewAuthClientWithResponses(
					controller.Server, controller.Authtoken,
				)
				if cerr != nil {
					return nil
				}

				res, err := client.GetControllerNetworksWithResponse(ctx)
				if err != nil {
					return err
				}

				networkIDs[i] = *res.JSON200
				return nil
			}
		}(i, controller))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	for i, controller := range controllers {
		if len(networkIDs[i]) > 0 {
			// It's safe to assume that all networks under a controller have the same
			// controller address, so we only need to cache the controller named by the
			// first network
			err := cache.SetControllerByAddress(GetControllerAddress(networkIDs[i][0]), controller)
			if err != nil {
				return nil, err
			}
		}
	}
	return networkIDs, nil
}

func GetNetworks(
	c echo.Context, controllers []Controller, ids [][]string,
) ([]map[string]zerotier.ControllerNetwork, error) {
	eg, ctx := errgroup.WithContext(c.Request().Context())
	networks := make([][]zerotier.ControllerNetwork, len(controllers))
	for i := range controllers {
		networks[i] = make([]zerotier.ControllerNetwork, len(ids[i]))
		for j := range ids[i] {
			networks[i][j] = zerotier.ControllerNetwork{}
		}
	}
	for i, controller := range controllers {
		client, cerr := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
		for j, id := range ids[i] {
			eg.Go(func(i int, client *zerotier.ClientWithResponses, j int, id string) func() error {
				return func() error {
					if cerr != nil {
						return nil
					}

					res, err := client.GetControllerNetworkWithResponse(ctx, id)
					if err != nil {
						return err
					}

					networks[i][j] = *res.JSON200
					return nil
				}
			}(i, client, j, id))
		}
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	keyedNetworks := make([]map[string]zerotier.ControllerNetwork, len(controllers))
	for i := range controllers {
		keyedNetworks[i] = make(map[string]zerotier.ControllerNetwork, len(ids[i]))
		for j, id := range ids[i] {
			keyedNetworks[i][id] = networks[i][j]
		}
	}

	return keyedNetworks, nil
}

// Network

func GetNetworkInfo(
	c echo.Context, controller Controller, id string,
) (*zerotier.ControllerNetwork, []string, error) {
	client, cerr := zerotier.NewAuthClientWithResponses(
		controller.Server, controller.Authtoken,
	)
	if cerr != nil {
		return nil, nil, cerr
	}

	var network *zerotier.ControllerNetwork
	var memberRevisions map[string]int
	eg, ctx := errgroup.WithContext(c.Request().Context())
	eg.Go(func() error {
		res, err := client.GetControllerNetworkWithResponse(ctx, id)
		if err != nil {
			return err
		}

		network = res.JSON200
		return nil
	})
	eg.Go(func() error {
		res, err := client.GetControllerNetworkMembersWithResponse(ctx, id)
		if err != nil {
			return err
		}

		err = json.Unmarshal(res.Body, &memberRevisions)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	memberAddresses := make([]string, 0, len(memberRevisions))
	for address := range memberRevisions {
		memberAddresses = append(memberAddresses, address)
	}

	return network, memberAddresses, nil
}

func CreateNetwork(c echo.Context, controller Controller) (*zerotier.ControllerNetwork, error) {
	client, cerr := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if cerr != nil {
		return nil, cerr
	}

	ctx := c.Request().Context()
	sRes, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	status := *sRes.JSON200

	private := true
	n6plane := true
	v6AssignMode := zerotier.V6AssignMode{
		N6plane: &n6plane,
		Rfc4193: nil,
		Zt:      nil,
	}
	ipv4Type := 2048
	ipv4ARPType := 2054
	ipv6Type := 34525
	rules := []map[string]interface{}{
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

	body := zerotier.GenerateControllerNetworkJSONRequestBody{}
	body.Private = &private
	body.V6AssignMode = &v6AssignMode
	body.Rules = &rules

	nRes, err := client.GenerateControllerNetworkWithResponse(
		ctx, *status.Address, body,
	)
	if err != nil {
		return nil, err
	}

	return nRes.JSON200, nil
}

func UpdateNetwork(
	c echo.Context,
	controller Controller,
	id string,
	network zerotier.SetControllerNetworkJSONRequestBody,
) error {
	client, err := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	_, err = client.SetControllerNetworkWithResponse(ctx, id, network)
	return err
}

func DeleteNetwork(c echo.Context, controller Controller, id string) error {
	client, err := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	_, err = client.DeleteControllerNetworkWithResponse(ctx, id)
	return err
}

// Network members

func GetNetworkMembersInfo(
	c echo.Context, controller Controller, networkID string, memberAddresses []string,
) (map[string]zerotier.ControllerNetworkMember, error) {
	client, cerr := zerotier.NewAuthClientWithResponses(
		controller.Server, controller.Authtoken,
	)
	if cerr != nil {
		return nil, cerr
	}

	eg, ctx := errgroup.WithContext(c.Request().Context())
	members := make([]zerotier.ControllerNetworkMember, len(memberAddresses))
	for i := range memberAddresses {
		members[i] = zerotier.ControllerNetworkMember{}
	}
	for i, memberAddress := range memberAddresses {
		eg.Go(func(i int, memberAddress string) func() error {
			return func() error {
				res, err := client.GetControllerNetworkMemberWithResponse(
					ctx, networkID, memberAddress,
				)
				if err != nil {
					return err
				}

				members[i] = *res.JSON200
				return nil
			}
		}(i, memberAddress))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	keyedMembers := make(map[string]zerotier.ControllerNetworkMember)
	for i, addr := range memberAddresses {
		keyedMembers[addr] = members[i]
	}

	return keyedMembers, nil
}

// Network member

func UpdateMember(
	c echo.Context,
	controller Controller,
	networkID string,
	memberAddress string,
	member zerotier.SetControllerNetworkMemberJSONRequestBody,
) error {
	client, err := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	_, err = client.SetControllerNetworkMemberWithResponse(
		ctx, networkID, memberAddress, member,
	)
	return err
}
