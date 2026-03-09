package stats

type PostParams struct {
	IDs, Likes, Comments, Shares []int64
}

type CommentParams struct {
	IDs, Likes, Replies []int64
}
