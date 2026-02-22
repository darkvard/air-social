package follow

import "air-social/internal/domain/common"

const (
	SortLatest   = "latest"
	SortOldest   = "oldest"
	SortNameASC  = "name_asc"
	SortNameDESC = "name_desc"
)

type GetFollowsParams struct {
	ViewerID int64
	TargetID int64
	Paging   common.OffsetQueryParams
}
