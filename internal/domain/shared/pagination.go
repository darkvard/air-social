package shared

import (
	"cmp"
	"math"
)

/*
 * ========================================================================================
 * PAGINATION STRATEGIES
 * ========================================================================================
 *
 * This system supports two pagination strategies:
 *   1) Offset Pagination (Page-based)
 *   2) Cursor Pagination (Keyset-based)
 *
 * ----------------------------------------------------------------------------------------
 * 1) OFFSET PAGINATION
 * ----------------------------------------------------------------------------------------
 *
 * Use Cases:
 * - Static or rarely changing datasets
 * - Admin dashboards
 * - Explicit page navigation (e.g., jump to page N)
 *
 * Core Concept:
 * - Calculates how many rows to skip before returning the next batch.
 * - Each request is independent and based on page number.
 *
 * Requirements:
 * - Stable sorting column.
 * - Suitable for datasets where deep pagination is limited.
 *
 * Characteristics:
 * - Simple and intuitive.
 * - Supports arbitrary page jumps.
 * - Performance degrades as page depth increases (linear skip cost).
 * - Susceptible to data drift (duplicates or missing records if data changes).
 *
 *
 * ----------------------------------------------------------------------------------------
 * 2) CURSOR (KEYSET) PAGINATION
 * ----------------------------------------------------------------------------------------
 *
 * Use Cases:
 * - Frequently updated datasets
 * - Infinite scrolling (feeds, timelines, event streams)
 * - Large datasets requiring consistent performance
 *
 * Core Concept:
 * - Uses a unique ordered column as an anchor (Cursor).
 * - Traverses records sequentially, similar to walking a Linked List:
 *     each request continues from the last seen record.
 *
 * Requirements:
 * - Cursor column must be strictly unique.
 * - Must be indexed.
 * - Must have deterministic ordering.
 *
 * Characteristics:
 * - Consistent performance regardless of depth.
 * - Uses index seek instead of row skipping.
 * - Immune to data drift (no duplicate/missing records).
 * - Does not support arbitrary page jumps.
 *
 *
 * ----------------------------------------------------------------------------------------
 * SUMMARY
 * ----------------------------------------------------------------------------------------
 *
 * Offset Pagination:
 *   - Best for page-based navigation and small/medium datasets.
 *
 * Cursor Pagination:
 *   - Best for scalable systems, real-time feeds, and large datasets.
 *
 * Choose based on data volatility and navigation requirements.
 *
 * ========================================================================================
 */

const (
	offsetPageMin  = 1
	offsetLimitMin = 1
	offsetLimitMax = 100
	offsetLimitDefault = 10
)

const (
	cursorLimitMin     = 1
	cursorLimitMax     = 50
	cursorLimitDefault = 10
)

const (
	SortLatest = "latest"
	SortOldest = "oldest"
)

// --- Offset Pagination (Traditional) ---

type OffsetQueryParams struct {
	Page  int
	Limit int
	Sort  string
}

func (q *OffsetQueryParams) NormalizePagination() {
	if q.Page < offsetPageMin {
		q.Page = offsetPageMin
	}
	if q.Limit < offsetLimitMin {
		q.Limit = offsetLimitDefault
	}
	if q.Limit > offsetLimitMax {
		q.Limit = offsetLimitMax
	}
}

func (q OffsetQueryParams) GetOffset() int {
	if q.Page < offsetPageMin {
		return 0
	}
	return (q.Page - 1) * q.Limit
}

func (q OffsetQueryParams) GetSortOrder() string {
	if q.Sort == SortOldest {
		return "ASC"
	}
	return "DESC"
}

type OffsetPaginatedResult[T any] struct {
	Data       []T
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

func NewOffsetPaginatedResult[T any](data []T, total int64, page, limit int) OffsetPaginatedResult[T] {
	var totalPages int
	if data == nil {
		data = []T{}
	}
	if limit > 0 {
		count := float64(total) / float64(limit)
		totalPages = int(math.Ceil(count))
	}
	return OffsetPaginatedResult[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}

// --- Cursor Pagination (Keyset) ---


// Cursorer defines the interface for cursor pagination.
// K (cmp.Ordered) supports numbers and strings.
// WARNING: If K is a string, you MUST use time-sortable IDs (e.g., ULID, UUIDv7).
// DO NOT use random strings (UUIDv4, MD5, uuid.New().String()) as they break chronological sorting in SQL.
type Cursorer[K cmp.Ordered] interface {
	GetCursor() K
}

type CursorQueryParams[K cmp.Ordered] struct {  
	Cursor K
	Limit  int
	Sort   string
}

func (q *CursorQueryParams[K]) NormalizePagination() {
	if q.Limit < cursorLimitMin {
		q.Limit = cursorLimitDefault
	}
	if q.Limit > cursorLimitMax {
		q.Limit = cursorLimitMax
	}
}

// GetFetchLimit returns Limit + 1 to fetch one extra record from the DB 
// to determine if a next page exists.
func (q CursorQueryParams[K]) GetFetchLimit() int {
	return q.Limit + 1
}

func (q CursorQueryParams[K]) GetSortOrder() string {
	if q.Sort == SortOldest {
		return "ASC"
	}
	return "DESC"
}

type CursorPaginatedResult[T any, K cmp.Ordered] struct {
	Data        []T
	NextCursor  K
	HasNextPage bool
}

func NewCursorPaginatedResult[T Cursorer[K], K cmp.Ordered](data []T, limit int) CursorPaginatedResult[T, K] {
	if data == nil {
		data = []T{}
	}

	var hasNextPage bool
	var nextCursor K

	// Since SQL fetched 'Limit + 1' records (via GetFetchLimit), 
    // if len(data) > limit, it proves there is a next page.
	if len(data) > limit {
		hasNextPage = true
		// Remove the extra element to return exactly the requested limit
		data = data[:limit]
	}

	if hasNextPage && len(data) > 0 {
		nextCursor = data[len(data)-1].GetCursor()
	}

	return CursorPaginatedResult[T, K]{
		Data:        data,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}
}
