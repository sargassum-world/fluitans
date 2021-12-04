package desec

import (
	"time"
)

type DesecAPISettings struct {
	ReadCacheTTL   time.Duration
	WriteSoftQuota float32
}
