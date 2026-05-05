package notification

import "air-social/internal/domain/common"

type GetParams struct {
	UserID int64
	Query  common.CursorQueryParams[int64]
}
