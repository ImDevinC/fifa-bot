# fifa-bot
This repository contains the code to run a bot designed to watch different FIFA matches and post the contents to a Slack channel.

## Usage
With a recent refactor, this is now a single binary that handles both matches and events. You will need to provide a redis server for persistence between reboots, and a slack webhook URL to send the messages to. But here's the general workflow:

1. On startup, redis is queried for any existing matches and events
1. Every X seconds (default 60), new matches are checked against the FIFA API and then events for all active matches are retrieved
1. Each event for each map is looped over until the last processed event is identified, and then new events are identified and saved to redis
1. The new events are then formatted into friendly text and send to Slack


## Setup
The following variables are needed to run:
- `SLACK_WEBHOOK_URL` - The Slack URL to send messages to
- `REDIS_ADDESS` - The address of the redis server (including port number)
- `REDIS_DB` - The database index of redis to use
- `REDIS_PASSWORD` - If authentication is required, the redis password

The following variables are all optional:
- `SLEEP_TIME_SECONDS` - How long to sleep between each loop of match events (default: `60` seconds)
- `COMPETITION_ID` - If provided, only matches from the specified competition are provided (TODO: Find a way to easily present competition ID's)
- `LOG_LEVEL` - The log level to use for the app
- `ENABLE_PROFILING` - Will enable pprof profiling endpoint
- `PROFILING_PORT` - If pprof is enabled, sets the port it listens on (defaults to `8080`)
