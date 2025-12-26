package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/slack-go/slack"
)

// SlashCommandPayload represents the incoming slash command from Redis
type SlashCommandPayload struct {
	Token       string `json:"token"`
	TeamID      string `json:"team_id"`
	TeamDomain  string `json:"team_domain"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Command     string `json:"command"`
	Text        string `json:"text"`
	ResponseURL string `json:"response_url"`
	TriggerID   string `json:"trigger_id"`
	APIAppID    string `json:"api_app_id"`
}

// ViewSubmissionPayload represents the incoming view submission from Redis
type ViewSubmissionPayload struct {
	Type string `json:"type"`
	View struct {
		State struct {
			Values map[string]map[string]struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"values"`
		} `json:"state"`
	} `json:"view"`
}

// PoppitCommand represents the command message to be published to Poppit
type PoppitCommand struct {
	Repo     string   `json:"repo"`
	Branch   string   `json:"branch"`
	Type     string   `json:"type"`
	Dir      string   `json:"dir"`
	Commands []string `json:"commands"`
}

// Config holds the application configuration
type Config struct {
	RedisAddr                  string
	RedisPassword              string
	RedisChannel               string
	RedisViewSubmissionChannel string
	RedisPoppitList            string
	SlackToken                 string
	GithubOrg                  string
	WorkingDir                 string
}

func loadConfig() (*Config, error) {
	config := &Config{
		RedisAddr:                  getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:              getEnv("REDIS_PASSWORD", ""),
		RedisChannel:               getEnv("REDIS_CHANNEL", "slack-commands"),
		RedisViewSubmissionChannel: getEnv("REDIS_VIEW_SUBMISSION_CHANNEL", "slack-relay-view-submission"),
		RedisPoppitList:            getEnv("REDIS_POPPIT_LIST", "poppit:notifications"),
		SlackToken:                 getEnv("SLACK_BOT_TOKEN", ""),
		GithubOrg:                  getEnv("GITHUB_ORG", ""),
		WorkingDir:                 getEnv("WORKING_DIR", "/tmp"),
	}

	if config.SlackToken == "" {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN must be set via environment variable")
	}

	if config.GithubOrg == "" {
		return nil, fmt.Errorf("GITHUB_ORG must be set via environment variable")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	log.Println("Starting SlashVibeRepo service...")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Slack client
	slackClient := slack.New(config.SlackToken)

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword, // empty means no password
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Connected to Redis at %s", config.RedisAddr)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, cleaning up...")
		cancel()
	}()

	// Subscribe to Redis channels
	log.Printf("Subscribing to Redis channel: %s", config.RedisChannel)
	pubsub := redisClient.Subscribe(ctx, config.RedisChannel)
	defer pubsub.Close()

	log.Printf("Subscribing to Redis view submission channel: %s", config.RedisViewSubmissionChannel)
	viewSubmissionPubsub := redisClient.Subscribe(ctx, config.RedisViewSubmissionChannel)
	defer viewSubmissionPubsub.Close()

	// Wait for subscription confirmation
	_, err = pubsub.Receive(ctx)
	if err != nil {
		log.Fatalf("Failed to subscribe to Redis channel: %v", err)
	}
	log.Println("Successfully subscribed to Redis channel")

	_, err = viewSubmissionPubsub.Receive(ctx)
	if err != nil {
		log.Fatalf("Failed to subscribe to view submission channel: %v", err)
	}
	log.Println("Successfully subscribed to view submission channel")

	// Process messages from both channels
	ch := pubsub.Channel()
	viewSubmissionCh := viewSubmissionPubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}
			handleMessage(ctx, slackClient, msg.Payload)
		case msg := <-viewSubmissionCh:
			if msg == nil {
				continue
			}
			handleViewSubmission(ctx, redisClient, config, msg.Payload)
		}
	}
}

func handleMessage(ctx context.Context, slackClient *slack.Client, payload string) {
	log.Printf("Received message: %s", payload)

	var cmd SlashCommandPayload
	if err := json.Unmarshal([]byte(payload), &cmd); err != nil {
		log.Printf("Failed to unmarshal payload: %v", err)
		return
	}

	log.Printf("Processing command: %s from user: %s", cmd.Command, cmd.UserName)

	switch cmd.Command {
	case "/new-repo":
		handleNewRepoCommand(ctx, slackClient, &cmd)
	default:
		log.Printf("Unknown command: %s", cmd.Command)
	}
}

func handleNewRepoCommand(ctx context.Context, slackClient *slack.Client, cmd *SlashCommandPayload) {
	log.Printf("Handling /new-repo command with trigger_id: %s", cmd.TriggerID)

	modalView := createNewRepoModal(cmd.Text)

	_, err := slackClient.OpenViewContext(ctx, cmd.TriggerID, modalView)
	if err != nil {
		log.Printf("Failed to open modal: %v", err)
		return
	}

	log.Println("Successfully opened new-repo modal")
}

func createNewRepoModal(repoName string) slack.ModalViewRequest {
	// Create the repository name input block
	repoNameInput := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, "my-awesome-repo", false, false),
		"repo_name_input",
	)
	// Pre-populate the repository name if provided in the command text
	if repoName != "" {
		repoNameInput = repoNameInput.WithInitialValue(repoName)
	}

	repoNameBlock := slack.NewInputBlock(
		"repo-name",
		slack.NewTextBlockObject(slack.PlainTextType, "Repository Name", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Letters, numbers, hyphens only (no spaces)", false, false),
		repoNameInput,
	)

	// Create the repository description input block
	repoDescInput := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, "A short description of this project", false, false),
		"repo_desc_input",
	)

	repoDescBlock := slack.NewInputBlock(
		"repo-description",
		slack.NewTextBlockObject(slack.PlainTextType, "Repository Description", false, false),
		nil,
		repoDescInput,
	)
	repoDescBlock.Optional = true

	// Create the AI prompt input block
	aiPromptInput := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, "A simple Go service", false, false),
		"ai_prompt_input",
	).WithMultiline(true)

	aiPromptBlock := slack.NewInputBlock(
		"ai-prompt",
		slack.NewTextBlockObject(slack.PlainTextType, "Copilot Issue Prompt", false, false),
		slack.NewTextBlockObject(slack.PlainTextType, "Describe what Copilot should generate as the first issue", false, false),
		aiPromptInput,
	)
	aiPromptBlock.Optional = true

	// Create the modal view
	modalView := slack.ModalViewRequest{
		Type: slack.VTModal,
		Title: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "New Repo",
		},
		Close: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Cancel",
		},
		Submit: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Submit",
		},
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				repoNameBlock,
				repoDescBlock,
				aiPromptBlock,
			},
		},
	}

	return modalView
}

// handleViewSubmission processes view submission payloads from Redis
func handleViewSubmission(ctx context.Context, redisClient *redis.Client, config *Config, payload string) {
	log.Printf("Received view submission: %s", payload)

	var submission ViewSubmissionPayload
	if err := json.Unmarshal([]byte(payload), &submission); err != nil {
		log.Printf("Failed to unmarshal view submission payload: %v", err)
		return
	}

	// Extract values from the view state
	values := extractViewValues(submission)
	log.Printf("Extracted values: %+v", values)

	// Get repository name and description
	repoName, ok := values["repo-name"]
	if !ok || repoName == "" {
		log.Printf("Missing repository name in view submission")
		return
	}

	// Validate repository name (GitHub allows alphanumeric, hyphens, underscores, dots)
	if !isValidRepoName(repoName) {
		log.Printf("Invalid repository name: %s", repoName)
		return
	}

	repoDesc := values["repo-description"]

	// Build the repository full name
	repoFullName := fmt.Sprintf("%s/%s", config.GithubOrg, repoName)

	// Build the gh repo create command
	ghRepoCreateCmd := fmt.Sprintf("gh repo create %s --public --add-readme --gitignore Go", repoFullName)
	if repoDesc != "" {
		// Use single quotes for better safety, but escape any single quotes in the description
		escapedDesc := strings.ReplaceAll(repoDesc, `'`, `'\''`)
		ghRepoCreateCmd = fmt.Sprintf("%s --description '%s'", ghRepoCreateCmd, escapedDesc)
	}

	ghRepoCloneCmd := fmt.Sprintf("gh repo clone %s", repoFullName)

	ghVibeInitCmd := fmt.Sprintf("gh vibe init %s", repoFullName)

	// Create Poppit command message
	poppitCmd := PoppitCommand{
		Repo:   repoFullName,
		Branch: "refs/heads/main",
		Type:   "slash-vibe-new-repo",
		Dir:    config.WorkingDir,
		Commands: []string{
			ghRepoCreateCmd,
			ghRepoCloneCmd,
			ghVibeInitCmd,
		},
	}

	// Push to Poppit list
	poppitPayload, err := json.Marshal(poppitCmd)
	if err != nil {
		log.Printf("Failed to marshal Poppit command: %v", err)
		return
	}

	err = redisClient.RPush(ctx, config.RedisPoppitList, string(poppitPayload)).Err()
	if err != nil {
		log.Printf("Failed to push to Poppit list: %v", err)
		return
	}

	log.Printf("Successfully pushed Poppit command to list %s: %s", config.RedisPoppitList, string(poppitPayload))
}

// extractViewValues extracts values from the view submission state
// Equivalent to: jq '.view.state.values | map_values(.[] | .value)'
func extractViewValues(submission ViewSubmissionPayload) map[string]string {
	result := make(map[string]string)

	for blockID, blockValues := range submission.View.State.Values {
		// Each block has a map of action_id -> value object
		// In practice, each block contains exactly one action_id
		// We extract the first (and only) value from each block
		for _, valueObj := range blockValues {
			result[blockID] = valueObj.Value
			break
		}
	}

	return result
}

// isValidRepoName validates that the repository name contains only valid characters
// GitHub allows alphanumeric characters, hyphens, underscores, and dots
func isValidRepoName(name string) bool {
	if name == "" || len(name) > 100 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}
