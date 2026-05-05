package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	notidomain "air-social/internal/domain/notification"
	"air-social/internal/infrastructure/postgres"
	"air-social/pkg"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

type notificationRow struct {
	ID         int64      `db:"id"`
	UserID     int64      `db:"user_id"`
	ActorID    int64      `db:"actor_id"`
	Type       string     `db:"type"`
	TargetID   *int64     `db:"target_id"`
	TargetType string     `db:"target_type"`
	Data       []byte     `db:"data"`
	Read       bool       `db:"read"`
	ReadAt     *time.Time `db:"read_at"`
	CreatedAt  time.Time  `db:"created_at"`
}

func (r notificationRow) toDomain() *notidomain.Notification {
	n := &notidomain.Notification{
		ID:         r.ID,
		UserID:     r.UserID,
		ActorID:    r.ActorID,
		Type:       notidomain.NotificationType(r.Type),
		TargetID:   r.TargetID,
		TargetType: r.TargetType,
		Read:       r.Read,
		ReadAt:     r.ReadAt,
		CreatedAt:  r.CreatedAt,
	}
	if len(r.Data) > 0 {
		_ = json.Unmarshal(r.Data, &n.Data)
	}
	return n
}

// Upsert inserts or updates a notification.
// Uses two separate partial-index conflict targets because PostgreSQL partial indexes
// require the exact same WHERE predicate in the ON CONFLICT clause.
func (r *repository) Upsert(ctx context.Context, n *notidomain.Notification) error {
	dataJSON, err := json.Marshal(n.Data)
	if err != nil {
		return pkg.ErrInternal
	}

	var scanErr error
	if n.TargetID != nil {
		query := `
			INSERT INTO notifications (user_id, actor_id, type, target_id, target_type, data, read, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, false, NOW())
			ON CONFLICT (user_id, actor_id, type, target_id) WHERE target_id IS NOT NULL
			DO UPDATE SET read = false, created_at = NOW(), data = EXCLUDED.data
			RETURNING id, created_at
		`
		scanErr = r.db.QueryRowContext(ctx, query,
			n.UserID, n.ActorID, string(n.Type), n.TargetID, n.TargetType, dataJSON,
		).Scan(&n.ID, &n.CreatedAt)
	} else {
		query := `
			INSERT INTO notifications (user_id, actor_id, type, target_id, target_type, data, read, created_at)
			VALUES ($1, $2, $3, NULL, $4, $5, false, NOW())
			ON CONFLICT (user_id, actor_id, type) WHERE target_id IS NULL
			DO UPDATE SET read = false, created_at = NOW(), data = EXCLUDED.data
			RETURNING id, created_at
		`
		scanErr = r.db.QueryRowContext(ctx, query,
			n.UserID, n.ActorID, string(n.Type), n.TargetType, dataJSON,
		).Scan(&n.ID, &n.CreatedAt)
	}
	return pkg.MapPostgresError(scanErr)
}

func (r *repository) Delete(ctx context.Context, userID, actorID int64, typ notidomain.NotificationType, targetID *int64) error {
	var err error
	if targetID != nil {
		_, err = r.db.ExecContext(ctx,
			`DELETE FROM notifications WHERE user_id=$1 AND actor_id=$2 AND type=$3 AND target_id=$4`,
			userID, actorID, string(typ), *targetID,
		)
	} else {
		_, err = r.db.ExecContext(ctx,
			`DELETE FROM notifications WHERE user_id=$1 AND actor_id=$2 AND type=$3 AND target_id IS NULL`,
			userID, actorID, string(typ),
		)
	}
	return pkg.MapPostgresError(err)
}

func (r *repository) GetList(ctx context.Context, params notidomain.GetParams) ([]notidomain.Notification, error) {
	var args []any
	var builder strings.Builder

	args = append(args, params.UserID)
	argID := 2

	builder.WriteString(`SELECT * FROM notifications WHERE user_id = $1`)

	if params.Query.Cursor > 0 {
		fmt.Fprintf(&builder, " AND id < $%d", argID)
		args = append(args, params.Query.Cursor)
		argID++
	}

	fmt.Fprintf(&builder, " ORDER BY id DESC LIMIT $%d", argID)
	args = append(args, params.Query.GetFetchLimit())

	var rows []notificationRow
	if err := r.db.SelectContext(ctx, &rows, builder.String(), args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make([]notidomain.Notification, len(rows))
	for i, row := range rows {
		result[i] = *row.toDomain()
	}
	return result, nil
}

func (r *repository) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`,
		userID,
	).Scan(&count); err != nil {
		return 0, pkg.MapPostgresError(err)
	}
	return count, nil
}

// MarkRead marks a single notification as read. Idempotent: returns nil if already read.
func (r *repository) MarkRead(ctx context.Context, userID, id int64) error {
	return postgres.ExecOne(ctx, r.db,
		`UPDATE notifications SET read = true, read_at = NOW() WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
}

// MarkAllRead marks all unread notifications as read. Safe to call even if none are unread.
func (r *repository) MarkAllRead(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read = true, read_at = NOW() WHERE user_id = $1 AND read = false`,
		userID,
	)
	return pkg.MapPostgresError(err)
}
