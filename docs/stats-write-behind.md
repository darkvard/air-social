# Stats — Write-Behind Cache with Base + Delta

## Design

Counting likes, comments, and shares with direct DB writes causes a **table lock contention** problem — every interaction hits the same row and increments the same counter, serialising updates under high concurrency.

The solution is **write-behind caching**: interactions never write to the DB directly. Instead they update a delta counter in Redis immediately, and a background syncer periodically flushes those deltas into PostgreSQL in bulk.

Reading stats combines both sources at query time:

```
Final count = max(0, Base_DB + Delta_Redis)
```

---

## Overview

```
User action (like / comment / share)
         │
         ▼
  publish event → RabbitMQ
         │
    Stats Worker
         │
   HINCRBY delta in Redis Hash
   (one field per entity ID)

         │ every 1 minute
    Stats Syncer (cron)
         │
   read all deltas from Redis
         │
   BulkUpsert → PostgreSQL
         │
   subtract synced values from Redis (Lua)
```

---

## Redis Data Structure

Stats are stored as **Redis Hash** — one hash per counter type:

```
Key:    social:stats:{state}:
Field:  entityID  (post_id or comment_id, as string)
Value:  delta     (signed int64, can be negative)
```

States tracked:

| State            | Tracks                         |
|------------------|--------------------------------|
| `post_likes`     | like / unlike on posts         |
| `post_shares`    | share / unshare on posts       |
| `post_comments`  | comment created / deleted      |
| `comment_likes`  | like / unlike on comments      |
| `comment_replies`| reply created / deleted on comments |

Each worker event maps to `HINCRBY key field +1` or `HINCRBY key field -1`. Multiple events on the same entity accumulate in the same field — no separate rows, no locking.

---

## Write Path — Stats Worker

Events arrive from RabbitMQ, dispatcher routes to the right handler:

```
EventPostLike      → HINCRBY post_likes     {postID}    ±1
EventPostShare     → HINCRBY post_shares    {postID}    ±1
EventCommentLike   → HINCRBY comment_likes  {commentID} ±1
EventCommentCreated→ HINCRBY post_comments  {postID}    +1
                     HINCRBY comment_replies {parentID}  +1  (if reply)
EventCommentDeleted→ HINCRBY post_comments  {postID}    -1
                     HINCRBY comment_replies {parentID}  -1  (if reply)
```

No DB touch. Redis `HINCRBY` is atomic — concurrent updates on the same field are safe without any application-level locking.

---

## Read Path — Base + Delta

When a post or comment is displayed, stats are fetched concurrently from two sources:

```
DB query (base)    ──┐
                     ├─► merge → Final = max(0, base + delta)
Redis HMGET (delta) ─┘
```

`max(0, ...)` prevents showing negative counts caused by temporary delta drift (e.g. unlike arrives before like is synced).

---

## Sync Path — Write-Behind Flush

A **Stats Syncer** runs on a 1-minute ticker:

1. `HGetAll` all fields from each Redis hash state — gets the full delta snapshot
2. Deduplicate entity IDs across all states (a post may have likes, comments, and shares all pending)
3. `BulkUpsert` into PostgreSQL using `INSERT ... ON CONFLICT DO UPDATE SET likes = likes + excluded.likes` — one query per entity type, not one query per entity
4. Subtract the synced values from Redis using a **Lua script** (not HDEL):

```lua
local current = redis.call('HINCRBY', KEYS[1], ARGV[1], -syncedValue)
if current == 0 then
    redis.call('HDEL', KEYS[1], ARGV[1])
end
```

The Lua script is necessary because between step 1 (snapshot) and step 4 (clear), new events may have arrived and incremented the same field. A plain `HDEL` would wipe those new deltas. The atomic decrement-then-conditionally-delete ensures only the synced amount is removed.

---

## Reconciliation (Fallback)

If `BulkUpsert` fails (DB down, connection error), the syncer immediately triggers a **reconciliation** in a background goroutine:

```
ReconcilePostStats → re-COUNT from source-of-truth tables
                   → overwrite DB with correct values
```

Reconciliation bypasses Redis entirely — it recalculates counts directly from `likes`, `comments`, `shares` tables via SQL `COUNT`. Slower but guaranteed correct. This runs as a fallback, not on the happy path.

---

## Notes

- Delta can go negative temporarily (e.g. unlike before the like is synced to DB). `max(0, base + delta)` handles this at read time.
- Syncer interval is 1 minute — stats may lag by up to 1 minute before appearing in the DB, but the read path always shows the real-time value via Redis.
- If Redis goes down, the read path degrades gracefully to DB-only (base without delta). Events that fail to write to Redis are retried up to 3 times then go to DLQ.
