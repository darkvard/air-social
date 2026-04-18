package modules

import (
	"context"
	"log"
	"time"

	"github.com/oklog/ulid/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"air-social/cmd/seed/config"
)

type participantSeedDoc struct {
	UserID   int64     `bson:"user_id"`
	Role     string    `bson:"role"`
	State    string    `bson:"state"`
	JoinedAt time.Time `bson:"joined_at"`
}

type conversationSeedDoc struct {
	ID           string               `bson:"_id"`
	Type         string               `bson:"type"`
	Participants []participantSeedDoc `bson:"participants"`
	Name         string               `bson:"name,omitempty"`
	CreatedBy    int64                `bson:"created_by"`
	CreatedAt    time.Time            `bson:"created_at"`
	UpdatedAt    time.Time            `bson:"updated_at"`
}

func SeedConversations(db *mongo.Database, userIDs []int64, cfg config.SeedConfig) {
	if len(userIDs) < 5 {
		log.Println("Seeded: 0 Conversations (not enough users)")
		return
	}

	coll := db.Collection("conversations")
	ctx := context.Background()
	now := time.Now().UTC()
	testerID := userIDs[0]

	var docs []interface{}
	offset := 0

	// --- Direct active: tester ↔ other users ---
	// Each with a different UpdatedAt so cursor pagination works correctly.
	directLimit := min(cfg.Conversations.DirectActive, len(userIDs)-1)
	for i := 0; i < directLimit; i++ {
		t := now.Add(-time.Duration(offset) * time.Hour)
		offset++
		docs = append(docs, conversationSeedDoc{
			ID:   ulid.Make().String(),
			Type: "direct",
			Participants: []participantSeedDoc{
				{UserID: testerID, Role: "member", State: "active", JoinedAt: t},
				{UserID: userIDs[i+1], Role: "member", State: "active", JoinedAt: t},
			},
			CreatedBy: testerID,
			CreatedAt: t,
			UpdatedAt: t,
		})
	}

	// --- Group conversations: tester as admin ---
	groupDefs := []struct {
		name    string
		members []int64
	}{
		{"Dev Team", []int64{userIDs[1], userIDs[2], userIDs[3]}},
		{"Design Squad", []int64{userIDs[4], userIDs[5], userIDs[6]}},
	}
	groupLimit := min(cfg.Conversations.Groups, len(groupDefs))
	for g := 0; g < groupLimit; g++ {
		t := now.Add(-time.Duration(offset) * time.Hour)
		offset++

		members := []participantSeedDoc{
			{UserID: testerID, Role: "admin", State: "active", JoinedAt: t},
		}
		for _, memberID := range groupDefs[g].members {
			members = append(members, participantSeedDoc{
				UserID:   memberID,
				Role:     "member",
				State:    "active",
				JoinedAt: t,
			})
		}
		docs = append(docs, conversationSeedDoc{
			ID:           ulid.Make().String(),
			Type:         "group",
			Name:         groupDefs[g].name,
			Participants: members,
			CreatedBy:    testerID,
			CreatedAt:    t,
			UpdatedAt:    t,
		})
	}

	// --- Pending: last N users DM'd tester (tester's state = pending) ---
	pendingLimit := min(cfg.Conversations.DirectPending, len(userIDs)-1)
	pendingStart := len(userIDs) - pendingLimit
	for i := pendingStart; i < len(userIDs); i++ {
		t := now.Add(-time.Duration(offset) * time.Hour)
		offset++
		docs = append(docs, conversationSeedDoc{
			ID:   ulid.Make().String(),
			Type: "direct",
			Participants: []participantSeedDoc{
				{UserID: userIDs[i], Role: "member", State: "active", JoinedAt: t},
				{UserID: testerID, Role: "member", State: "pending", JoinedAt: t},
			},
			CreatedBy: userIDs[i],
			CreatedAt: t,
			UpdatedAt: t,
		})
	}

	if _, err := coll.InsertMany(ctx, docs); err != nil {
		log.Panicf("batch insert conversations failed: %v", err)
	}

	log.Printf("Seeded: %d Conversations (%d direct active, %d groups, %d pending)",
		len(docs), directLimit, groupLimit, pendingLimit)
}

func TruncateConversations(db *mongo.Database) {
	ctx := context.Background()
	if _, err := db.Collection("conversations").DeleteMany(ctx, bson.M{}); err != nil {
		log.Panicf("cannot clean conversations: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
