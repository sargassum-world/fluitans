package client

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

// All Network Members

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

// Individual Network Member

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
