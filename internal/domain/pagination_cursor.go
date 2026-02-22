package domain
// todo: remove

/*
 * ========================================================================================
 * CURSOR (KEYSET) PAGINATION ALGORITHM
 * ========================================================================================
 *
 * A.Mechanism:
 * Traverses an index similarly to a Linked List. Utilizes the ID of the last seen record
 * as an anchor (Cursor) to directly fetch the next batch of `Limit` items.
 *
 * B.Prerequisites:
 * 1. Unique Cursor: Must be a strictly unique column (e.g., auto-increment `id` or Snowflake).
 * 2. Sorted & Directional: The query MUST pair `ORDER BY` with the correct `WHERE` operator:
 * - Descending (Newest first): `WHERE id < cursor ORDER BY id DESC`
 * - Ascending  (Oldest first): `WHERE id > cursor ORDER BY id ASC`
 * * Note: For the initial request (cursor = 0), omit the `WHERE` clause.
 *
 * C.Performance Comparison:
 * +-----------------+------------------------------------------+--------------------------------------+
 * | Metric          | Cursor Pagination                        | Offset Pagination                    |
 * +-----------------+------------------------------------------+--------------------------------------+
 * | DB Execution    | B-Tree Index Seek (O(log N) jump)        | Sequential Scan & Discard            |
 * | Time Complexity | O(1) - Constant performance at any depth | O(N) - Linear degradation per page   |
 * | Data Integrity  | Absolute - Immune to Data Drift          | High risk of duplicates/missing data |
 * | Use Case        | Infinite Scroll, Real-time Feeds         | Admin Dashboards, Explicit Pages     |
 * +-----------------+------------------------------------------+--------------------------------------+
 */
const (
	cursorLimitMin     = 1
	cursorLimitMax     = 50
	cursorLimitDefault = 10
)

type CursorQueryParams struct {
	Cursor int64
	Limit  int
}

func (q *CursorQueryParams) EnsureDefaults() {
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
