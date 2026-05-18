distributed-rate-limiter/
├── cmd/
│   └── node/
│       └── main.go
│
├── internal/
│   ├── bucket/
│   │   ├── token.go
│   │   └── token_test.go
│   │
│   ├── gossip/
│   │   ├── peer.go
│   │   ├── state.go
│   │   └── transport.go
│   │
│   ├── fairness/
│   │   ├── allocator.go
│   │   └── weight.go
│   │
│   ├── reconciler/
│   │   └── loop.go
│   │
│   ├── storage/
│   │   ├── store.go
│   │   └── eviction.go
│   │
│   └── config/
│       └── config.go
│
├── api/
│   └── proto/
│       └── ratelimiter.proto
│
├── pkg/
│   └── client/
│       └── client.go
│
├── configs/
│   └── default.yaml
│
├── test/
│   ├── integration_test.go
│   └── chaos_test.go
│
├── scripts/
│   └── run-cluster.sh
│
├── go.mod
├── Makefile
└── README.md
