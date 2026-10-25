# SlashVibeRepo

![CI](https://github.com/its-the-vibe/SlashVibeRepo/actions/workflows/ci.yaml/badge.svg?branch=main)

A simple Go service that subscribes to Slack slash commands and view submissions via Redis and performs operations.

## Features

- Subscribes to Redis channels to receive Slack slash command and view submission payloads
- Processes `/new-repo` command to display a modal for creating new repositories
- Processes view submissions to push repository creation commands to Poppit
- Configurable via environment variables
- Docker and Docker Compose support with scratch runtime for minimal image size

## Code organization

- The root `main.go` handles service startup; configuration loading is in `internal/config`.
- `internal/models` contains shared payload and configuration types.
- `internal/handlers` processes slash commands and view submissions.
- `pkg/slack`, `pkg/github`, `pkg/redis`, and `pkg/logger` provide focused Slack modal, GitHub command, Redis client/subscription, and logging functionality.

## Prerequisites

- Go 1.24 or later
- Redis server
- Slack Bot Token with appropriate permissions (including `commands` and `views:write`)

## Configuration

The service uses a two-file configuration approach that separates secrets from non-secret settings.

### 1. Secrets (`.env` file)

Create a `.env` file (git-ignored) from the provided example:

```bash
cp .env.example .env
```

Edit `.env` with your actual secrets:

```env
REDIS_PASSWORD=your-redis-password-if-any
SLACK_BOT_TOKEN=xoxb-your-slack-bot-token-here
```

### 2. Non-secret configuration (`config.json`)

Create a `config.json` file (git-ignored) from the provided example:

```bash
cp config.example.json config.json
```

Edit `config.json` as needed. The file path can be overridden with the `CONFIG_FILE` environment variable.

```json
{
  "redis": {
    "addr": "localhost:6379",
    "channel": "slack-commands",
    "viewSubmissionChannel": "slack-relay-view-submission",
    "poppitList": "poppit:notifications",
    "slackLinerList": "slack_messages"
  },
  "slack": {
    "channelNewRepo": "#new-repo"
  },
  "github": {
    "org": "my-org",
    "templateRepos": []
  },
  "paths": {
    "workingDir": "/tmp",
    "vibeopsWorkingDir": "",
    "githubPrivateWorkingDir": ""
  },
  "logging": {
    "level": "info"
  },
  "issueCreationDelay": 5
}
```

### 3. Loading order

1. Config file (`config.json` or `$CONFIG_FILE`) is read first and provides non-secret settings.
2. `.env` file provides secrets (`REDIS_PASSWORD`, `SLACK_BOT_TOKEN`).
3. Environment variables always override both the config file and `.env` values.

### Environment Variable Overrides

All settings can be overridden via environment variables:

| Variable | Config field | Default |
|---|---|---|
| `CONFIG_FILE` | *(path to config file)* | `config.json` |
| `REDIS_ADDR` | `redis.addr` | `localhost:6379` |
| `REDIS_PASSWORD` | *(secret, from `.env`)* | *(empty)* |
| `REDIS_CHANNEL` | `redis.channel` | `slack-commands` |
| `REDIS_VIEW_SUBMISSION_CHANNEL` | `redis.viewSubmissionChannel` | `slack-relay-view-submission` |
| `REDIS_POPPIT_LIST` | `redis.poppitList` | `poppit:notifications` |
| `REDIS_SLACKLINER_LIST` | `redis.slackLinerList` | `slack_messages` |
| `SLACK_BOT_TOKEN` | *(secret, from `.env`, required)* | *(none)* |
| `SLACK_CHANNEL_NEW_REPO` | `slack.channelNewRepo` | `#new-repo` |
| `GITHUB_ORG` | `github.org` | *(required)* |
| *(config file only)* | `github.templateRepos` | `[]` |
| `WORKING_DIR` | `paths.workingDir` | `/tmp` |
| `VIBEOPS_WORKING_DIR` | `paths.vibeopsWorkingDir` | *(empty)* |
| `GITHUB_PRIVATE_WORKING_DIR` | `paths.githubPrivateWorkingDir` | *(empty)* |
| `LOG_LEVEL` | `logging.level` | `info` |
| `ISSUE_CREATION_DELAY` | `issueCreationDelay` | `5` |

### Template Repositories

The `github.templateRepos` config file field accepts an optional list of GitHub repository names (in `owner/repo` format) that users can select as a starting template when creating a new repository via the `/new-repo` slash command.

```json
"github": {
  "org": "my-org",
  "templateRepos": [
    "my-org/template-go-service",
    "my-org/template-python-app"
  ]
}
```

When templates are configured, the *New Repo* modal displays a **Template Repository** dropdown containing a *No Template* option (the default) plus each configured template. If the user selects a template, the repository is created with `gh repo create --template <template>` instead of the default `--add-readme --gitignore Go` flags. If no templates are configured the dropdown is not shown.

### Log Levels

The service supports the following log levels (from most to least verbose):

- `debug` - Detailed diagnostic information including message payloads and extracted values
- `info` - General informational messages about normal operations (startup, connections, command processing)
- `warn` - Warning messages for unexpected but recoverable situations
- `error` - Error messages for failures that prevent operations

Example:
```bash
export LOG_LEVEL=debug  # Show all log messages
export LOG_LEVEL=info   # Default - show info, warn, and error messages
export LOG_LEVEL=warn   # Show only warnings and errors
export LOG_LEVEL=error  # Show only errors
```

## Development

Common development tasks are managed via the `Makefile`:

| Target  | Description                              |
|---------|------------------------------------------|
| `build` | Compile a static binary (`slashviberepo`) |
| `test`  | Run tests with race detection and coverage |
| `lint`  | Run `go vet` to check for common issues  |
| `clean` | Remove the binary and coverage artifacts |

```bash
make build   # build the binary
make test    # run tests
make lint    # lint the code
make clean   # clean up artifacts
```

## Running Locally

1. Install dependencies:
```bash
go mod download
```

2. Build the service:
```bash
make build
```

3. Run the service:
```bash
cp .env.example .env          # fill in your secrets
cp config.example.json config.json  # edit non-secret settings
./slashviberepo
```

## Running with Docker Compose

1. Create the local config file and set the `SLACK_BOT_TOKEN` environment variable:

```bash
cp config.example.json config.json
```

Edit `config.json` with the non-secret settings for the container. It is mounted read-only at `/config.json`.

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
- **VibeOps Configuration** (optional) - Checkbox to create sample VibeOps configuration files

When the user submits the modal, the service will:
1. Receive the view submission payload on the `REDIS_VIEW_SUBMISSION_CHANNEL`
2. Extract the repository name and description from the submission
3. Generate a GitHub CLI command to create the repository
4. Push a Poppit command to the `REDIS_POPPIT_CHANNEL`
5. If the "Create sample VibeOps configuration files" checkbox is selected and `VIBEOPS_WORKING_DIR` is set:
   - Push an additional Poppit command to create VibeOps configuration files
   - The VibeOps commands will run in the `VIBEOPS_WORKING_DIR` directory
   - Create a new branch `bootstrap-<repo-name>`
   - Run `vibeops new-project` to generate configuration files
   - Commit and push the changes
   - Create a draft PR with the configuration files
   - If `GITHUB_PRIVATE_WORKING_DIR` is also set:
     - After running `vibeops new-project`, which updates the `projects.json` file (a symlink to `.github-private` repository)
     - Create a new branch in `.github-private` repository: `add-project-<repo-name>`
     - Add and commit the updated `projects.json`
     - Push the branch to `.github-private` repository
     - Create a draft PR in `.github-private` with the `projects.json` changes
6. Send a confirmation message to the `#new-repo` Slack channel via SlackLiner with:
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
