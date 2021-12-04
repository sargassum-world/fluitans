package conf

import (
	"os"
)

func getDomainName() string {
	return os.Getenv("FLUITANS_DOMAIN_NAME")
}
