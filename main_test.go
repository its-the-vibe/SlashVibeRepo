package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/slack-go/slack"
)

// TestLoggerLevels tests that log levels are properly filtered
func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      string
		expectedLevel LogLevel
		message       string
		logFunc       func(*Logger, string, ...interface{})
		shouldLog     bool
	}{
		// Debug level tests
		{"DebugLevel_DebugMessage", "debug", LogLevelDebug, "debug message", (*Logger).Debug, true},
		{"DebugLevel_InfoMessage", "debug", LogLevelDebug, "info message", (*Logger).Info, true},
		{"DebugLevel_WarnMessage", "debug", LogLevelDebug, "warn message", (*Logger).Warn, true},
		{"DebugLevel_ErrorMessage", "debug", LogLevelDebug, "error message", (*Logger).Error, true},

		// Info level tests
		{"InfoLevel_DebugMessage", "info", LogLevelInfo, "debug message", (*Logger).Debug, false},
		{"InfoLevel_InfoMessage", "info", LogLevelInfo, "info message", (*Logger).Info, true},
		{"InfoLevel_WarnMessage", "info", LogLevelInfo, "warn message", (*Logger).Warn, true},
		{"InfoLevel_ErrorMessage", "info", LogLevelInfo, "error message", (*Logger).Error, true},

		// Warn level tests
		{"WarnLevel_DebugMessage", "warn", LogLevelWarn, "debug message", (*Logger).Debug, false},
		{"WarnLevel_InfoMessage", "warn", LogLevelWarn, "info message", (*Logger).Info, false},
		{"WarnLevel_WarnMessage", "warn", LogLevelWarn, "warn message", (*Logger).Warn, true},
		{"WarnLevel_ErrorMessage", "warn", LogLevelWarn, "error message", (*Logger).Error, true},

		// Error level tests
		{"ErrorLevel_DebugMessage", "error", LogLevelError, "debug message", (*Logger).Debug, false},
		{"ErrorLevel_InfoMessage", "error", LogLevelError, "info message", (*Logger).Info, false},
		{"ErrorLevel_WarnMessage", "error", LogLevelError, "warn message", (*Logger).Warn, false},
		{"ErrorLevel_ErrorMessage", "error", LogLevelError, "error message", (*Logger).Error, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			logger := NewLogger(tt.logLevel)

			// Verify the logger's level was set correctly
			if logger.level != tt.expectedLevel {
				t.Errorf("Expected log level %v, got %v", tt.expectedLevel, logger.level)
			}

			// Call the logging function
			tt.logFunc(logger, tt.message)

			output := buf.String()

			// Check if the message was logged
			if tt.shouldLog {
				if !strings.Contains(output, tt.message) {
					t.Errorf("Expected log output to contain '%s', but it didn't. Output: %s", tt.message, output)
				}
			} else {
				if strings.Contains(output, tt.message) {
					t.Errorf("Expected log output to NOT contain '%s', but it did. Output: %s", tt.message, output)
				}
			}
		})
	}
}

// TestNewLoggerDefaultLevel tests that the logger defaults to info level for invalid input
func TestNewLoggerDefaultLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"info", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"warning", LogLevelWarn},
		{"error", LogLevelError},
		{"DEBUG", LogLevelDebug},
		{"INFO", LogLevelInfo},
		{"invalid", LogLevelInfo}, // Should default to info
		{"", LogLevelInfo},        // Should default to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			logger := NewLogger(tt.input)
			if logger.level != tt.expected {
				t.Errorf("NewLogger(%q) = %v, want %v", tt.input, logger.level, tt.expected)
			}
		})
	}
}

// TestIsValidRepoName tests the repository name validation
func TestIsValidRepoName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid_AlphanumericHyphen", "my-awesome-repo", true},
		{"Valid_AlphanumericUnderscore", "my_awesome_repo", true},
		{"Valid_AlphanumericDot", "my.awesome.repo", true},
		{"Valid_Mixed", "My-Repo_2.0", true},
		{"Invalid_Space", "my repo", false},
		{"Invalid_SpecialChar", "my@repo", false},
		{"Invalid_Empty", "", false},
		{"Invalid_TooLong", strings.Repeat("a", 101), false},
		{"Valid_MaxLength", strings.Repeat("a", 100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidRepoName(tt.input)
			if got != tt.want {
				t.Errorf("isValidRepoName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// makeViewPayload is a test helper that builds a ViewSubmissionPayload from a flat
// map of blockID -> ViewStateValue entries.
func makeViewPayload(values map[string]map[string]ViewStateValue) ViewSubmissionPayload {
	var p ViewSubmissionPayload
	p.View.State.Values = values
	return p
}

// TestExtractViewValues tests the extraction of values from view submissions
func TestExtractViewValues(t *testing.T) {
	selectedOption := func(val string) *struct {
		Value string `json:"value"`
	} {
		v := struct {
			Value string `json:"value"`
		}{Value: val}
		return &v
	}

	tests := []struct {
		name     string
		payload  ViewSubmissionPayload
		expected map[string]string
	}{
		{
			name: "TextInputOnly",
			payload: makeViewPayload(map[string]map[string]ViewStateValue{
				"repo-name": {
					"repo_name_input": {Type: "plain_text_input", Value: "test-repo"},
				},
			}),
			expected: map[string]string{"repo-name": "test-repo"},
		},
		{
			name: "CheckboxSelected",
			payload: makeViewPayload(map[string]map[string]ViewStateValue{
				"vibeops-config": {
					"vibeops_config_checkbox": {
						Type: "checkboxes",
						SelectedOptions: []struct {
							Value string `json:"value"`
						}{{Value: "create_vibeops_config"}},
					},
				},
			}),
			expected: map[string]string{"vibeops-config": "true"},
		},
		{
			name: "CheckboxNotSelected",
			payload: makeViewPayload(map[string]map[string]ViewStateValue{
				"vibeops-config": {
					"vibeops_config_checkbox": {Type: "checkboxes"},
				},
			}),
			expected: map[string]string{"vibeops-config": "false"},
		},
		{
			name: "MixedTextAndCheckbox",
			payload: makeViewPayload(map[string]map[string]ViewStateValue{
				"repo-name": {
					"repo_name_input": {Type: "plain_text_input", Value: "my-repo"},
				},
				"vibeops-config": {
					"vibeops_config_checkbox": {
						Type: "checkboxes",
						SelectedOptions: []struct {
							Value string `json:"value"`
						}{{Value: "create_vibeops_config"}},
					},
				},
			}),
			expected: map[string]string{
				"repo-name":      "my-repo",
				"vibeops-config": "true",
			},
		},
		{
			name: "WithAIPrompt",
			payload: makeViewPayload(map[string]map[string]ViewStateValue{
				"repo-name": {
					"repo_name_input": {Type: "plain_text_input", Value: "my-repo"},
				},
				"ai-prompt": {
					"ai_prompt_input": {Type: "plain_text_input", Value: "Test AI Prompt"},
				},
			}),
			expected: map[string]string{
				"repo-name": "my-repo",
				"ai-prompt": "Test AI Prompt",
			},
		},
		{
			name: "StaticSelectWithTemplate",
			payload: makeViewPayload(map[string]map[string]ViewStateValue{
				"repo-name": {
					"repo_name_input": {Type: "plain_text_input", Value: "my-repo"},
				},
				"template-repo": {
					"template_repo_select": {
						Type:           "static_select",
						SelectedOption: selectedOption("my-org/template-go"),
					},
				},
			}),
			expected: map[string]string{
				"repo-name":     "my-repo",
				"template-repo": "my-org/template-go",
			},
		},
		{
			name: "StaticSelectNoTemplate",
			payload: makeViewPayload(map[string]map[string]ViewStateValue{
				"repo-name": {
					"repo_name_input": {Type: "plain_text_input", Value: "my-repo"},
				},
				"template-repo": {
					"template_repo_select": {
						Type:           "static_select",
						SelectedOption: selectedOption("none"),
					},
				},
			}),
			expected: map[string]string{
				"repo-name":     "my-repo",
				"template-repo": "none",
			},
		},
		{
			name: "StaticSelectNilOption",
			payload: makeViewPayload(map[string]map[string]ViewStateValue{
				"template-repo": {
					"template_repo_select": {Type: "static_select"},
				},
			}),
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractViewValues(tt.payload)
			if len(result) != len(tt.expected) {
				t.Errorf("extractViewValues() returned %d values, expected %d", len(result), len(tt.expected))
			}
			for key, expectedValue := range tt.expected {
				if gotValue, ok := result[key]; !ok {
					t.Errorf("extractViewValues() missing key %q", key)
				} else if gotValue != expectedValue {
					t.Errorf("extractViewValues()[%q] = %q, want %q", key, gotValue, expectedValue)
				}
			}
		})
	}
}

// TestCreateNewRepoModalWithTemplates tests that the modal includes a template selector when templates are configured
func TestCreateNewRepoModalWithTemplates(t *testing.T) {
	t.Run("NoTemplates", func(t *testing.T) {
		modal := createNewRepoModal("", nil)
		for _, block := range modal.Blocks.BlockSet {
			if ib, ok := block.(*slack.InputBlock); ok {
				if ib.BlockID == "template-repo" {
					t.Error("modal should not include template-repo block when no templates configured")
				}
			}
		}
	})

	t.Run("EmptyTemplateList", func(t *testing.T) {
		modal := createNewRepoModal("", []string{})
		for _, block := range modal.Blocks.BlockSet {
			if ib, ok := block.(*slack.InputBlock); ok {
				if ib.BlockID == "template-repo" {
					t.Error("modal should not include template-repo block when template list is empty")
				}
			}
		}
	})

	t.Run("WithTemplates", func(t *testing.T) {
		templates := []string{"my-org/template-go", "my-org/template-python"}
		modal := createNewRepoModal("", templates)
		found := false
		for _, block := range modal.Blocks.BlockSet {
			if ib, ok := block.(*slack.InputBlock); ok && ib.BlockID == "template-repo" {
				found = true
				if !ib.Optional {
					t.Error("template-repo block should be optional")
				}
			}
		}
		if !found {
			t.Error("modal should include template-repo block when templates are configured")
		}
	})

	t.Run("WithRepoNamePrePopulated", func(t *testing.T) {
		modal := createNewRepoModal("my-repo", []string{"my-org/template-go"})
		found := false
		for _, block := range modal.Blocks.BlockSet {
			if ib, ok := block.(*slack.InputBlock); ok && ib.BlockID == "repo-name" {
				found = true
				_ = ib
			}
		}
		if !found {
			t.Error("modal should include repo-name block")
		}
	})
}

// TestLoadConfigWithTemplateRepos tests that templateRepos are loaded from the config file
func TestLoadConfigWithTemplateRepos(t *testing.T) {
	content := `{
		"github": {"org": "test-org", "templateRepos": ["my-org/template-go", "my-org/template-python"]},
		"logging": {"level": "info"}
	}`
	f, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil && !os.IsNotExist(err) {
			t.Errorf("os.Remove() error = %v", err)
		}
	}()
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Setenv("CONFIG_FILE", f.Name()); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("SLACK_BOT_TOKEN", "test-token"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("CONFIG_FILE"); err != nil {
			t.Errorf("os.Unsetenv(CONFIG_FILE) error = %v", err)
		}
		if err := os.Unsetenv("SLACK_BOT_TOKEN"); err != nil {
			t.Errorf("os.Unsetenv(SLACK_BOT_TOKEN) error = %v", err)
		}
	}()

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	if len(config.TemplateRepos) != 2 {
		t.Fatalf("expected 2 template repos, got %d", len(config.TemplateRepos))
	}
	if config.TemplateRepos[0] != "my-org/template-go" {
		t.Errorf("TemplateRepos[0] = %q, want my-org/template-go", config.TemplateRepos[0])
	}
	if config.TemplateRepos[1] != "my-org/template-python" {
		t.Errorf("TemplateRepos[1] = %q, want my-org/template-python", config.TemplateRepos[1])
	}
}

// TestLoadConfigWithGithubPrivateWorkingDir tests that the GITHUB_PRIVATE_WORKING_DIR is correctly loaded
func TestLoadConfigWithGithubPrivateWorkingDir(t *testing.T) {
	// Set up test environment
	if err := os.Setenv("SLACK_BOT_TOKEN", "test-token"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("GITHUB_ORG", "test-org"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("GITHUB_PRIVATE_WORKING_DIR", "/test/github-private"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("SLACK_BOT_TOKEN"); err != nil {
			t.Errorf("os.Unsetenv(SLACK_BOT_TOKEN) error = %v", err)
		}
		if err := os.Unsetenv("GITHUB_ORG"); err != nil {
			t.Errorf("os.Unsetenv(GITHUB_ORG) error = %v", err)
		}
		if err := os.Unsetenv("GITHUB_PRIVATE_WORKING_DIR"); err != nil {
			t.Errorf("os.Unsetenv(GITHUB_PRIVATE_WORKING_DIR) error = %v", err)
		}
	}()

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	if config.GithubPrivateWorkingDir != "/test/github-private" {
		t.Errorf("GithubPrivateWorkingDir = %q, want %q", config.GithubPrivateWorkingDir, "/test/github-private")
	}
}

// TestLoadFileConfig tests the JSON config file parsing
func TestLoadFileConfig(t *testing.T) {
	t.Run("ValidConfigFile", func(t *testing.T) {
		content := `{
			"redis": {"addr": "redis:6380", "channel": "my-channel", "viewSubmissionChannel": "my-vsub", "poppitList": "my-poppit", "slackLinerList": "my-liner"},
			"slack": {"channelNewRepo": "#my-new-repo"},
			"github": {"org": "my-org"},
			"paths": {"workingDir": "/work", "vibeopsWorkingDir": "/vibeops", "githubPrivateWorkingDir": "/private"},
			"logging": {"level": "debug"},
			"issueCreationDelay": 10
		}`
		f, err := os.CreateTemp("", "config-*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Remove(f.Name()); err != nil && !os.IsNotExist(err) {
				t.Errorf("os.Remove() error = %v", err)
			}
		}()
		if _, err := f.WriteString(content); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

		if err := os.Setenv("CONFIG_FILE", f.Name()); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Unsetenv("CONFIG_FILE"); err != nil {
				t.Errorf("os.Unsetenv(CONFIG_FILE) error = %v", err)
			}
		}()

		fc, path, err := loadFileConfig()
		if err != nil {
			t.Fatalf("loadFileConfig() error: %v", err)
		}
		if path != f.Name() {
			t.Errorf("path = %q, want %q", path, f.Name())
		}
		if fc.Redis.Addr != "redis:6380" {
			t.Errorf("Redis.Addr = %q, want %q", fc.Redis.Addr, "redis:6380")
		}
		if fc.Github.Org != "my-org" {
			t.Errorf("Github.Org = %q, want %q", fc.Github.Org, "my-org")
		}
		if fc.IssueCreationDelay != 10 {
			t.Errorf("IssueCreationDelay = %d, want 10", fc.IssueCreationDelay)
		}
		if fc.Logging.Level != "debug" {
			t.Errorf("Logging.Level = %q, want debug", fc.Logging.Level)
		}
	})

	t.Run("MissingConfigFile", func(t *testing.T) {
		if err := os.Setenv("CONFIG_FILE", "/nonexistent/config.json"); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Unsetenv("CONFIG_FILE"); err != nil {
				t.Errorf("os.Unsetenv(CONFIG_FILE) error = %v", err)
			}
		}()

		fc, _, err := loadFileConfig()
		if err != nil {
			t.Fatalf("loadFileConfig() should not error for missing file, got: %v", err)
		}
		if fc.Redis.Addr != "" {
			t.Errorf("expected zero-value fileConfig for missing file")
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		f, err := os.CreateTemp("", "config-*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Remove(f.Name()); err != nil && !os.IsNotExist(err) {
				t.Errorf("os.Remove() error = %v", err)
			}
		}()
		if _, err := f.WriteString("{not valid json}"); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}

		if err := os.Setenv("CONFIG_FILE", f.Name()); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Unsetenv("CONFIG_FILE"); err != nil {
				t.Errorf("os.Unsetenv(CONFIG_FILE) error = %v", err)
			}
		}()

		_, _, err = loadFileConfig()
		if err == nil {
			t.Error("loadFileConfig() should return error for invalid JSON")
		}
	})
}

// TestLoadConfigFromFile tests that loadConfig reads values from config.json
func TestLoadConfigFromFile(t *testing.T) {
	content := `{
		"redis": {"addr": "redis-from-file:6379"},
		"github": {"org": "file-org"},
		"slack": {"channelNewRepo": "#from-file"},
		"logging": {"level": "warn"},
		"issueCreationDelay": 7
	}`
	f, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil && !os.IsNotExist(err) {
			t.Errorf("os.Remove() error = %v", err)
		}
	}()
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Setenv("CONFIG_FILE", f.Name()); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("SLACK_BOT_TOKEN", "test-token"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("CONFIG_FILE"); err != nil {
			t.Errorf("os.Unsetenv(CONFIG_FILE) error = %v", err)
		}
		if err := os.Unsetenv("SLACK_BOT_TOKEN"); err != nil {
			t.Errorf("os.Unsetenv(SLACK_BOT_TOKEN) error = %v", err)
		}
	}()

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}
	if config.RedisAddr != "redis-from-file:6379" {
		t.Errorf("RedisAddr = %q, want redis-from-file:6379", config.RedisAddr)
	}
	if config.GithubOrg != "file-org" {
		t.Errorf("GithubOrg = %q, want file-org", config.GithubOrg)
	}
	if config.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want warn", config.LogLevel)
	}
	if config.IssueCreationDelay != 7 {
		t.Errorf("IssueCreationDelay = %d, want 7", config.IssueCreationDelay)
	}
}

// TestLoadConfigEnvOverridesFile tests that env vars override config file values
func TestLoadConfigEnvOverridesFile(t *testing.T) {
	content := `{"redis": {"addr": "file-redis:6379"}, "github": {"org": "file-org"}}`
	f, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil && !os.IsNotExist(err) {
			t.Errorf("os.Remove() error = %v", err)
		}
	}()
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Setenv("CONFIG_FILE", f.Name()); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("SLACK_BOT_TOKEN", "test-token"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("REDIS_ADDR", "env-redis:6380"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Unsetenv("CONFIG_FILE"); err != nil {
			t.Errorf("os.Unsetenv(CONFIG_FILE) error = %v", err)
		}
		if err := os.Unsetenv("SLACK_BOT_TOKEN"); err != nil {
			t.Errorf("os.Unsetenv(SLACK_BOT_TOKEN) error = %v", err)
		}
		if err := os.Unsetenv("REDIS_ADDR"); err != nil {
			t.Errorf("os.Unsetenv(REDIS_ADDR) error = %v", err)
		}
	}()

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}
	if config.RedisAddr != "env-redis:6380" {
		t.Errorf("RedisAddr = %q, want env-redis:6380 (env should override file)", config.RedisAddr)
	}
}

// TestLoadConfigWithIssueCreationDelay tests that the ISSUE_CREATION_DELAY is correctly loaded
func TestLoadConfigWithIssueCreationDelay(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		{"ValidDelay", "10", 10},
		{"DefaultDelay", "", 5},
		{"InvalidDelay", "abc", 5},
		{"ZeroDelay", "0", 0},
		{"NegativeDelay", "-1", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			if err := os.Setenv("SLACK_BOT_TOKEN", "test-token"); err != nil {
				t.Fatal(err)
			}
			if err := os.Setenv("GITHUB_ORG", "test-org"); err != nil {
				t.Fatal(err)
			}
			if tt.envValue != "" {
				if err := os.Setenv("ISSUE_CREATION_DELAY", tt.envValue); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Unsetenv("ISSUE_CREATION_DELAY"); err != nil {
					t.Fatal(err)
				}
			}
			defer func() {
				if err := os.Unsetenv("SLACK_BOT_TOKEN"); err != nil {
					t.Errorf("os.Unsetenv(SLACK_BOT_TOKEN) error = %v", err)
				}
				if err := os.Unsetenv("GITHUB_ORG"); err != nil {
					t.Errorf("os.Unsetenv(GITHUB_ORG) error = %v", err)
				}
				if err := os.Unsetenv("ISSUE_CREATION_DELAY"); err != nil {
					t.Errorf("os.Unsetenv(ISSUE_CREATION_DELAY) error = %v", err)
				}
			}()

			config, err := loadConfig()
			if err != nil {
				t.Fatalf("loadConfig() failed: %v", err)
			}

			if config.IssueCreationDelay != tt.expected {
				t.Errorf("IssueCreationDelay = %d, want %d", config.IssueCreationDelay, tt.expected)
			}
		})
	}
}
