package chat

import (
	"time"
)

type ReactionType string

const (
	ReactionLike  ReactionType = "👍"
	ReactionLove  ReactionType = "❤️"
	ReactionHaha  ReactionType = "😂"
	ReactionWow   ReactionType = "😮"
	ReactionSad   ReactionType = "😢"
	ReactionAngry ReactionType = "😠"
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
