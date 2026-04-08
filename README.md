# FIFA Bot

>[!NOTE]
>AI was only used in this project to generate this readme. It was not used for any of the actual code.

A real-time FIFA match monitoring bot that tracks live matches and sends notifications to Slack when events occur.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [How It Works](#how-it-works)
- [Configuration](#configuration)
- [Installation & Usage](#installation--usage)
- [Architecture](#architecture)
- [Slack Webhook Setup](#slack-webhook-setup)
- [Event Types](#event-types)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Dependencies](#dependencies)
- [Contributing](#contributing)
- [License](#license)

## Overview

FIFA Bot monitors live FIFA matches and sends real-time notifications to a Slack channel when match events happen, including:

- ⚽ Goals (regular, penalty, own goals)
- 🟨 Yellow cards
- 🟥 Red cards
- 🔄 Substitutions
- 🕐 Match start/end and half-time
- ❌ Penalty misses
- ⚠️ Penalty awarded

## Features

- **Real-time monitoring**: Continuously polls FIFA API for live match events
- **Slack integration**: Sends formatted notifications with team flags and emojis
- **Redis persistence**: Stores match state to handle restarts and avoid duplicate notifications
- **Competition filtering**: Optional filtering by specific competition ID
- **Concurrent processing**: Handles multiple matches simultaneously using goroutines
- **Docker support**: Containerized deployment ready with distroless base image
- **Profiling support**: Optional pprof endpoint for performance monitoring
- **Graceful state recovery**: Restores match state from Redis on restart

## How It Works

FIFA Bot operates on a continuous polling loop that:

1. **Fetches live matches** from the FIFA API using the [go-fifa](https://github.com/imdevinc/go-fifa) client
2. **Filters matches** by competition ID (if configured)
3. **Stores match metadata** in Redis for state persistence
4. **Monitors events** for each active match concurrently
5. **Detects new events** by comparing against previously seen event IDs
6. **Sends notifications** to Slack for relevant events
7. **Cleans up** completed matches from Redis automatically

### State Persistence

The bot uses Redis to persist match state, which enables:

- **Restart resilience**: On startup, the bot loads all active matches from Redis
- **Duplicate prevention**: Event IDs are tracked to avoid sending duplicate notifications
- **Automatic cleanup**: Completed matches are removed from Redis with 24-hour TTL as a safety net

### Concurrent Event Processing

Multiple matches are processed concurrently using Go's `errgroup` package, ensuring efficient handling during tournaments with many simultaneous games.

## Configuration

The bot is configured via environment variables:

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `SLACK_WEBHOOK_URL` | Slack webhook URL for sending notifications | `https://hooks.slack.com/services/...` |
| `REDIS_ADDRESS` | Redis server address (host:port) | `localhost:6379` |
| `REDIS_DB` | Redis database number | `0` |

### Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `COMPETITION_ID` | Filter matches by specific competition (empty = all) | _(empty)_ |
| `SLEEP_TIME_SECONDS` | Polling interval in seconds | `60` |
| `REDIS_PASSWORD` | Redis password if authentication is required | _(empty)_ |
| `LOG_LEVEL` | Logging level: `DEBUG`, `INFO`, `WARN`, `ERROR` | `WARN` |
| `ENABLE_PROFILING` | Enable pprof endpoint for profiling | `false` |
| `PROFILING_PORT` | Port for pprof server | `8080` |

### Competition IDs

Some common FIFA competition IDs:

| Competition | ID |
|-------------|-----|
| FIFA World Cup | `17` |
| FIFA Women's World Cup | `103` |
| FIFA Club World Cup | `123` |
| FIFA U-20 World Cup | `104` |

> **Note**: Competition IDs may change between seasons. Check the FIFA API for current IDs.

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
      - LOG_LEVEL=INFO
    depends_on:
      - redis
    restart: unless-stopped
  
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped

volumes:
  redis-data:
```

### Kubernetes

Example deployment manifest:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fifa-bot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: fifa-bot
  template:
    metadata:
      labels:
        app: fifa-bot
    spec:
      containers:
      - name: fifa-bot
        image: ghcr.io/imdevinc/fifa-bot:latest
        env:
        - name: SLACK_WEBHOOK_URL
          valueFrom:
            secretKeyRef:
              name: fifa-bot-secrets
              key: slack-webhook-url
        - name: REDIS_ADDRESS
          value: "redis-service:6379"
        - name: REDIS_DB
          value: "0"
        - name: LOG_LEVEL
          value: "INFO"
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
```

### Local Development

1. **Prerequisites**:
   - Go 1.23+
   - Redis server (local or Docker)

2. **Clone the repository**:
   ```bash
   git clone https://github.com/imdevinc/fifa-bot.git
   cd fifa-bot
   ```

3. **Install dependencies**:
   ```bash
   go mod download
   ```

4. **Start Redis** (if not already running):
   ```bash
   docker run -d -p 6379:6379 redis:alpine
   ```

5. **Set environment variables**:
   ```bash
   export SLACK_WEBHOOK_URL=your_webhook_url
   export REDIS_ADDRESS=localhost:6379
   export REDIS_DB=0
   export LOG_LEVEL=DEBUG  # Optional: for verbose output
   ```

6. **Run the bot**:
   ```bash
   go run cmd/server.go
   ```

## Architecture

```
fifa-bot/
├── cmd/
│   └── server.go          # Application entry point, config loading, HTTP servers
├── pkg/
│   ├── app/
│   │   ├── app.go         # Core application logic, main event loop
│   │   └── config.go      # Environment variable configuration
│   ├── database/
│   │   ├── database.go    # Database interface definition
│   │   └── redis.go       # Redis implementation
│   ├── fifa/
│   │   ├── fifa.go        # FIFA API integration, event processing
│   │   └── consts.go      # Event types to skip, country flag mappings
│   ├── helper/
│   │   └── tracing.go     # Tracing utilities
│   └── models/
│       ├── match.go       # Match data structure, Redis serialization
│       └── slackmessage.go # Slack message payload structure
├── Dockerfile             # Multi-stage Docker build (distroless)
├── go.mod                 # Go module definition
└── go.sum                 # Dependency checksums
```

### Component Overview

| Component | Responsibility |
|-----------|---------------|
| **cmd/server.go** | Entry point; initializes config, Redis client, FIFA client; starts HTTP servers for profiling; runs main application loop |
| **pkg/app** | Core business logic; manages match polling loop, event detection, Slack notifications |
| **pkg/database** | Abstracts database operations; currently implements Redis for state persistence |
| **pkg/fifa** | Wraps go-fifa client; processes raw FIFA events into Slack-friendly messages |
| **pkg/models** | Data structures for matches and Slack payloads; handles Redis serialization |

## Slack Webhook Setup

1. Go to [Slack API Apps](https://api.slack.com/apps)
2. Click **Create New App** → **From scratch**
3. Name your app (e.g., "FIFA Bot") and select your workspace
4. In the left sidebar, click **Incoming Webhooks**
5. Toggle **Activate Incoming Webhooks** to On
6. Click **Add New Webhook to Workspace**
7. Select the channel where notifications should be posted
8. Copy the webhook URL and use it as `SLACK_WEBHOOK_URL`

### Example Slack Message

When a goal is scored, the bot sends a message like:

```
45' ⚽ Lionel Messi scores from a brilliant free kick 2 ARG 🇦🇷 : 🇫🇷 FRA 1
```

## Event Types

### Tracked Events (Notifications Sent)

| Event | Emoji | Description |
|-------|-------|-------------|
| Goal | ⚽ | Regular goals |
| Own Goal | ⚽ | Own goals (noted in description) |
| Penalty Goal | ⚽ | Goals from penalty kicks |
| Yellow Card | 🟨 | Yellow card shown |
| Double Yellow | 🟨 | Second yellow card (red card) |
| Red Card | 🟥 | Direct red card |
| Match Start | 🕐 | Match kickoff |
| Half End | 🕧 | End of first half, extra time periods |
| Match End | 🕐 | Full time (includes final score) |
| Penalty Missed | 🚫 | Missed penalty kicks |
| Penalty Awarded | ⚠️ | Penalty awarded (except during shootouts) |

### Skipped Events (No Notification)

The following events are tracked but don't generate notifications:

- Substitutions
- Match paused/resumed
- Goal attempts
- Offsides
- Corner kicks
- Free kick posts
- Blocked shots
- Fouls
- Throw-ins
- Clearances
- Crossbar hits
- Goalie saves
- Dropped balls
- Assists

## Development

### Running Tests

```bash
go test ./...
```

### Running Tests with Coverage

```bash
go test -cover ./...
```

### Building

```bash
go build -o fifa-bot cmd/server.go
```

### Docker Build

```bash
docker build -t fifa-bot .
```

### Multi-architecture Build

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t fifa-bot .
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run staticcheck (install first: go install honnef.co/go/tools/cmd/staticcheck@latest)
staticcheck ./...
```

## Troubleshooting

### Common Issues

#### No notifications are being sent

1. **Check log level**: Set `LOG_LEVEL=DEBUG` to see detailed logs
2. **Verify webhook URL**: Test with a curl command:
   ```bash
   curl -X POST -H 'Content-type: application/json' \
     --data '{"text":"Test message"}' \
     YOUR_SLACK_WEBHOOK_URL
   ```
3. **Check Redis connection**: Ensure Redis is running and accessible
4. **Verify live matches**: The FIFA API only returns currently live matches

#### Duplicate notifications

1. **Check Redis persistence**: Ensure Redis data is persisted across restarts
2. **Verify single instance**: Only run one instance of the bot per Redis database

#### Missing events

1. **Adjust polling interval**: Lower `SLEEP_TIME_SECONDS` for faster updates (minimum recommended: 15)
2. **Check network connectivity**: Ensure the bot can reach the FIFA API

#### Redis connection errors

1. **Verify address format**: Use `host:port` format (e.g., `localhost:6379`)
2. **Check authentication**: Set `REDIS_PASSWORD` if Redis requires auth
3. **Verify network access**: Ensure firewall rules allow the connection

### Debugging Tips

- Enable debug logging: `LOG_LEVEL=DEBUG`
- Enable profiling for performance issues: `ENABLE_PROFILING=true`
- Access pprof at: `http://localhost:8080/debug/pprof/`

## Dependencies

| Package | Purpose |
|---------|---------|
| [github.com/imdevinc/go-fifa](https://github.com/imdevinc/go-fifa) | FIFA API client library |
| [github.com/redis/go-redis/v9](https://github.com/redis/go-redis) | Redis client for state persistence |
| [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) | Structured logging |
| [github.com/kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig) | Environment-based configuration |
| [github.com/getsentry/sentry-go](https://github.com/getsentry/sentry-go) | Error monitoring and tracking |
| [golang.org/x/sync](https://golang.org/x/sync) | Concurrent processing utilities |

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork the repository** and create your branch from `main`
2. **Follow conventional commits** for commit messages:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `refactor:` for code refactoring
3. **Run tests** before submitting: `go test ./...`
4. **Update documentation** if you change configuration or behavior
5. **Submit a pull request** with a clear description of changes

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/fifa-bot.git
cd fifa-bot

# Add upstream remote
git remote add upstream https://github.com/imdevinc/fifa-bot.git

# Create a feature branch
git checkout -b feat/your-feature

# Make changes, commit, and push
git add .
git commit -m "feat: add your feature"
git push origin feat/your-feature
```

## License

This project is licensed under the MIT License.
