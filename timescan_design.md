# Project Specification & Architecture Design: `timescan` 
**A Go-Native Time-Series Analytics, Anomaly Detection & Vector Pattern Engine**

## 1. The Problem & The Opportunity
Currently, Go is the undisputed king of infrastructure and observability (Prometheus, Grafana, InfluxDB). However, when developers need to perform advanced time-series analytics (e.g., detecting anomalies in metric streams, predicting traffic spikes, or decomposing seasonality), they are often forced to rely on external HTTP calls to Python microservices running `scikit-learn` or `statsmodels`.

**The Goal:** Build `timescan`, a lightweight, pure Go toolkit for time-series analysis and anomaly detection. It must be designed as a "Plug & Play" library that can be embedded directly into monitoring pipelines, custom Prometheus collectors, and stream processors without requiring heavy infrastructure overhead.

---

## 2. Go Architecture & Design Patterns
To ensure scalability and maintainability, `timescan` must strictly adhere to modern Go idiomatic patterns:

- **Ports and Adapters (Hexagonal Architecture):** Core domain logic (anomaly detection, math) must be completely decoupled from external integrations. Ingestion sources (Prometheus, OpenTelemetry) and egress sinks (Qdrant, AlertManager) are implemented as interchangeable Adapters around the Domain Ports.
- **Concurrency Model (Worker Pools & Fan-in/Fan-out):** High-throughput metric processing will utilize Go channels and worker pools. A `Dispatcher` component will distribute metric streams across multiple Goroutines without locking bottlenecks.
- **Memory Management (Zero-Allocation Paths):** Time-series ingestion causes heavy GC (Garbage Collection) pressure. The architecture must utilize `sync.Pool` for object reuse (e.g., recycling `DataPoint` slices) and pre-allocated Ring Buffers to achieve zero-allocations in the hot path.
- **Interface Segregation:** Small, focused interfaces (e.g., `io.Reader`-style) for metrics ingestion and vector matching, enabling easy mocking during unit testing.

---

## 3. Core Software Components & Subsystems
The system is divided into functional subsystems:

1. **Ingestion Layer (Adapters):** 
   - Translates external data formats (Prometheus remote-read, OpenTelemetry OTLP, JSON streams) into the internal `timescan.DataPoint` format.
2. **Pipeline Coordinator (The Engine):** 
   - Manages the lifecycle of a stream. Routes incoming points to the correct `RingBufferWindow`, triggers the `Decomposition` engine, and passes the result to the `AnomalyDetector`.
3. **Analytics & Math Domain (Core):** 
   - Pure, stateless mathematical functions for STL decomposition, Moving Averages, and statistical thresholds (Z-Score, MAD, EWMA).
4. **Vector Storage & Pattern Matching Layer:** 
   - Converts time-series shapes into embeddings (using PAA/FFT) and interfaces with the Vector DB to find historical similarities.
5. **Event Dispatcher / Egress (Sinks):** 
   - When an anomaly is detected, this component formats the event and sends it to external sinks (Webhooks, Prometheus Exporter `/metrics` endpoint, Alertmanager).

---

## 4. Supported Vector Databases & Search Engines
`timescan` supports two tiers of Vector DBs via a unified `vector.Store` interface, enabling usage in both edge devices and large clusters.

### Tier 1: Embedded / Native Go Stores (Zero external infrastructure)
- **`Bbolt` / `BadgerDB` + Native HNSW:** Pure Go embedded vector index using Hierarchical Navigable Small World (HNSW) graphs persisted on local key-value stores.
- **`SQLite` (via `sqlite-vec` extension):** Embedded relation-vector storage.

### Tier 2: Enterprise / External Vector Databases
- **Qdrant (Primary Target):** High-performance gRPC client support, leveraging Qdrant's advanced payload filtering (e.g., filter by `tenant_id`).
- **Milvus:** Support for massive-scale enterprise observability clusters.
- **Pgvector (PostgreSQL):** Native integration for setups leveraging PostgreSQL.

---

## 5. Package Structure & Directory Layout
```text
timescan/
├── internal/         # Private memory management (sync.Pool wrappers)
├── timeseries/       # Domain primitives, Ring buffers, Welford's statistics
├── pipeline/         # Stream coordinator, Fan-in/Fan-out goroutine workers
├── decomposition/    # STL, Additive/Multiplicative Moving Average
├── anomaly/          # Z-Score, EWMA, MAD, Isolation Forest (pure Go)
├── vector/           # Embeddings & Vector matching domain
│   ├── driver/       # Infrastructure Adapters: qdrant, pgvector, bbolt
│   └── embedding/    # Shape-to-vector encoders (SAX, PAA, FFT)
├── ingestion/        # Infrastructure Adapters: prometheus, otel
├── egress/           # Alert dispatchers, webhooks
└── examples/         # Runnable quickstarts and benchmarks
