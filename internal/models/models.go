package models

const (
	SevenDaysTTL           = 7 * 24 * 60 * 60
	NewRepoModalCallbackID = "create_github_repo_modal"
	NewProjectBranchPrefix = "new-project"
)

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

type ViewStateValue struct {
	Type           string `json:"type"`
	Value          string `json:"value"`
	SelectedOption *struct {
		Value string `json:"value"`
	} `json:"selected_option"`
	SelectedOptions []struct {
		Value string `json:"value"`
	} `json:"selected_options"`
}

type ViewSubmissionPayload struct {
	Type string `json:"type"`
	View struct {
		CallbackID string `json:"callback_id"`
		State      struct {
			Values map[string]map[string]ViewStateValue `json:"values"`
		} `json:"state"`
	} `json:"view"`
}

type PoppitCommand struct {
	Repo     string   `json:"repo"`
	Branch   string   `json:"branch"`
	Type     string   `json:"type"`
	Dir      string   `json:"dir"`
	Commands []string `json:"commands"`
}

type SlackLinerMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
	TTL     int    `json:"ttl,omitempty"`
}

type Config struct {
	RedisAddr                  string
	RedisPassword              string
	RedisChannel               string
	RedisViewSubmissionChannel string
	RedisPoppitList            string
	RedisSlackLinerList        string
	SlackToken                 string
	SlackChannelNewRepo        string
	GithubOrg                  string
	TemplateRepos              []string
	WorkingDir                 string
	VibeOpsWorkingDir          string
	GithubPrivateWorkingDir    string
	LogLevel                   string
	IssueCreationDelay         int
}

type FileConfig struct {
	Redis struct {
		Addr                  string `json:"addr"`
		Channel               string `json:"channel"`
		ViewSubmissionChannel string `json:"viewSubmissionChannel"`
		PoppitList            string `json:"poppitList"`
		SlackLinerList        string `json:"slackLinerList"`
	} `json:"redis"`
	Slack struct {
		ChannelNewRepo string `json:"channelNewRepo"`
	} `json:"slack"`
	Github struct {
		Org           string   `json:"org"`
		TemplateRepos []string `json:"templateRepos"`
	} `json:"github"`
	Paths struct {
		WorkingDir              string `json:"workingDir"`
		VibeOpsWorkingDir       string `json:"vibeopsWorkingDir"`
		GithubPrivateWorkingDir string `json:"githubPrivateWorkingDir"`
	} `json:"paths"`
	Logging struct {
		Level string `json:"level"`
	} `json:"logging"`
	IssueCreationDelay int `json:"issueCreationDelay"`
}
