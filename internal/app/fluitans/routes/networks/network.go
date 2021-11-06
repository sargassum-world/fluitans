package networks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func getNetworkInfo(
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

func getNetworkMembersInfo(
	c echo.Context, controller Controller, id string, addresses []string,
) (map[string]zerotier.ControllerNetworkMember, error) {
	client, cerr := zerotier.NewAuthClientWithResponses(
		controller.Server, controller.Authtoken,
	)
	if cerr != nil {
		return nil, cerr
	}

	eg, ctx := errgroup.WithContext(c.Request().Context())
	members := make([]zerotier.ControllerNetworkMember, len(addresses))
	for i := range addresses {
		members[i] = zerotier.ControllerNetworkMember{}
	}
	for i, address := range addresses {
		eg.Go(func(i int, address string) func() error {
			return func() error {
				res, err := client.GetControllerNetworkMemberWithResponse(ctx, id, address)
				if err != nil {
					return err
				}

				members[i] = *res.JSON200
				return nil
			}
		}(i, address))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	keyedMembers := make(map[string]zerotier.ControllerNetworkMember)
	for i, addr := range addresses {
		keyedMembers[addr] = members[i]
	}

	return keyedMembers, nil
}

func getControllerAddress(networkID string) string {
	addressLength := 10
	return networkID[:addressLength]
}

func addNDPAddresses(
	id string,
	v6AssignMode zerotier.V6AssignMode,
	members map[string]zerotier.ControllerNetworkMember,
) error {
	n6plane := v6AssignMode.N6plane
	if n6plane != nil && *n6plane {
		for _, member := range members {
			n6PlaneAddress, err := zerotier.Get6Plane(id, *member.Address)
			if err != nil {
				return err
			}

			if member.IpAssignments == nil {
				ipAssignments := []string{n6PlaneAddress}
				member.IpAssignments = &ipAssignments
			} else {
				*member.IpAssignments = append(*member.IpAssignments, n6PlaneAddress)
			}
		}
	}
	return nil
}

type NetworkData struct {
	Controller       Controller
	Network          zerotier.ControllerNetwork
	Members          map[string]zerotier.ControllerNetworkMember
	JSONPrintedRules string
}

func getNetworkData(c echo.Context, address string, id string) (*NetworkData, error) {
	controller, err := findControllerByAddress(c, storedControllers, address)
	if err != nil {
		return nil, err
	}

	network, memberAddresses, err := getNetworkInfo(c, *controller, id)
	if err != nil {
		return nil, err
	}

	members, err := getNetworkMembersInfo(c, *controller, id, memberAddresses)
	if err != nil {
		return nil, err
	}

	err = addNDPAddresses(id, *network.V6AssignMode, members)
	if err != nil {
		return nil, err
	}

	rules, err := json.MarshalIndent(*network.Rules, "", "  ")
	if err != nil {
		return nil, err
	}

	return &NetworkData{
		Controller:       *controller,
		Network:          *network,
		Members:          members,
		JSONPrintedRules: string(rules),
	}, nil
}

func network(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks/network.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Parse params
		id := c.Param("id")
		address := getControllerAddress(id)

		// Run queries
		networkData, err := getNetworkData(c, address, id)
		if err != nil {
			return err
		}

		// Handle Etag
		etagData, err := json.Marshal(networkData)
		if err != nil {
			return err
		}

		if noContent, err := caching.ProcessEtag(c, tte, fingerprint.Compute(etagData)); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta   template.Meta
			Embeds template.Embeds
			Data   NetworkData
		}{
			Meta: template.Meta{
				Path:       c.Request().URL.Path,
				DomainName: os.Getenv("FLUITANS_DOMAIN_NAME"),
			},
			Embeds: g.Embeds,
			Data:   *networkData,
		})
	}, nil
}

func setNetworkName(c echo.Context, controller Controller, id string, name string) error {
	client, err := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	_, err = client.SetControllerNetworkWithResponse(
		ctx, id,
		zerotier.SetControllerNetworkJSONRequestBody{Name: &name},
	)
	return err
}

func setNetworkRules(c echo.Context, controller Controller, id string, jsonRules string) error {
	client, err := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if err != nil {
		return err
	}

	rules := make([]map[string]interface{}, 0)
	if err = json.Unmarshal([]byte(jsonRules), &rules); err != nil {
		return err
	}

	ctx := c.Request().Context()
	_, err = client.SetControllerNetworkWithResponse(
		ctx, id,
		zerotier.SetControllerNetworkJSONRequestBody{Rules: &rules},
	)
	return err
}

func deleteNetwork(c echo.Context, controller Controller, id string) error {
	client, err := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	_, err = client.DeleteControllerNetworkWithResponse(ctx, id)
	return err
}

func postNetwork(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	return func(c echo.Context) error {
		// Parse params
		id := c.Param("id")
		address := getControllerAddress(id)
		method := c.FormValue("method")
		name := c.FormValue("name")
		rules := c.FormValue("rules")

		// Run queries
		controller, err := findControllerByAddress(c, storedControllers, address)
		if err != nil {
			return err
		}

		switch method {
		case "RENAME":
			err = setNetworkName(c, *controller, id, name)
			if err != nil {
				return err
			}
		case "SETRULES":
			err = setNetworkRules(c, *controller, id, rules)
			if err != nil {
				return err
			}

			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s#network-%s-rules", id, id))
		case "DELETE":
			err = deleteNetwork(c, *controller, id)
			if err != nil {
				return err
			}

			return c.Redirect(http.StatusSeeOther, "/networks")
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", id))
	}, nil
}
