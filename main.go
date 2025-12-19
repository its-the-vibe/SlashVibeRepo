package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
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

// Config holds the application configuration
type Config struct {
	RedisAddr    string
	RedisChannel string
	SlackToken   string
}

func loadConfig() (*Config, error) {
	config := &Config{
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		RedisChannel: getEnv("REDIS_CHANNEL", "slack-commands"),
		SlackToken:   getEnv("SLACK_BOT_TOKEN", ""),
	}

	// Try to load Slack token from .secret file if not set via env var
	if config.SlackToken == "" {
		if token, err := os.ReadFile(".secret"); err == nil {
			config.SlackToken = string(token)
		}
	}

	if config.SlackToken == "" {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN must be set via environment variable or .secret file")
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
	log.Println("Starting SlashVibe service...")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Slack client
	slackClient := slack.New(config.SlackToken)

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
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

	// Subscribe to Redis channel
	log.Printf("Subscribing to Redis channel: %s", config.RedisChannel)
	pubsub := redisClient.Subscribe(ctx, config.RedisChannel)
	defer pubsub.Close()

	// Wait for subscription confirmation
	_, err = pubsub.Receive(ctx)
	if err != nil {
		log.Fatalf("Failed to subscribe to Redis channel: %v", err)
	}
	log.Println("Successfully subscribed to Redis channel")

	// Process messages
	ch := pubsub.Channel()
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
		slack.NewTextBlockObject(slack.PlainTextType, "<repo_name>", false, false),
		"repo_name_input",
	)

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
