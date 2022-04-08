// Package handling provides utilities for handlers
package handling

import (
	"context"
	"time"
)

func Repeat(ctx context.Context, interval time.Duration, f func() (done bool, err error)) error {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := ctx.Err(); err != nil {
				// Context was also canceled and it should have priority
				return err
			}

			done, err := f()
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}
