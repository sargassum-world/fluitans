package zerotier

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

// All Network Members

func (c *Client) GetNetworkMembers(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberAddresses []string,
) (map[string]zerotier.ControllerNetworkMember, error) {
	client, cerr := controller.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	eg, ctx := errgroup.WithContext(ctx)
	members := make([]zerotier.ControllerNetworkMember, len(memberAddresses))
	for i := range memberAddresses {
		members[i] = zerotier.ControllerNetworkMember{}
	}
	for i, memberAddress := range memberAddresses {
		eg.Go(func(i int, memberAddress string) func() error {
			return func() error {
				res, err := client.GetControllerNetworkMemberWithResponse(ctx, networkID, memberAddress)
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

// Individual Network Member

func (c *Client) GetNetworkMember(
	ctx context.Context, controller ztcontrollers.Controller, networkID string, memberAddress string,
) (*zerotier.ControllerNetworkMember, error) {
	client, cerr := controller.NewClient()
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.GetControllerNetworkMemberWithResponse(ctx, networkID, memberAddress)
	if err != nil {
		return nil, err
	}
	return res.JSON200, nil
}

func (c *Client) UpdateMember(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberAddress string, member zerotier.SetControllerNetworkMemberJSONRequestBody,
) error {
	client, err := controller.NewClient()
	if err != nil {
		return err
	}

	_, err = client.SetControllerNetworkMemberWithResponse(ctx, networkID, memberAddress, member)
	return err
}
