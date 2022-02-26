package tmplfunc

import (
	"strings"

	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func IdentifyNetwork(network zerotier.ControllerNetwork) string {
	if strings.TrimSpace(*network.Name) != "" {
		return *network.Name
	}
	return *network.Id
}

func GetNetworkHostAddress(id string) string {
	return id[:10]
}

func GetNetworkNumber(id string) string {
	return id[10:]
}
