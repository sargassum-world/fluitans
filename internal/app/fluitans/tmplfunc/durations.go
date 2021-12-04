package tmplfunc

import (
	"time"
)

func durationToSec(i time.Duration) float64 {
	return i.Seconds()
}
