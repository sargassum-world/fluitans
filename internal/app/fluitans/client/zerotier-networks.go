package client

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

// All Networks

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

// Individual Network

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
