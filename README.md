# SlashVibe

A simple Go service that subscribes to Slack slash commands via Redis and performs operations.

## Features

- Subscribes to a Redis channel to receive Slack slash command payloads
- Processes `/new-repo` command to display a modal for creating new repositories
- Configurable via environment variables
- Docker and Docker Compose support with scratch runtime for minimal image size

## Prerequisites

- Go 1.24 or later
- Redis server
- Slack Bot Token with appropriate permissions (including `commands` and `views:write`)

## Configuration

The service can be configured via environment variables or a `.secret` file:

### Environment Variables

- `REDIS_ADDR` - Redis server address (default: `localhost:6379`)
- `REDIS_CHANNEL` - Redis channel to subscribe to (default: `slack-commands`)
- `SLACK_BOT_TOKEN` - Slack bot token (required)

### .secret File

Alternatively, you can create a `.secret` file in the root directory containing your Slack bot token:

```bash
cp .secret.example .secret
# Edit .secret and add your Slack bot token
```

## Running Locally

1. Install dependencies:
```bash
go mod download
```

2. Build the service:
```bash
go build -o slashvibe
```

3. Run the service:
```bash
export SLACK_BOT_TOKEN=xoxb-your-token-here
./slashvibe
```

Or use the `.secret` file:
```bash
echo "xoxb-your-token-here" > .secret
./slashvibe
```

## Running with Docker Compose

1. Create a `.secret` file with your Slack bot token or set the `SLACK_BOT_TOKEN` environment variable

2. Start the services:
```bash
docker-compose up --build
```

This will start:
- Redis server on port 6379
- SlashVibe service connected to Redis

## Slash Command Payload Format

The service expects slash command payloads in the following JSON format on the Redis channel:

```json
{
  "token": "<redacted>",
  "team_id": "<redacted>",
  "team_domain": "<redacted>",
  "channel_id": "<redacted>",
  "channel_name": "directmessage",
  "user_id": "<redacted>",
  "user_name": "vibechung",
  "command": "/new-repo",
  "text": "<repo name>",
  "response_url": "https://hooks.slack.com/commands/<redacted>/<redacted>/<redacted>",
  "trigger_id": "<redacted>",
  "api_app_id": "<redacted>"
}
```

## Supported Commands

### `/new-repo`

Opens a modal dialog for creating a new repository with the following fields:
- **Repository Name** (required) - Letters, numbers, hyphens only
- **Repository Description** (optional) - A short description
- **Copilot Issue Prompt** (optional) - Describe what Copilot should generate

## Testing

You can test the service by publishing a message to the Redis channel:

```bash
redis-cli PUBLISH slack-commands '{"token":"test","team_id":"T123","team_domain":"test","channel_id":"C123","channel_name":"general","user_id":"U123","user_name":"testuser","command":"/new-repo","text":"my-repo","response_url":"https://example.com","trigger_id":"123.456.abc","api_app_id":"A123"}'
```

## License

MIT

