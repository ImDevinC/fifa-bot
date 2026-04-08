# FIFA Bot
>[!NOTE]
>AI was only used in this project to generate this readme. It was not used for any of the actual code.

A real-time FIFA match monitoring bot that tracks live matches and sends notifications to Slack when events occur.

## Overview

FIFA Bot monitors live FIFA matches and sends real-time notifications to a Slack channel when match events happen, including:

- ⚽ Goals (regular, penalty, own goals)
- 🟨 Yellow cards
- 🟥 Red cards
- 🔄 Substitutions
- 🕐 Match start/end and half-time
- ❌ Penalty misses

## Features

- **Real-time monitoring**: Continuously polls FIFA API for live match events
- **Slack integration**: Sends formatted notifications with team flags and emojis
- **Redis persistence**: Stores match state to handle restarts and avoid duplicate notifications
- **Competition filtering**: Optional filtering by specific competition ID
- **Concurrent processing**: Handles multiple matches simultaneously
- **Docker support**: Containerized deployment ready
- **Profiling support**: Optional pprof endpoint for performance monitoring
- **Health checks**: Built-in `/healthz` endpoint for Kubernetes liveness/readiness probes

## Configuration

The bot is configured via environment variables:

### Required
- `SLACK_WEBHOOK_URL`: Slack webhook URL for sending notifications
- `REDIS_ADDRESS`: Redis server address
- `REDIS_DB`: Redis database number

### Optional
- `COMPETITION_ID`: Filter matches by specific competition (empty = all competitions)
- `SLEEP_TIME_SECONDS`: Polling interval in seconds (default: 60)
- `REDIS_PASSWORD`: Redis password if required
- `LOG_LEVEL`: Logging level - DEBUG, INFO, WARN, ERROR (default: WARN)
- `ENABLE_PROFILING`: Enable pprof endpoint (default: false)
- `PROFILING_PORT`: pprof server port (default: 8080)
- `ENABLE_HEALTH_CHECK`: Enable health check endpoint (default: true)
- `HEALTH_CHECK_PORT`: Health check server port (default: 8081)

## Health Checks

The bot includes a built-in health check endpoint for use with Kubernetes liveness and readiness probes. This endpoint validates that the application is running and can communicate with Redis.

### Endpoint

- **URL**: `GET /healthz`
- **Port**: Configurable via `HEALTH_CHECK_PORT` (default: 8081)

### Responses

| Status | Body | Description |
|--------|------|-------------|
| 200 OK | `ok` | Application is healthy and Redis is reachable |
| 503 Service Unavailable | `unhealthy: redis connection failed` | Redis connection failed |

### Usage Example

```bash
# Check health status
curl http://localhost:8081/healthz
```

### Kubernetes Configuration

Configure liveness and readiness probes in your pod specification:

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: fifa-bot
      image: ghcr.io/imdevinc/fifa-bot:latest
      ports:
        - containerPort: 8081
          name: health
      livenessProbe:
        httpGet:
          path: /healthz
          port: health
        initialDelaySeconds: 10
        periodSeconds: 30
        timeoutSeconds: 5
        failureThreshold: 3
      readinessProbe:
        httpGet:
          path: /healthz
          port: health
        initialDelaySeconds: 5
        periodSeconds: 10
        timeoutSeconds: 5
        failureThreshold: 3
```

### Disabling Health Checks

If you don't need the health check endpoint (e.g., running outside of Kubernetes), you can disable it:

```bash
export ENABLE_HEALTH_CHECK=false
```

## Installation & Usage

### Docker (Recommended)

```bash
docker run -e SLACK_WEBHOOK_URL=your_webhook_url \
           -e REDIS_ADDRESS=redis:6379 \
           -e REDIS_DB=0 \
           ghcr.io/imdevinc/fifa-bot:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  fifa-bot:
    image: ghcr.io/imdevinc/fifa-bot:latest
    environment:
      - SLACK_WEBHOOK_URL=your_webhook_url
      - REDIS_ADDRESS=redis:6379
      - REDIS_DB=0
      - COMPETITION_ID=17  # Optional: World Cup
      - SLEEP_TIME_SECONDS=30
    depends_on:
      - redis
  
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
```

### Local Development

1. **Prerequisites**:
   - Go 1.23+
   - Redis server

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set environment variables**:
   ```bash
   export SLACK_WEBHOOK_URL=your_webhook_url
   export REDIS_ADDRESS=localhost:6379
   export REDIS_DB=0
   ```

4. **Run the bot**:
   ```bash
   go run cmd/server.go
   ```

## Architecture

- **`cmd/server.go`**: Application entry point and configuration
- **`pkg/app/`**: Core application logic and match monitoring
- **`pkg/fifa/`**: FIFA API integration and event processing
- **`pkg/database/`**: Redis database operations
- **`pkg/models/`**: Data structures for matches and Slack messages

## Slack Webhook Setup

1. Create a Slack app at https://api.slack.com/apps
2. Add Incoming Webhooks feature
3. Create a webhook for your desired channel
4. Use the webhook URL as `SLACK_WEBHOOK_URL`

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o fifa-bot cmd/server.go
```

### Docker Build
```bash
docker build -t fifa-bot .
```

## Dependencies

- **FIFA API Client**: `github.com/imdevinc/go-fifa` - FIFA API integration
- **Redis**: `github.com/redis/go-redis/v9` - Match state persistence  
- **Logging**: `github.com/sirupsen/logrus` - Structured logging
- **Configuration**: `github.com/kelseyhightower/envconfig` - Environment-based config
- **Error Tracking**: `github.com/getsentry/sentry-go` - Error monitoring

## License

This project is licensed under the MIT License.
