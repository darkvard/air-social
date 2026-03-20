# Newsfeed — Fanout-on-Write

## Design

When a user creates a post, the system **proactively pushes (fanout)** that post into the feed cache of every follower immediately — rather than having each follower aggregate their own feed on demand.

This is the **fanout-on-write** strategy (write-heavy, read-cheap):
- **Write path** is heavier: each post triggers N writes to Redis (N = follower count)
- **Read path** is fast: fetching the feed is just one Redis lookup + one batch DB query

To avoid slowing down the create-post API, the entire fanout runs **asynchronously** via RabbitMQ.

---

## Overview

```
CreatePost ──► PostgreSQL (persist)
           └─► RabbitMQ: EventPostCreated
                              │
                         Feed Worker
                              │
                    GetFollowerIDs (Postgres)
                              │
                    ZADD into Redis ZSET
                    for each follower
                              │
                    GetNewsfeed ◄── user opens app
                    (read from Redis, hydrate from DB)
```

---

## Write Path — Fanout

**Trigger:** after `postRepo.Create` succeeds, the post usecase publishes an event:

```
EventPostCreated {
    post_id:   int64
    author_id: int64
    timestamp: post.CreatedAt.UnixMilli()   ← used as the score in Redis
}
```

**Feed Worker processing:**

1. Fetch `followerIDs` of the author from PostgreSQL
2. Append `authorID` to the list (author sees their own post)
3. Write to Redis via Pipeline for all followers in one round-trip:

```
ZADD  social:newsfeed:user:{id}  {timestamp}  {postID}
ZREMRANGEBYRANK  social:newsfeed:user:{id}  0  -501
```

`ZREMRANGEBYRANK 0 -501` keeps the 500 newest posts and automatically evicts the oldest when the feed exceeds the cap.

**On post deletion**, the same flow in reverse — `EventPostDeleted` → `ZREM` the postID from every follower's feed.

---

## Redis Data Structure

```
Key:    social:newsfeed:user:{userID}
Type:   Sorted Set
Member: postID
Score:  CreatedAt (Unix millisecond)
```

Higher score = newer post. Querying in descending score order gives a newest-first timeline.

---

## Read Path — GetNewsfeed

Cursor pagination is **timestamp-based** (not postID-based):

```
Page 1:  cursor = 0  →  ZREVRANGEBYSCORE ... +inf  -inf  LIMIT 0  21
Page 2:  cursor = T  →  ZREVRANGEBYSCORE ... (T    -inf  LIMIT 0  21
                                              ↑ "(" means exclusive
```

Fetch `limit+1` from Redis to detect whether a next page exists. If 21 items come back with limit=20, `has_next_page = true`, trim to 20, `next_cursor = CreatedAt.UnixMilli()` of the 20th item.

Once the ordered `postIDs` list is obtained from Redis, hydrate with full post data from DB:

```
postIDs (ordered by Redis) → GetPostsByIDs (SQL IN, unordered)
                           → build map[postID]*Post
                           → walk postIDs → reconstruct ordered list
```

SQL `IN` does not guarantee order. Use the map for O(1) lookups, then walk the Redis-ordered `postIDs` to restore the correct sequence — O(N) overall.

---

## Idempotency & Retry (Consumer)

Each message is protected by a Redis lock to prevent double-processing:

```
SetNX("worker:feed:processed:{msgID}", "processing", 10m)
  └─ false → already processing or done → Ack, skip
  └─ true  → proceed
               ├─ success → Set key = "done", TTL 24h → Ack
               └─ failure → retry up to 3 times → DLQ
```

---

## Notes

- Feed only receives posts from the moment of follow; no backfill of older posts.
- A share post publishes two independent events: `EventPostShare` → Stats Worker, `EventPostCreated` → Feed Worker.
- If a `EventPostDeleted` event fails, the postID stays in Redis but the DB returns nothing for it — silently filtered out at query time, no visible error.
