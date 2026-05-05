CREATE TABLE
    IF NOT EXISTS notifications (
        id BIGSERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL,
        actor_id BIGINT NOT NULL,
        type VARCHAR(50) NOT NULL,
        target_id BIGINT,
        target_type VARCHAR(50),
        data JSONB,
        read BOOLEAN NOT NULL DEFAULT FALSE,
        read_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX idx_notifications_inbox ON notifications (user_id, read, created_at DESC);

CREATE UNIQUE INDEX uniq_notif_with_target ON notifications (user_id, actor_id, type, target_id)
WHERE
    target_id IS NOT NULL;

CREATE UNIQUE INDEX uniq_notif_no_target ON notifications (user_id, actor_id, type)
WHERE
    target_id IS NULL;