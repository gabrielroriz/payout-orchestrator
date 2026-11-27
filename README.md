# Payout Orchestrator

[![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![Apache Kafka](https://img.shields.io/badge/Apache_Kafka-231F20?style=flat&logo=apachekafka&logoColor=white)](https://kafka.apache.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![CEL](https://img.shields.io/badge/CEL-555555?style=flat)](https://cel.dev/)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-6BA539?style=flat&logo=openapiinitiative&logoColor=white)](https://www.openapis.org/)
[![gRPC](https://img.shields.io/badge/gRPC-244C5A?style=flat)](https://grpc.io/)
[![Protocol Buffers](https://img.shields.io/badge/Protocol_Buffers-555555?style=flat)](https://protobuf.dev/)
[![Buf](https://img.shields.io/badge/Buf-555555?style=flat)](https://buf.build/)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-425CC7?style=flat&logo=opentelemetry&logoColor=white)](https://opentelemetry.io/)
[![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=flat&logo=prometheus&logoColor=white)](https://prometheus.io/)
[![Grafana](https://img.shields.io/badge/Grafana-F46800?style=flat&logo=grafana&logoColor=white)](https://grafana.com/)
[![Testcontainers](https://img.shields.io/badge/Testcontainers-555555?style=flat)](https://testcontainers.com/)
[![k6](https://img.shields.io/badge/k6-7D64FF?style=flat&logo=k6&logoColor=white)](https://k6.io/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white)](https://www.docker.com/)
[![Terraform](https://img.shields.io/badge/Terraform-844FBA?style=flat&logo=terraform&logoColor=white)](https://developer.hashicorp.com/terraform)
[![Amazon MSK](https://img.shields.io/badge/Amazon_MSK-8C4FFF?style=flat)](https://aws.amazon.com/msk/)
[![Amazon RDS](https://img.shields.io/badge/Amazon_RDS-4053D6?style=flat)](https://aws.amazon.com/rds/)
[![Amazon EKS](https://img.shields.io/badge/Amazon_EKS-FF9900?style=flat)](https://aws.amazon.com/eks/)

Pay creators for their results, exactly once, across many payout providers.

Payout Orchestrator is a payouts backend for creator and influencer platforms. Platforms send what their creators achieve, such as a published post or a post's view count. Declarative rules turn those events into earnings, a double-entry ledger records every cent, and payouts are routed across several providers with failover that never pays anyone twice.

> **Status:** design phase, no code yet. The [documentation](#documentation) is a first draft, open for review.

## Tech stack

| Concern | Choice |
| --- | --- |
| Language | Go |
| Streaming | Apache Kafka (franz-go), Amazon MSK on AWS |
| Database | PostgreSQL (pgx, sqlc, goose) |
| Rules | CEL (cel-go) |
| APIs | REST + OpenAPI (public), ConnectRPC/gRPC (internal), Protobuf + buf |
| Observability | OpenTelemetry, Prometheus, Grafana |
| Infrastructure | Docker Compose (local), Terraform (AWS) |

## Roadmap

| Milestone | Exit criteria |
| --- | --- |
| **M0** Foundations | `make up` starts the stack, and a test event flows end to end in a single trace |
| **M1** Money core | Fund the account → pay a payee → receive a signed webhook, with the ledger balanced. An unfunded payout is blocked, never sent |
| **M2** Orchestration | Kill a provider in the middle of a 100k batch: 0 duplicates, 0 lost |
| **M3** Performance engine | A post's view snapshots, with duplicates, out-of-order snapshots and a final removal → exactly $2 per 1,000 views accrued, capped at $500, then clawed back |
| **M4** Production-grade | Latency and throughput targets met under load, a clean reconciliation report, Terraform for AWS |

## Getting started

There is nothing to run yet. From M0, the full stack (Kafka, Postgres, Grafana and every service) will start locally with one command, `make up`, on Docker Compose.

## Documentation

- [Business definition](docs/00-business.md): why it exists. The customer, the problem, the offer, how a tenant uses it, the business model, and possible future segments.
- [Project scope](docs/01-scope.md): what it must do. Goals and non-goals, and requirements.
- [Design](docs/02-design.md): how it does it. Architecture, Kafka topics, rules engine, payout state machine, key decisions and capacity.

## License

To be defined.
