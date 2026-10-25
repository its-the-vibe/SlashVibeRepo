package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/its-the-vibe/SlashVibeRepo/internal/models"
	githubcmd "github.com/its-the-vibe/SlashVibeRepo/pkg/github"
	slackmodal "github.com/its-the-vibe/SlashVibeRepo/pkg/slack"
	redisclient "github.com/redis/go-redis/v9"
	slackapi "github.com/slack-go/slack"
)

type Logger interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
}

func handleMessage(ctx context.Context, logger Logger, slackClient *slackapi.Client, config *models.Config, payload string) {
	logger.Debug("Received message: %s", payload)

	var cmd models.SlashCommandPayload
	if err := json.Unmarshal([]byte(payload), &cmd); err != nil {
		logger.Error("Failed to unmarshal payload: %v", err)
		return
	}

	logger.Debug("Processing command: %s from user: %s", cmd.Command, cmd.UserName)

	switch cmd.Command {
	case "/new-repo":
		handleNewRepoCommand(ctx, logger, slackClient, config, &cmd)
	default:
		logger.Debug("Unknown command: %s", cmd.Command)
	}
}

func handleNewRepoCommand(ctx context.Context, logger Logger, slackClient *slackapi.Client, config *models.Config, cmd *models.SlashCommandPayload) {
	logger.Debug("Handling /new-repo command with trigger_id: %s", cmd.TriggerID)

	modalView := slackmodal.NewRepoModal(cmd.Text, config.TemplateRepos)

	_, err := slackClient.OpenViewContext(ctx, cmd.TriggerID, modalView)
	if err != nil {
		logger.Error("Failed to open modal: %v", err)
		return
	}

	logger.Info("Successfully opened new-repo modal for user: %s", cmd.UserName)
}

func handleViewSubmission(ctx context.Context, logger Logger, redisClient *redisclient.Client, config *models.Config, payload string) {
	logger.Debug("Received view submission: %s", payload)

	var submission models.ViewSubmissionPayload
	if err := json.Unmarshal([]byte(payload), &submission); err != nil {
		logger.Error("Failed to unmarshal view submission payload: %v", err)
		return
	}

	// Only handle our specific callback_id
	if submission.View.CallbackID != models.NewRepoModalCallbackID {
		logger.Debug("Ignoring view submission with callback_id: %s", submission.View.CallbackID)
		return
	}

	// Extract values from the view state
	values := extractViewValues(submission)
	logger.Debug("Extracted values: %+v", values)

	// Get repository name and description
	repoName, ok := values["repo-name"]
	if !ok || repoName == "" {
		logger.Error("Missing repository name in view submission")
		return
	}

	// Validate repository name (GitHub allows alphanumeric, hyphens, underscores, dots)
	if !isValidRepoName(repoName) {
		logger.Error("Invalid repository name: %s", repoName)
		return
	}

	repoDesc := values["repo-description"]

	// Build the repository full name
	repoFullName := fmt.Sprintf("%s/%s", config.GithubOrg, repoName)

	templateRepo := values["template-repo"]

	// Create Poppit command message
	poppitCmd := models.PoppitCommand{
		Repo:     repoFullName,
		Branch:   "refs/heads/main",
		Type:     "slash-vibe-new-repo",
		Dir:      config.WorkingDir,
		Commands: githubcmd.NewRepositoryCommands(repoFullName, templateRepo, repoDesc),
	}

	// Check if AI prompt is provided and create an initial issue
	aiPrompt := values["ai-prompt"]
	poppitCmd.Commands = append(poppitCmd.Commands, githubcmd.InitialIssueCommands(repoFullName, aiPrompt, config.WorkingDir, config.IssueCreationDelay)...)

	// Push to Poppit list
	poppitPayload, err := json.Marshal(poppitCmd)
	if err != nil {
		logger.Error("Failed to marshal Poppit command: %v", err)
		return
	}

	err = redisClient.RPush(ctx, config.RedisPoppitList, string(poppitPayload)).Err()
	if err != nil {
		logger.Error("Failed to push to Poppit list: %v", err)
		return
	}

	logger.Info("Successfully pushed Poppit command for repo: %s", repoFullName)
	logger.Debug("Poppit command payload: %s", string(poppitPayload))

	// Check if VibeOps configuration should be created
	createVibeOpsConfig := values["vibeops-config"] == "true"
	if createVibeOpsConfig && config.VibeOpsWorkingDir != "" {
		logger.Info("Creating VibeOps configuration for repo: %s", repoName)
		createVibeOpsConfiguration(ctx, logger, redisClient, config, repoFullName, repoName)
	} else if createVibeOpsConfig && config.VibeOpsWorkingDir == "" {
		logger.Warn("VibeOps configuration requested but VIBEOPS_WORKING_DIR is not set")
	}

	// Send confirmation message to SlackLiner
	sendNewRepoConfirmation(ctx, logger, redisClient, config, repoFullName, repoDesc)
}

// createVibeOpsConfiguration creates VibeOps configuration files and a draft PR
// Note: These commands are executed by Poppit, which handles error handling.
// The commands are run sequentially in the VibeOps working directory.

func createVibeOpsConfiguration(ctx context.Context, logger Logger, redisClient *redisclient.Client, config *models.Config, repoFullName, repoName string) {
	vibeOpsPoppitCmd := githubcmd.NewVibeOpsCommand(repoFullName, repoName, config.VibeOpsWorkingDir)

	// Push to Poppit list
	vibeOpsPayload, err := json.Marshal(vibeOpsPoppitCmd)
	if err != nil {
		logger.Error("Failed to marshal VibeOps Poppit command: %v", err)
		return
	}

	err = redisClient.RPush(ctx, config.RedisPoppitList, string(vibeOpsPayload)).Err()
	if err != nil {
		logger.Error("Failed to push VibeOps command to Poppit list: %v", err)
		return
	}

	logger.Info("Successfully pushed VibeOps configuration command for repo: %s", repoFullName)
	logger.Debug("VibeOps Poppit command payload: %s", string(vibeOpsPayload))

	// If GithubPrivateWorkingDir is configured, create a PR for projects.json in .github-private
	if config.GithubPrivateWorkingDir != "" {
		createGithubPrivatePR(ctx, logger, redisClient, config, repoFullName, repoName)
	} else {
		logger.Debug("GITHUB_PRIVATE_WORKING_DIR not set, skipping .github-private PR creation")
	}
}

// createGithubPrivatePR creates a PR in .github-private repository for projects.json changes
// Note: These commands are executed by Poppit, which handles error handling.
// The commands are run sequentially in the .github-private working directory.

func createGithubPrivatePR(ctx context.Context, logger Logger, redisClient *redisclient.Client, config *models.Config, repoFullName, repoName string) {
	githubPrivatePoppitCmd := githubcmd.NewPrivateProjectCommand(config.GithubOrg, repoFullName, repoName, config.GithubPrivateWorkingDir)

	// Push to Poppit list
	githubPrivatePayload, err := json.Marshal(githubPrivatePoppitCmd)
	if err != nil {
		logger.Error("Failed to marshal .github-private Poppit command: %v", err)
		return
	}

	err = redisClient.RPush(ctx, config.RedisPoppitList, string(githubPrivatePayload)).Err()
	if err != nil {
		logger.Error("Failed to push .github-private command to Poppit list: %v", err)
		return
	}

	logger.Info("Successfully pushed .github-private PR command for repo: %s", repoFullName)
	logger.Debug(".github-private Poppit command payload: %s", string(githubPrivatePayload))
}

// sendNewRepoConfirmation sends a confirmation message to SlackLiner

func sendNewRepoConfirmation(ctx context.Context, logger Logger, redisClient *redisclient.Client, config *models.Config, repoFullName, repoDesc string) {
	// Build the GitHub repository URL
	repoURL := fmt.Sprintf("https://github.com/%s", repoFullName)

	// Build the confirmation message
	confirmationText := fmt.Sprintf("✅ New repository creation initiated!\n\n*Repository:* <%s|%s>", repoURL, repoFullName)
	if repoDesc != "" {
		confirmationText = fmt.Sprintf("%s\n*Description:* %s", confirmationText, repoDesc)
	}

	// Create the SlackLiner message with 7 days TTL
	slackMessage := models.SlackLinerMessage{
		Channel: config.SlackChannelNewRepo,
		Text:    confirmationText,
		TTL:     models.SevenDaysTTL,
	}

	// Marshal to JSON
	messagePayload, err := json.Marshal(slackMessage)
	if err != nil {
		logger.Error("Failed to marshal SlackLiner message: %v", err)
		return
	}

	// Push to SlackLiner Redis list
	err = redisClient.RPush(ctx, config.RedisSlackLinerList, string(messagePayload)).Err()
	if err != nil {
		logger.Error("Failed to push to SlackLiner list: %v", err)
		return
	}

	logger.Info("Successfully sent confirmation message to SlackLiner for repo: %s", repoFullName)
}

// extractViewValues extracts values from the view submission state
// Equivalent to: jq '.view.state.values | map_values(.[] | .value)'
// For checkboxes, it checks if any options are selected

func extractViewValues(submission models.ViewSubmissionPayload) map[string]string {
	result := make(map[string]string)

	for blockID, blockValues := range submission.View.State.Values {
		// Each block has a map of action_id -> value object
		// In practice, each block contains exactly one action_id
		// We extract the first (and only) value from each block
		for _, valueObj := range blockValues {
			// Handle checkboxes (checkboxes type)
			switch valueObj.Type {
			case "checkboxes":
				if len(valueObj.SelectedOptions) > 0 {
					result[blockID] = "true"
				} else {
					result[blockID] = "false"
				}
			case "static_select":
				// Handle static select dropdowns
				if valueObj.SelectedOption != nil {
					result[blockID] = valueObj.SelectedOption.Value
				}
			default:
				// Handle text inputs and other types
				result[blockID] = valueObj.Value
			}
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
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '_' && c != '.' {
			return false
		}
	}
	return true
}

func HandleMessage(ctx context.Context, logger Logger, slackClient *slackapi.Client, config *models.Config, payload string) {
	handleMessage(ctx, logger, slackClient, config, payload)
}

func HandleViewSubmission(ctx context.Context, logger Logger, redisClient *redisclient.Client, config *models.Config, payload string) {
	handleViewSubmission(ctx, logger, redisClient, config, payload)
}

func ExtractViewValues(submission models.ViewSubmissionPayload) map[string]string {
	return extractViewValues(submission)
}

func IsValidRepoName(name string) bool {
	return isValidRepoName(name)
}
