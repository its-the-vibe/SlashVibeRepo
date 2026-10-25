package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Logger interface {
	Info(string, ...interface{})
	Error(string, ...interface{})
}

func NewClient(address, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
	})
}

func Listen(
	ctx context.Context,
	client *redis.Client,
	commandChannel, viewSubmissionChannel string,
	logger Logger,
	onCommand, onViewSubmission func(string),
) error {
	logger.Info("Subscribing to Redis channel: %s", commandChannel)
	commands := client.Subscribe(ctx, commandChannel)
	defer func() {
		if err := commands.Close(); err != nil {
			logger.Error("Failed to close Redis channel subscription: %v", err)
		}
	}()

	logger.Info("Subscribing to Redis view submission channel: %s", viewSubmissionChannel)
	submissions := client.Subscribe(ctx, viewSubmissionChannel)
	defer func() {
		if err := submissions.Close(); err != nil {
			logger.Error("Failed to close Redis view submission subscription: %v", err)
		}
	}()

	if _, err := commands.Receive(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to Redis channel: %w", err)
	}
	logger.Info("Successfully subscribed to Redis channel")
	if _, err := submissions.Receive(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to view submission channel: %w", err)
	}
	logger.Info("Successfully subscribed to view submission channel")

	commandMessages := commands.Channel()
	submissionMessages := submissions.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case message := <-commandMessages:
			if message != nil {
				onCommand(message.Payload)
			}
		case message := <-submissionMessages:
			if message != nil {
				onViewSubmission(message.Payload)
			}
		}
	}
}
