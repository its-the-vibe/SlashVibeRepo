# SlashVibeRepo

A simple Go service that subscribes to Slack slash commands and view submissions via Redis and performs operations.

## Features

- Subscribes to Redis channels to receive Slack slash command and view submission payloads
- Processes `/new-repo` command to display a modal for creating new repositories
- Processes view submissions to push repository creation commands to Poppit
- Configurable via environment variables
- Docker and Docker Compose support with scratch runtime for minimal image size

## Prerequisites

- Go 1.24 or later
- Redis server
- Slack Bot Token with appropriate permissions (including `commands` and `views:write`)

## Configuration

The service can be configured via environment variables:

### Environment Variables

- `REDIS_ADDR` - Redis server address (default: `localhost:6379`)
- `REDIS_PASSWORD` - Redis server password (optional, set if your Redis requires authentication)
- `REDIS_CHANNEL` - Redis channel to subscribe to for slash commands (default: `slack-commands`)
- `REDIS_VIEW_SUBMISSION_CHANNEL` - Redis channel to subscribe to for view submissions (default: `slack-relay-view-submission`)
- `REDIS_POPPIT_LIST` - Redis list to push Poppit commands to (default: `poppit-commands`)
- `REDIS_SLACKLINER_LIST` - Redis list to push SlackLiner messages to (default: `slack_messages`)
- `SLACK_BOT_TOKEN` - Slack bot token (required)
- `GITHUB_ORG` - GitHub organization name for creating repositories (required)
- `WORKING_DIR` - Working directory for Poppit commands (default: `/tmp`)

## Running Locally

1. Install dependencies:
```bash
go mod download
```

2. Build the service:
```bash
go build -o slashviberepo
```

3. Run the service:
```bash
export SLACK_BOT_TOKEN=xoxb-your-token-here
export GITHUB_ORG=your-github-org
./slashviberepo
```

## Running with Docker Compose

1. Set the `SLACK_BOT_TOKEN` environment variable

2. Start the services:
```bash
docker-compose up --build
```

This will start:
- Redis server on port 6379
- SlashVibeRepo service connected to Redis

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

When the user submits the modal, the service will:
1. Receive the view submission payload on the `REDIS_VIEW_SUBMISSION_CHANNEL`
2. Extract the repository name and description from the submission
3. Generate a GitHub CLI command to create the repository
4. Push a Poppit command to the `REDIS_POPPIT_CHANNEL`
5. Send a confirmation message to the `#new-repo` Slack channel via SlackLiner with:
   - Repository name and link
   - Repository description (if provided)
   - 7-day TTL for automatic message cleanup

## View Submission Payload Format

The service expects view submission payloads in the following JSON format on the view submission channel:

```json
{
  "type": "view_submission",
  "view": {
    "state": {
      "values": {
        "repo-name": {
          "repo_name_input": {
            "type": "plain_text_input",
            "value": "ExampleRepo"
          }
        },
        "repo-description": {
          "repo_desc_input": {
            "type": "plain_text_input",
            "value": "Description for the example repository"
          }
        },
        "ai-prompt": {
          "ai_prompt_input": {
            "type": "plain_text_input",
            "value": "Sample AI prompt"
          }
        }
      }
    }
  }
}
```

## Poppit Command Output

When a view submission is processed, the service pushes a command to the Poppit list:

```json
{
  "repo": "your-org/ExampleRepo",
  "branch": "refs/heads/main",
  "type": "slash-vibe-new-repo",
  "dir": "/tmp",
  "commands": [
    "gh repo create your-org/ExampleRepo --public --add-readme --gitignore Go --description 'Description for the example repository'"
  ]
}
```

## Testing

You can test the service by publishing a message to the Redis channel:

```bash
redis-cli PUBLISH slack-commands '{"token":"test","team_id":"T123","team_domain":"test","channel_id":"C123","channel_name":"general","user_id":"U123","user_name":"testuser","command":"/new-repo","text":"my-repo","response_url":"https://example.com","trigger_id":"123.456.abc","api_app_id":"A123"}'
```

## License

MIT

