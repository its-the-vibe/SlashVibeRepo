package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	configpkg "github.com/its-the-vibe/SlashVibeRepo/internal/config"
	"github.com/its-the-vibe/SlashVibeRepo/internal/handlers"
	loggerpkg "github.com/its-the-vibe/SlashVibeRepo/pkg/logger"
	redispkg "github.com/its-the-vibe/SlashVibeRepo/pkg/redis"
	slack "github.com/slack-go/slack"
)

func main() {
	// Create initial logger for startup (before config is loaded)
	logger := loggerpkg.New("info")
	logger.Info("Starting SlashVibeRepo service...")

	config, err := configpkg.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	// Update logger with configured log level
	logger = loggerpkg.New(config.LogLevel)
	logger.Info("Log level set to: %s", config.LogLevel)

	// Initialize Slack client
	slackClient := slack.New(config.SlackToken)

	// Initialize Redis client
	redisClient := redispkg.NewClient(config.RedisAddr, config.RedisPassword)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis client: %v", err)
		}
	}()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis: %v", err)
	}
	logger.Info("Connected to Redis at %s", config.RedisAddr)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, cleaning up...")
		cancel()
	}()

	err = redispkg.Listen(
		ctx,
		redisClient,
		config.RedisChannel,
		config.RedisViewSubmissionChannel,
		logger,
		func(payload string) {
			handlers.HandleMessage(ctx, logger, slackClient, config, payload)
		},
		func(payload string) {
			handlers.HandleViewSubmission(ctx, logger, redisClient, config, payload)
		},
	)
	if err != nil {
		logger.Fatal("%v", err)
	}
}
