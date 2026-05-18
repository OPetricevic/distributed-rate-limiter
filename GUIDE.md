# Distributed Rate Limiter — Implementation Guide

This document walks through what you need to build, step by step. Read it, think about what each piece requires, then implement. Come back to validate your thinking before or after each step.

---

## Phase 1: The Token Bucket (Local, Single Key)

**Goal:** A pure algorithm that tracks how many requests a single client can make.

**What is a token bucket?**
Imagine a bucket that holds tokens. It starts full. Every request costs 1 token. Tokens refill at a fixed rate (e.g., 10 per second). If the bucket is empty, the request is denied.

**What you need to figure out:**
- A struct that holds: max capacity, current token count, refill rate (tokens per second), last refill timestamp
- A method `Allow(cost int) bool` — deducts tokens if available, returns whether the request is allowed
- A method `Refill()` — calculates how many tokens to add based on elapsed time since last refill
- Burst handling: the bucket can hold up to 2x the rate for 1 second, but the average over a longer window must respect the configured rate

**Key decisions:**
- Do you refill lazily (on each `Allow` call) or eagerly (background ticker)?
- How do you handle time? `time.Now()` directly, or inject a clock interface for testing?
- What type is the token count? `int64`? `float64`? Think about precision when refilling fractional tokens.

**Concurrency:**
- Multiple goroutines will call `Allow()` on the same bucket simultaneously
- You need synchronization — `sync.Mutex` is the simplest choice
- Think about whether you lock the entire bucket or can be more granular

**Test cases to write:**
1. Fresh bucket allows requests up to capacity
2. Depleted bucket denies requests
3. After waiting, tokens refill and requests are allowed again
4. Burst: 2x tokens allowed in a short window, then clamped
5. Concurrent access doesn't corrupt state

**Files:**
- `internal/bucket/token.go`
- `internal/bucket/token_test.go`

---

## Phase 2: Key Storage (Multiple Clients)

**Goal:** Track rate limits for many clients simultaneously by key (e.g., API key, IP address).

**What you need to figure out:**
- A map from `string` (the key) to `*Bucket`
- Thread-safe access — many goroutines will look up different keys at the same time
- A `GetOrCreate(key string, config BucketConfig) *Bucket` method that atomically returns an existing bucket or creates a new one

**The hot key problem:**
- One key getting 100k req/s shouldn't slow down lookups for other keys
- Solution: shard the map. Instead of one big `map[string]*Bucket` with one lock, use N maps (e.g., 256) each with their own lock. Hash the key to pick the shard.

**Memory management:**
- You can't keep buckets for every key forever — millions of keys will eat your RAM
- You need eviction: remove keys that haven't been accessed recently
- LRU (least recently used) is the standard choice
- Track a `lastAccess` timestamp on each bucket, run a background sweep to evict stale ones
- Set a max memory or max key count threshold

**Files:**
- `internal/storage/store.go`
- `internal/storage/eviction.go`

---

## Phase 3: HTTP Server (Make It Callable)

**Goal:** Expose the rate limiter over the network so other services can ask "is this request allowed?"

**What you need to figure out:**
- A single endpoint: `POST /check` with a JSON body like `{"key": "user-123", "tier": "free", "cost": 1}`
- Response: `{"allowed": true, "remaining": 7, "reset_at": "2024-01-01T00:00:01Z"}`
- Wire it up: parse request → look up bucket in storage → call `Allow()` → return result

**Keep it simple for now:**
- Use `net/http` from the standard library, no frameworks
- The handler function should be ~20 lines — it's just glue between HTTP and your bucket logic
- Configuration (listen address, default rates) comes from a config struct, not hardcoded

**Graceful shutdown:**
- Listen for SIGTERM/SIGINT
- Call `server.Shutdown(ctx)` to drain in-flight requests
- This matters later when nodes join/leave the cluster

**Files:**
- `cmd/node/main.go`
- `internal/config/config.go`

---

## Phase 4: Tiered Fairness

**Goal:** Different client tiers get different rate limits, and under contention, allocation is proportional to weight.

**What you need to figure out:**
- Each tier has: a base rate, a burst multiplier, and a weight
- Under normal load: each client gets their tier's full rate
- Under contention (total demand > total capacity): distribute available capacity proportionally by weight
- Guarantee: no tier ever gets zero allocation while capacity remains (starvation prevention)

**The algorithm (weighted fair queuing, simplified):**
1. Calculate total demand across all active tiers
2. If total demand <= total capacity: everyone gets what they want, done
3. If total demand > total capacity: each tier gets `(tier_weight / total_weight) * total_capacity`
4. Floor: minimum allocation = 1 req/s regardless of weight (starvation prevention)

**Where this plugs in:**
- When creating a bucket for a key, the bucket's rate/capacity comes from the fairness allocator, not directly from config
- The allocator recalculates periodically as demand changes

**Files:**
- `internal/fairness/allocator.go`
- `internal/fairness/weight.go`

---

## Phase 5: Gossip Protocol (Distributed State)

**Goal:** Multiple nodes share information about how much of each key's budget has been consumed, so the global limit is enforced across the cluster.

**What you need to figure out:**

### The Problem
Node A allows 5 requests for key "user-123". Node B allows 5 requests for the same key. If the limit is 10, that's fine. But neither node knows about the other's count without communication.

### CRDT Counters (Conflict-free Replicated Data Type)
- Each node maintains a vector: `map[nodeID]int64` per key — how many requests each node has allowed
- To get the total count: sum all entries in the vector
- To merge two vectors: take the max of each entry (this is the CRDT merge rule — it's idempotent and commutative, so order doesn't matter)
- This means gossip messages can arrive late, duplicated, or out of order, and the result is still correct

### Transport
- UDP is fine for gossip — it's fast and you don't need reliability (gossip is redundant by design)
- Every N milliseconds (e.g., 100ms), pick a random peer, serialize your state, send it
- When you receive state from a peer, merge it into your local state

### What to gossip
- You can't send every key's full vector every time — too much data
- Send a digest: for each key, send `(key, total_count, your_node_count)`
- If a peer's total differs significantly from yours, do a full exchange for that key

### Peer management
- Start simple: static list of peers from config
- Track liveness: if a peer hasn't sent gossip in 3 intervals, mark it as suspect
- After 10 intervals with no contact, mark it dead and stop sending to it

**Files:**
- `internal/gossip/peer.go`
- `internal/gossip/state.go`
- `internal/gossip/transport.go`

---

## Phase 6: Reconciliation (Bounding the Error)

**Goal:** Ensure that even with stale gossip data, the system never allows more than 110% of the configured rate.

**What you need to figure out:**

### The drift problem
- Gossip is eventually consistent — there's always a window where your local view is stale
- During that window, each node might over-allow because it doesn't know about requests other nodes approved

### The reconciliation loop
- A background goroutine running every N ms (e.g., 500ms)
- For each key: compare your local count against the merged gossip state
- If the cluster-wide total is approaching the limit, reduce your local bucket's available tokens preemptively
- Formula: `local_allowed = max_rate - estimated_others_consumption - safety_margin`

### Bounded overshoot guarantee
- The worst case: all N nodes allow requests simultaneously before any gossip arrives
- Maximum overshoot = `rate_per_node * gossip_interval * N`
- To keep this under 110%: `rate_per_node = total_rate / N * 0.9` (reserve 10% as buffer)
- The reconciler corrects once gossip catches up, tightening the local budget

### Partition behavior
- If a node loses contact with all peers: it must decide locally
- Conservative strategy: reduce local rate to `total_rate / N` (assume others are still serving)
- This means during a partition, total cluster throughput drops, but never exceeds the limit

**Files:**
- `internal/reconciler/loop.go`

---

## Phase 7: gRPC (Production Transport)

**Goal:** Replace or supplement the HTTP endpoint with gRPC for better performance and type safety.

**What you need to figure out:**
- Define the service in `api/proto/ratelimiter.proto`
- Two RPCs: `CheckRate` and `GetStatus`
- Generate Go code with `protoc`
- The gRPC handler calls the same bucket/storage logic as the HTTP handler

**Why gRPC over HTTP for this:**
- Binary protocol — smaller payloads, faster serialization
- Streaming — could be useful for real-time rate status updates
- Strong typing — the proto file is the contract
- HTTP/2 multiplexing — many requests over one connection

**Files:**
- `api/proto/ratelimiter.proto`
- Generated code goes in `pkg/protocol/` or wherever `protoc` outputs

---

## Phase 8: Chaos Testing

**Goal:** Prove your invariants hold under adversarial conditions.

**Test scenarios:**
1. **Clock skew:** Two nodes disagree on the current time by 500ms. Does the bucket still refill correctly?
2. **Network partition:** Node A can't reach B or C for 5 seconds. Does total allowed stay within 110%?
3. **Node crash:** Node B dies mid-request. Do A and C detect it and redistribute?
4. **Hot key flood:** 100k req/s on one key. Does it degrade other keys?
5. **Memory pressure:** 10 million unique keys. Does eviction keep memory bounded?

**How to test:**
- Inject a `Clock` interface so you can simulate time
- Inject a `Transport` interface so you can drop/delay messages
- Run multiple nodes in-process (no actual network) for fast, deterministic tests
- Property-based testing: generate random request sequences, assert invariants always hold

**Files:**
- `test/integration_test.go`
- `test/chaos_test.go`

---

## Summary: Build Order

| Phase | What | Validates |
|-------|------|-----------|
| 1 | Token bucket | Core algorithm correctness |
| 2 | Key storage | Multi-client, memory management |
| 3 | HTTP server | End-to-end single-node works |
| 4 | Fairness | Weighted allocation logic |
| 5 | Gossip | Distributed state sharing |
| 6 | Reconciler | Bounded overshoot guarantee |
| 7 | gRPC | Production-grade transport |
| 8 | Chaos tests | System holds under failure |

Each phase builds on the previous one. You can stop after phase 3 and have a working single-node rate limiter. Phases 4-8 make it distributed and production-ready.

---

## Questions to Ask Yourself Before Each Phase

- What are the inputs and outputs of this component?
- What invariants must always hold?
- What happens when this component fails?
- How do I test this in isolation?
- What's the simplest thing that could work?
