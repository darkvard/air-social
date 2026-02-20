package shared

import "math"

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
)

const (
	cursorLimitMin     = 1
	cursorLimitMax     = 50
	cursorLimitDefault = 10
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
		q.Limit = offsetLimitMin
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

type CursorQueryParams struct {
	Cursor int64
	Limit  int
}

func (q *CursorQueryParams) NormalizePagination() {
	if q.Limit < cursorLimitMin {
		q.Limit = cursorLimitDefault
	}
	if q.Limit > cursorLimitMax {
		q.Limit = cursorLimitMax
	}
}

type CursorPaginatedResult[T any] struct {
	Data        []T
	NextCursor  int64
	HasNextPage bool
}

func NewCursorPaginatedResult[T any](data []T, nextCursor int64, hasNext bool) CursorPaginatedResult[T] {
	if data == nil {
		data = []T{}
	}
	return CursorPaginatedResult[T]{
		Data:        data,
		NextCursor:  nextCursor,
		HasNextPage: hasNext,
	}
}
