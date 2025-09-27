# Night Fury Radar

Night Fury Radar is a lightweight system monitor. It ships two binaries:

- **Astrid** continuously samples system metrics (CPU, memory, etc.) using `gopsutil` and persists them in PostgreSQL.
- **Hiccup** streams stored metrics to WebSocket clients, replaying recent history and pushing new samples as they arrive.

## System Design

## Overview

Astrid collects the metrics and saves it to DB. Hiccup serves the metrics to clients via websocket.

### Astrid
- `MetricCollector` is an interface that is implemented by different metric collectors whose job is to collect their respective metric like CPU, Memory etc.
- `MetricProcessor` orchestrates all the various `MetricCollector`s.
- The constructor initiliases three things:
    1. `metric` channel. All collectors write to this channel.
    2. `Storage` which stores and reads from DB.
    3. `Context` with a cancellation signal
- Start the metric collection process by calling `start()` which starts individual `MetricCollector`s.
- This starts an infinite loop where a goroutine waits until
    1. Either a metric is received from some collector in the `metric` channel, or
    2. The processor is stopped via `cancel` func of the contex
- Each `MetricCollector` starts its goroutine, which starts a new ticker and on every tick it gets the corresponding metric and write it to the `metric` channel.
- When a metric is received in the channel it is written to the DB.
- **Shutdown** is gracefully handled when an `os.Interrupt` signal is received.
    1. On shutdown, `stop()` method is called on `MetricProcessor`
    2. This sends cancel signal up the processor context, and then waits until all goroutines are finished.
    3. And at last closes the channel and db connections.

### Hiccup
- Hiccup sets up a new Server, which has two jobs
    1. Receive metrics from Astrid, and
    2. Broadcast it to clients
- An endpoint is exposed for each metric type which handles connection with the client and adds it to the pool of clients.
- Receiver reads the metric from DB and writes it to the metric channel.
- Broadcaster reads from the metric channel and writes it to each client that requested for that metric type.

### DB Layer
- PostgreSQL pub/sub model is used by Hiccup to read metrics in real-time.
- Whenever Astrid writes to DB, a notification is sent to which Hiccup subscribes, It consumes the event and processes it.

### Benefits of this architecture
- Durable pipeline: Astrid persists every sample to Postgres, so restarts, Hiccup downtime, or client drop-offs do not lose data.
- Low-latency delivery: Postgres LISTEN/NOTIFY paired with WebSockets keeps the stream near real-time without custom message brokers.
- Pluggable collectors: The `MetricCollector` interface makes it straightforward to add new system probes as needs evolve.
- Simple operations: Two binaries with `.env` configuration and standard dependencies keep deployment friction low.

## Prerequisites
- Go 1.24 or newer
- PostgreSQL 14+ (running locally or reachable through your network)
- A `.env` file that can be loaded by [`godotenv`](https://github.com/joho/godotenv)

## Environment configuration
Astrid and Hiccup both load environment variables from `.env` at startup. At minimum, define the database user and name that Postgres should use:

```dotenv
DB_USER=your_postgres_user
DB_NAME=night_fury_radar
```
## Database setup
Create the database (if it does not exist), the table Astrid writes into, and a trigger that broadcasts new rows so Hiccup can relay them to WebSocket clients:

## Running locally
1. Ensure Postgres is running and `.env` is populated.
2. Start Astrid (collector):
   ```bash
   go run ./cmd/astrid
   ```
   Astrid will sample CPU and memory usage once per second and insert the readings into `system_metrics`.
3. Start Hiccup (WebSocket server) in another shell:
   ```bash
   go run ./cmd/hiccup
   ```
   The server listens on `:8080` and exposes WebSocket endpoints at `/cpu` and `/memory`.

## Consuming the WebSocket stream
Any WebSocket client can subscribe to a metric type. Example using [`websocat`](https://github.com/vi/websocat):

```bash
websocat ws://localhost:8080/cpu
```

Hiccup will immediately replay the most recent 100 samples for the requested metric type, then push new samples as Astrid records them. Messages are JSON objects with the form:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "type": "cpu",
  "values": { "percent": 12.5 }
}
```