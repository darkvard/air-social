package chat

import (
	"time"
)

type ReactionType string

const (
	ReactionLike  ReactionType = "like"
	ReactionLove  ReactionType = "love"
	ReactionHaha  ReactionType = "haha"
	ReactionWow   ReactionType = "wow"
	ReactionSad   ReactionType = "sad"
	ReactionAngry ReactionType = "angry"
)

var allowedReactions = map[ReactionType]bool{
	ReactionLike: true, ReactionLove: true, ReactionHaha: true,
	ReactionWow: true, ReactionSad: true, ReactionAngry: true,
}

func (r ReactionType) IsValid() bool {
	return allowedReactions[r]
}

type Reaction struct {
	UserID    int64
	Type      ReactionType
	CreatedAt time.Time
}
