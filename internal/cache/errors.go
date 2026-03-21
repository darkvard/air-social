package cache

import "errors"

var ErrCacheMiss = errors.New("cache: key not found")

func Miss(err error) bool {
	return errors.Is(err, ErrCacheMiss)
}
