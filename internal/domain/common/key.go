package common

import (
	"fmt"
	"strings"
)

// BuildCacheKey builds a Redis key by joining parts with ":".
// Example: BuildCacheKey("user", "summary", 16) → "user:summary:16"
func BuildCacheKey(parts ...any) string {
	strs := make([]string, len(parts))
	for i, p := range parts {
		strs[i] = fmt.Sprintf("%v", p)
	}
	return strings.Join(strs, ":")
}
