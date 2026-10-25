package github

import (
	"reflect"
	"testing"
)

func TestNewRepositoryCommands(t *testing.T) {
	tests := []struct {
		name        string
		repo        string
		template    string
		description string
		want        []string
	}{
		{
			name: "without template or description",
			repo: "example-org/example",
			want: []string{
				"gh repo create example-org/example --public --add-readme --gitignore Go",
				"gh repo clone example-org/example",
				"gh vibe init example-org/example",
			},
		},
		{
			name:        "template and quoted description",
			repo:        "example-org/example",
			template:    "example-org/template",
			description: "it's a project",
			want: []string{
				"gh repo create example-org/example --public --template example-org/template --description 'it'\\''s a project'",
				"gh repo clone example-org/example",
				"gh vibe init example-org/example",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NewRepositoryCommands(test.repo, test.template, test.description); !reflect.DeepEqual(got, test.want) {
				t.Errorf("NewRepositoryCommands() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestInitialIssueCommands(t *testing.T) {
	got := InitialIssueCommands("example-org/example", "it's ready", "/work", 5)
	want := []string{
		"sleep 5",
		"gh issue create -R example-org/example --title 'Initial Issue' --body 'it'\\''s ready'",
		"/work/vibe-notify/bin/vibe-notify --config /work/vibe-notify/config.yaml issue https://github.com/example-org/example/issues/1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("InitialIssueCommands() = %#v, want %#v", got, want)
	}

	if got := InitialIssueCommands("example-org/example", "", "/work", 5); got != nil {
		t.Errorf("InitialIssueCommands() with an empty prompt = %#v, want nil", got)
	}
}

func TestProjectConfigurationCommands(t *testing.T) {
	vibeOps := NewVibeOpsCommand("example-org/example", "example", "/vibeops")
	if vibeOps.Repo != "example-org/example" ||
		vibeOps.Branch != "refs/heads/new-project/bootstrap-example" ||
		vibeOps.Type != "slash-vibe-vibeops-config" ||
		vibeOps.Dir != "/vibeops" {
		t.Errorf("NewVibeOpsCommand() returned unexpected metadata: %#v", vibeOps)
	}
	if got, want := vibeOps.Commands[3], "git checkout -b new-project/bootstrap-example"; got != want {
		t.Errorf("VibeOps branch command = %q, want %q", got, want)
	}

	privateRepo := NewPrivateProjectCommand("example-org", "example-org/example", "example", "/private")
	if privateRepo.Repo != "example-org/.github-private" ||
		privateRepo.Branch != "refs/heads/new-project/add-project-example" ||
		privateRepo.Type != "slash-vibe-github-private-pr" ||
		privateRepo.Dir != "/private" {
		t.Errorf("NewPrivateProjectCommand() returned unexpected metadata: %#v", privateRepo)
	}
	if got, want := privateRepo.Commands[4], "git diff-index --quiet HEAD || git commit -m \"Add example to projects.json\""; got != want {
		t.Errorf("private repository commit command = %q, want %q", got, want)
	}
}
