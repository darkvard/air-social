package pkg

import (
	"context"
	"fmt"
	"reflect"
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

func TimeNowUTC() time.Time {
	return time.Now().UTC()
}

// IsStructNilOrEmpty checks if the input is nil, a nil pointer, or a struct with all zero/nil fields.
// It is primarily used to validate PATCH requests where an empty payload ({}) must be rejected.
func IsStructNilOrEmpty(d any) bool {
	// 1. Safety check: nil interface is strictly empty.
	if d == nil {
		return true
	}

	v := reflect.ValueOf(d)

	// 2. Handle Pointers: Check if the pointer itself is nil before dereferencing.
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return true // Nil pointer (e.g., var req *Request = nil) is empty.
		}
		v = v.Elem() // Dereference to access the actual struct value.
	}

	// 3. Type validation: Ensure we are inspecting a Struct.
	if v.Kind() != reflect.Struct {
		return false
	}

	// 4. Field Inspection: Iterate through all fields to find any non-zero value.
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		// Skip unexported fields to prevent runtime panic.
		if !field.CanInterface() {
			continue
		}

		// Check if the field has a value.
		if !field.IsZero() {
			return false
		}
	}

	// 5. Conclusion: No non-zero fields found -> Struct is empty.
	return true
}

// IsNil checks if an interface holds a nil value, handling the "Typed Nil" problem in Go.
func IsNil(i any) bool {
	if i == nil {
		return true
	}

	v := reflect.ValueOf(i)
	kind := v.Kind()

	// Only check IsNil for types that support it to avoid panic.
	if kind == reflect.Pointer ||
		kind == reflect.Interface ||
		kind == reflect.Map ||
		kind == reflect.Slice ||
		kind == reflect.Chan ||
		kind == reflect.Func {
		return v.IsNil()
	}

	return false
}