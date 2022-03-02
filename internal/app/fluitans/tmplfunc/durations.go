package tmplfunc

import (
	"time"
)

func DurationToSec(i time.Duration) float64 {
	return i.Seconds()
}
