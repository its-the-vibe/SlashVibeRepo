package main

import (
	"github.com/its-the-vibe/SlashVibeRepo/internal/config"
	"github.com/its-the-vibe/SlashVibeRepo/internal/handlers"
	"github.com/its-the-vibe/SlashVibeRepo/internal/models"
	"github.com/its-the-vibe/SlashVibeRepo/pkg/logger"
	modal "github.com/its-the-vibe/SlashVibeRepo/pkg/slack"
	slackapi "github.com/slack-go/slack"
)

// Compatibility aliases keep the existing package-level API available to tests and integrations.
type SlashCommandPayload = models.SlashCommandPayload
type ViewStateValue = models.ViewStateValue
type ViewSubmissionPayload = models.ViewSubmissionPayload
type PoppitCommand = models.PoppitCommand
type SlackLinerMessage = models.SlackLinerMessage
type Config = models.Config
type fileConfig = models.FileConfig

type LogLevel = logger.Level

const (
	LogLevelDebug = logger.DebugLevel
	LogLevelInfo  = logger.InfoLevel
	LogLevelWarn  = logger.WarnLevel
	LogLevelError = logger.ErrorLevel

	SevenDaysTTL           = models.SevenDaysTTL
	NewRepoModalCallbackID = models.NewRepoModalCallbackID
	NewProjectBranchPrefix = models.NewProjectBranchPrefix
)

type Logger struct {
	level    LogLevel
	delegate *logger.Logger
}

func NewLogger(levelName string) *Logger {
	delegate := logger.New(levelName)
	return &Logger{level: delegate.Level(), delegate: delegate}
}

func (l *Logger) Debug(format string, v ...interface{}) { l.delegate.Debug(format, v...) }
func (l *Logger) Info(format string, v ...interface{})  { l.delegate.Info(format, v...) }
func (l *Logger) Warn(format string, v ...interface{})  { l.delegate.Warn(format, v...) }
func (l *Logger) Error(format string, v ...interface{}) { l.delegate.Error(format, v...) }
func (l *Logger) Fatal(format string, v ...interface{}) { l.delegate.Fatal(format, v...) }

func loadConfig() (*Config, error)                { return config.Load() }
func loadFileConfig() (fileConfig, string, error) { return config.LoadFile() }
func getEnv(key, defaultValue string) string      { return config.GetEnv(key, defaultValue) }

func createNewRepoModal(repoName string, templateRepos []string) slackapi.ModalViewRequest {
	return modal.NewRepoModal(repoName, templateRepos)
}

func extractViewValues(submission ViewSubmissionPayload) map[string]string {
	return handlers.ExtractViewValues(submission)
}

func isValidRepoName(name string) bool { return handlers.IsValidRepoName(name) }
