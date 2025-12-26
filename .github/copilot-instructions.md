# Copilot Instructions for SlashVibeRepo

## Project Overview

SlashVibeRepo is a Go service that bridges Slack slash commands with GitHub repository creation via Redis pub/sub. It subscribes to Redis channels to receive Slack slash command and view submission payloads, then processes them to create GitHub repositories using the GitHub CLI.

## Technology Stack

- **Language**: Go 1.25.5
- **Key Dependencies**:
  - `github.com/redis/go-redis/v9` - Redis client
  - `github.com/slack-go/slack` - Slack API client
- **Infrastructure**: Redis pub/sub, Docker, Docker Compose
- **Deployment**: Multi-stage Docker build with scratch runtime image

## Repository Structure

```
.
├── main.go              # Main application code (all logic in single file)
├── go.mod              # Go module definition
├── go.sum              # Go dependency checksums
├── Dockerfile          # Multi-stage Docker build
├── docker-compose.yml  # Docker Compose configuration
└── README.md           # Comprehensive documentation
```

## Build and Test Commands

### Building the Application

```bash
# Standard build
go build -o slashviberepo

# Build for production (static binary)
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o slashviberepo .

# Docker build
docker-compose up --build
```

### Testing

```bash
# Run tests (if any exist)
go test ./...

# Run with verbose output
go test -v ./...
```

### Running Locally

```bash
# Install dependencies
go mod download

# Run the service
export SLACK_BOT_TOKEN=xoxb-your-token-here
export GITHUB_ORG=your-github-org
go run main.go
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code for common issues
go vet ./...

# Run linter (if golangci-lint is available)
golangci-lint run
```

## Configuration

The application is configured via environment variables:

- `REDIS_ADDR` - Redis server address (default: `localhost:6379`)
- `REDIS_CHANNEL` - Redis channel for slash commands (default: `slack-commands`)
- `REDIS_VIEW_SUBMISSION_CHANNEL` - Redis channel for view submissions (default: `slack-relay-view-submission`)
- `REDIS_POPPIT_LIST` - Redis list for Poppit commands (default: `poppit:notifications`)
- `SLACK_BOT_TOKEN` - Slack bot token (required)
- `GITHUB_ORG` - GitHub organization name (required)
- `WORKING_DIR` - Working directory for Poppit commands (default: `/tmp`)

## Coding Conventions

### General Practices

1. **Single File Structure**: All code is currently in `main.go`. Maintain this structure for simplicity unless the file becomes too large (>1000 lines).

2. **Error Handling**: Always log errors with context:
   ```go
   if err != nil {
       log.Printf("Failed to perform operation: %v", err)
       return
   }
   ```

3. **Logging**: Use descriptive log messages with structured data:
   ```go
   log.Printf("Processing command: %s from user: %s", cmd.Command, cmd.UserName)
   ```

4. **JSON Handling**: Use struct tags for JSON marshaling/unmarshaling:
   ```go
   type MyStruct struct {
       FieldName string `json:"field_name"`
   }
   ```

5. **Context Usage**: Always pass `context.Context` to operations that may block or timeout.

### Naming Conventions

- Use camelCase for local variables: `repoName`, `redisClient`
- Use PascalCase for exported types and functions: `SlashCommandPayload`, `PoppitCommand`
- Use descriptive names for configuration fields: `RedisAddr`, `SlackToken`

### Docker Considerations

- The production image uses `scratch` (minimal base image)
- Binary must be statically linked (`CGO_ENABLED=0`)
- Include CA certificates for HTTPS requests
- Use multi-stage builds to minimize image size

## Key Patterns

### Redis Pub/Sub

The service uses two Redis subscriptions:
1. Slash commands on `REDIS_CHANNEL`
2. View submissions on `REDIS_VIEW_SUBMISSION_CHANNEL`

Always handle the subscription confirmation:
```go
_, err = pubsub.Receive(ctx)
if err != nil {
    log.Fatalf("Failed to subscribe: %v", err)
}
```

### Slack Modal Views

When creating modals, use the Slack block kit builder pattern:
- Input blocks for text fields
- Set `Optional = true` for optional fields
- Use appropriate hints and placeholders

### GitHub Repository Names

Repository names must be validated:
- Alphanumeric characters, hyphens, underscores, and dots only
- Maximum 100 characters
- Cannot be empty

## Security Considerations

1. **Input Validation**: Always validate repository names before processing
2. **Command Injection**: Escape single quotes in descriptions when building shell commands
3. **Read-Only Container**: Docker container runs with `read_only: true`

## Common Tasks

### Adding a New Slash Command

1. Add a new case in `handleMessage` switch statement
2. Create a handler function following the pattern: `handleXxxCommand(ctx context.Context, slackClient *slack.Client, cmd *SlashCommandPayload)`
3. Implement modal or response logic

### Modifying Modal Fields

1. Update `createNewRepoModal` function
2. Add new input blocks using `slack.NewInputBlock`
3. Update `extractViewValues` if needed
4. Modify `handleViewSubmission` to process new fields

### Changing Poppit Command Format

1. Update the `PoppitCommand` struct
2. Modify the command building logic in `handleViewSubmission`
3. Update README documentation

## Testing the Service

Manual testing with Redis CLI:
```bash
# Test slash command
redis-cli PUBLISH slack-commands '{"token":"test","team_id":"T123","team_domain":"test","channel_id":"C123","channel_name":"general","user_id":"U123","user_name":"testuser","command":"/new-repo","text":"my-repo","response_url":"https://example.com","trigger_id":"123.456.abc","api_app_id":"A123"}'
```

## Dependencies

When adding new dependencies:
1. Use `go get <package>` to add the dependency
2. Run `go mod tidy` to clean up
3. Update documentation if the dependency affects configuration or usage

## Troubleshooting

Common issues:
- **Redis connection failed**: Check `REDIS_ADDR` configuration
- **Slack API errors**: Verify `SLACK_BOT_TOKEN` is valid and has required scopes
- **Modal not opening**: Check that `trigger_id` is valid (expires after 3 seconds)
- **Poppit commands not received**: Verify `REDIS_POPPIT_LIST` name matches consumer

## Additional Notes

- The service uses graceful shutdown with signal handling (SIGTERM, SIGINT)
- All operations are non-blocking to handle multiple concurrent messages
- The Docker Compose setup includes host network access for Redis connectivity
- Repository names are pre-populated in the modal if provided in the slash command text
