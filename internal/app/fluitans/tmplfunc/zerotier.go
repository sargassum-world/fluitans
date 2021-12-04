package tmplfunc

import (
	"strings"

	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func identifyNetwork(network zerotier.ControllerNetwork) string {
	if strings.TrimSpace(*network.Name) != "" {
		return *network.Name
	}
	return *network.Id
}

func getNetworkHostAddress(id string) string {
	return id[:10]
}

func getNetworkNumber(id string) string {
	return id[10:]
}
