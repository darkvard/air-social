package dto

type IDPathParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}
