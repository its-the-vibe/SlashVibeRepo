package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/its-the-vibe/SlashVibeRepo/internal/models"
)

func LoadFile() (models.FileConfig, string, error) {
	path := GetEnv("CONFIG_FILE", "config.json")
	var fc models.FileConfig
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fc, path, nil
		}
		return fc, path, fmt.Errorf("failed to read config file %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &fc); err != nil {
		return fc, path, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}
	return fc, path, nil
}

func Load() (*models.Config, error) {
	fc, configPath, err := LoadFile()
	if err != nil {
		return nil, err
	}
	if configPath != "" {
		if _, statErr := os.Stat(configPath); statErr == nil {
			log.Printf("[INFO] Loaded configuration from %s", configPath)
		}
	}

	fileOrDefault := func(fileVal, defaultVal string) string {
		if fileVal != "" {
			return fileVal
		}
		return defaultVal
	}

	config := &models.Config{
		RedisAddr:                  GetEnv("REDIS_ADDR", fileOrDefault(fc.Redis.Addr, "localhost:6379")),
		RedisPassword:              GetEnv("REDIS_PASSWORD", ""),
		RedisChannel:               GetEnv("REDIS_CHANNEL", fileOrDefault(fc.Redis.Channel, "slack-commands")),
		RedisViewSubmissionChannel: GetEnv("REDIS_VIEW_SUBMISSION_CHANNEL", fileOrDefault(fc.Redis.ViewSubmissionChannel, "slack-relay-view-submission")),
		RedisPoppitList:            GetEnv("REDIS_POPPIT_LIST", fileOrDefault(fc.Redis.PoppitList, "poppit:notifications")),
		RedisSlackLinerList:        GetEnv("REDIS_SLACKLINER_LIST", fileOrDefault(fc.Redis.SlackLinerList, "slack_messages")),
		SlackToken:                 GetEnv("SLACK_BOT_TOKEN", ""),
		SlackChannelNewRepo:        GetEnv("SLACK_CHANNEL_NEW_REPO", fileOrDefault(fc.Slack.ChannelNewRepo, "#new-repo")),
		GithubOrg:                  GetEnv("GITHUB_ORG", fc.Github.Org),
		TemplateRepos:              fc.Github.TemplateRepos,
		WorkingDir:                 GetEnv("WORKING_DIR", fileOrDefault(fc.Paths.WorkingDir, "/tmp")),
		VibeOpsWorkingDir:          GetEnv("VIBEOPS_WORKING_DIR", fc.Paths.VibeOpsWorkingDir),
		GithubPrivateWorkingDir:    GetEnv("GITHUB_PRIVATE_WORKING_DIR", fc.Paths.GithubPrivateWorkingDir),
		LogLevel:                   GetEnv("LOG_LEVEL", fileOrDefault(fc.Logging.Level, "info")),
	}

	delayDefault := 5
	if fc.IssueCreationDelay != 0 {
		delayDefault = fc.IssueCreationDelay
	}
	delayStr := os.Getenv("ISSUE_CREATION_DELAY")
	if delayStr != "" {
		delay, err := strconv.Atoi(delayStr)
		if err != nil {
			log.Printf("[WARN] Invalid ISSUE_CREATION_DELAY '%s', using default of %d seconds", delayStr, delayDefault)
			config.IssueCreationDelay = delayDefault
		} else {
			config.IssueCreationDelay = delay
		}
	} else {
		config.IssueCreationDelay = delayDefault
	}

	if config.SlackToken == "" {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN must be set via environment variable or .env file")
	}
	if config.GithubOrg == "" {
		return nil, fmt.Errorf("GITHUB_ORG must be set via config file, environment variable, or .env file")
	}
	return config, nil
}

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
