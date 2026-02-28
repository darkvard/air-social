package like

type PathIDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}
