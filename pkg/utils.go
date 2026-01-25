package pkg

import (
	"context"
	"fmt"
	"time"
)

const (
	DEVELOPMENT = "development"
	PRODUCTION  = "production"
	DEBUG       = "debug"
)

// FormatTTLVerbose converts a time.Duration (TTL) into a human-readable string.
// Example outputs:
//
//	49h35m  -> "2 days 1 hour 35 minutes"
//	26h     -> "1 day 2 hours"
//	45m     -> "45 minutes"
//	<= 0    -> "expired"
//
// The function progressively extracts:
//   - days  (24h blocks)
//   - hours (remaining hours)
//   - minutes (remaining minutes)
//
// Larger units take priority; smaller units may be omitted depending on UX rules.
func FormatTTLVerbose(d time.Duration) string {
	// Expired or invalid duration
	if d <= 0 {
		return "expired"
	}

	// Extract number of full days
	days := int(d / (24 * time.Hour))
	d -= time.Duration(days) * 24 * time.Hour

	// Extract remaining full hours
	hours := int(d / time.Hour)
	d -= time.Duration(hours) * time.Hour

	// Extract remaining full minutes
	minutes := int(d / time.Minute)

	// Choose output format based on available units
	switch {
	// Days + hours + minutes
	case days > 0 && hours > 0 && minutes > 0:
		return fmt.Sprintf(
			"%s %s %s",
			plural(days, "day"),
			plural(hours, "hour"),
			plural(minutes, "minute"),
		)

	// Days + hours
	case days > 0 && hours > 0:
		return fmt.Sprintf(
			"%s %s",
			plural(days, "day"),
			plural(hours, "hour"),
		)

	// Only days
	case days > 0:
		return plural(days, "day")

	// Only hours
	case hours > 0:
		return plural(hours, "hour")

	// Only minutes
	case minutes > 0:
		return plural(minutes, "minute")

	// Fallback to seconds (very small durations)
	default:
		return plural(int(d.Seconds()), "second")
	}
}

func plural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

func Retry(ctx context.Context, attempts int, sleep time.Duration, fn func() error) error {
	for i := range attempts {
		if err := fn(); err == nil {
			return nil // success
		}

		if i == attempts-1 { // failed
			return fmt.Errorf("after %d attempts, last error: %w", attempts, fn())
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(sleep):
			continue
		}
	}

	return fmt.Errorf("retry failed after %d attempts", attempts)
}
