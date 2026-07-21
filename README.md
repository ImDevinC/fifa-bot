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
- **Sentry integration**: Automatically captures unknown event types to Sentry for tracking

## Configuration

The bot is configured via a YAML file (default: `./config.yaml`) with environment variable overrides. Environment variables take precedence over values in the config file.

### Config File (`config.yaml`)

```yaml
slack_webhook_url: "https://hooks.slack.com/services/..."
competition_id: "17"              # Optional: filter by competition
sleep_time_seconds: 60            # Polling interval (default: 60)
redis:
  address: "localhost:6379"       # Required
  password: ""                    # Optional
  database: 0                     # Required
log_level: "WARN"                 # DEBUG, INFO, WARN, ERROR (default: WARN)
enable_profiling: false           # Enable pprof endpoint (default: false)
profiling_port: 8080              # pprof server port (default: 8080)
sentry_dsn: "https://..."        # Optional: Sentry DSN for tracking unknown events
```

### Environment Variable Overrides

Any config value can be overridden by its corresponding environment variable:

| Variable | Overrides | Required |
|---|---|---|
| `SLACK_WEBHOOK_URL` | `slack_webhook_url` | Yes |
| `REDIS_ADDRESS` | `redis.address` | Yes |
| `REDIS_DB` | `redis.database` | Yes |
| `COMPETITION_ID` | `competition_id` | No |
| `SLEEP_TIME_SECONDS` | `sleep_time_seconds` | No |
| `REDIS_PASSWORD` | `redis.password` | No |
| `LOG_LEVEL` | `log_level` | No |
| `ENABLE_PROFILING` | `enable_profiling` | No |
| `PROFILING_PORT` | `profiling_port` | No |
| `SENTRY_DSN` | `sentry_dsn` | No |

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

## Unknown Event Tracking with Sentry

When the bot encounters a match event type it doesn't recognize, instead of sending a generic Slack message, it creates a new issue in Sentry for tracking. This helps identify new or undocumented FIFA API event types.

### Setup

1. Create a Sentry account and project at https://sentry.io
2. Copy your project's DSN (found in Project Settings -> Client Keys)
3. Add `sentry_dsn` to your config file or set the `SENTRY_DSN` environment variable

### What Gets Captured

Each unknown event creates a Sentry issue containing:

- **Tags**: Event type code, match ID, stage ID, season ID, competition ID, and team abbreviations — for easy filtering in Sentry
- **Extra data**: The full JSON of the event payload and the complete match info, accessible in the Sentry issue details

### Example Sentry Issue

When enabled, an unknown event will create an issue in Sentry with a message like:
```
Unknown event type: 42 in match 400021528 (ARG vs BRA)
```

You can then use the tags and extra data to research the new event type and add support for it.

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
- **Configuration**: `github.com/spf13/viper` - YAML + environment config
- **Error Tracking**: `github.com/getsentry/sentry-go` - Error monitoring

## License

This project is licensed under the MIT License.
