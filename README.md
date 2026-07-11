[![CI](https://github.com/peymanahmadi/payment-risk-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/peymanahmadi/payment-risk-guard/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

# Payment Risk Guard

A small, production-shaped Go service that detects anomalous payment
transactions in near real time. It ingests transaction events from Kafka
(and/or a synchronous REST endpoint), scores each one against a pluggable
set of risk rules, and raises persisted alerts for anything that crosses a
configurable threshold.

It's a deliberately focused portfolio project built around the stack and
concerns of a payments risk platform: Go, PostgreSQL, Kafka, REST, Docker,
clean architecture, and a real test suite — rather than a toy CRUD app.

## What it detects

Three independent, composable rules ship out of the box:

| Rule | Signal |
|---|---|
| **Velocity** | Too many transactions from one account within a rolling time window (account takeover, bonus abuse, bot activity) |
| **Amount spike** | A transaction that's an unusual multiple of the account's recent average (sudden large withdrawal after an account compromise) |
| **Geo mismatch** | "Impossible travel" — the account's country changes faster than a real person could travel |

Each rule returns an independent score; the engine combines them into a
single verdict and reasons list. New rules are added by implementing one
interface (`risk.Rule`) — nothing else in the system needs to change.

## Architecture

```
cmd/
  service/       HTTP API + Kafka consumer, sharing one use case
  simulator/     publishes realistic + anomalous events to Kafka for demos

internal/
  domain/        entities (Transaction, Alert) + repository interfaces (ports)
  risk/          the rule engine — pure logic, zero external dependencies
  usecase/       ProcessTransaction: the single place scoring + persistence happen
  adapter/
    httpapi/     REST handlers (net/http, Go 1.23 routing, no framework)
    kafkaconsumer/  consumes payments.transactions
    kafkaproducer/  used by the simulator
    postgres/    repository implementations + embedded SQL migrations
  config/        environment-based configuration
```

This follows a fairly standard clean/hexagonal architecture:

- **`domain`** defines *what* a transaction and an alert are, and *what*
  persistence looks like (interfaces only — no SQL, no Kafka).
- **`risk`** is pure business logic: given a transaction and read access to
  history, is this risky? It has no knowledge of HTTP, Kafka, or Postgres,
  which is what makes it trivially unit-testable (see `engine_test.go`,
  no database or broker required).
- **`usecase`** is the only place that wires rules to persistence — both the
  HTTP handler and the Kafka consumer call the exact same
  `ProcessTransaction.Execute`, so behaviour never diverges between the
  sync and async entry points.
- **`adapter`** implements the ports for real infrastructure. Swapping
  Postgres for something else, or REST for gRPC, only touches this layer.

One deliberate ordering detail: a transaction is *scored before it's
persisted*. History-based rules (velocity, average amount) read prior
transactions from the repository, so scoring first guarantees a
transaction never counts against its own history. This came up as a real
bug while building this out — see the commit history / `process_transaction.go`.

## Running it

### Everything, via Docker Compose

```bash
docker compose up --build
```

This starts Postgres, a single-node Kafka broker (KRaft mode — no
ZooKeeper), the service (HTTP on `:8080` +
Kafka consumer), and a simulator that publishes a mix of normal and
anomalous transactions every 500ms. Watch the alerts appear:

```bash
curl http://localhost:8080/api/v1/alerts | jq
```

### Locally

```bash
# requires a running Postgres + Kafka, e.g. via `docker compose up postgres kafka`
make run-service
make run-simulator
```

### Tests only (no infrastructure required)

```bash
make test
```

## REST API

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/transactions` | Score a transaction synchronously, get the verdict back |
| `GET` | `/api/v1/alerts` | List alerts (`account_id`, `severity`, `limit`, `offset`) |
| `GET` | `/api/v1/alerts/{id}` | Fetch a single alert |
| `GET` | `/api/v1/accounts/{id}/transactions` | List an account's transaction history |
| `GET` | `/healthz` | Liveness check |

Example:

```bash
curl -X POST http://localhost:8080/api/v1/transactions \
  -H 'Content-Type: application/json' \
  -d '{"account_id":"acc-01","amount":9000,"currency":"USD","country":"US","payment_type":"withdrawal"}'
```

```json
{
  "transaction": { "id": "...", "account_id": "acc-01", "amount": 9000, ... },
  "risk_score": 78,
  "flagged": true,
  "reasons": ["[amount_spike] amount 9000.00 is 90.0x the account's 720h0m0s average (100.00)"],
  "alert": { "id": "...", "severity": "high", ... }
}
```

## Design notes / trade-offs

- **No ORM.** `database/sql` + `lib/pq` directly. At this scale an ORM adds
  indirection without much payoff, and hand-written SQL is easier to reason
  about for the kind of aggregate queries (`AVG(amount) WHERE ...`) the risk
  rules need.
- **Hand-rolled migrations.** A ~50-line embedded runner rather than pulling
  in a full migration framework — enough for this project's needs, and one
  less dependency.
- **`kafka-go` over `sarama` / confluent-kafka-go.** Pure Go, no cgo, simple
  reader/writer API; good fit for a service this size. Pinned to the latest
  release (v0.4.50).
- **Kafka runs in KRaft mode, no ZooKeeper.** Kafka 4.x (bundled in
  Confluent Platform 8.x, used here via `confluentinc/cp-kafka:8.2.2`)
  dropped ZooKeeper support entirely, so the single broker in
  `docker-compose.yml` acts as its own controller — accurate to how you'd
  run Kafka today, and one less moving part for a local demo.
- **Engine combines rule scores by taking the max, not a weighted sum.**
  Simpler to reason about and tune per-rule in isolation. A weighted-sum or
  ensemble-scoring mode would be a natural next step (see below).
- **A malformed Kafka message is logged and skipped, not dead-lettered.** A
  real production system would route it to a dead-letter topic; that's
  called out explicitly in `kafkaconsumer/consumer.go` as a known
  simplification.

## What I'd add next

- A dead-letter topic for unprocessable Kafka messages
- Weighted/ensemble scoring instead of "max wins"
- An admin endpoint to acknowledge/resolve alerts (currently write-only)
- Integration tests against real Postgres/Kafka via `testcontainers-go`
  (unit tests currently use in-memory fakes for the rule engine and use
  case, which need no infrastructure — see `internal/risk/engine_test.go`)
- Structured metrics (Prometheus) alongside the existing structured logs

## License

MIT — do whatever you like with it.
