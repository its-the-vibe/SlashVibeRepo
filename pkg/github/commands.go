package github

import (
	"fmt"
	"strings"

	"github.com/its-the-vibe/SlashVibeRepo/internal/models"
)

func NewRepositoryCommands(repoFullName, templateRepo, description string) []string {
	var createCommand string
	if templateRepo != "" && templateRepo != "none" {
		createCommand = fmt.Sprintf("gh repo create %s --public --template %s", repoFullName, templateRepo)
	} else {
		createCommand = fmt.Sprintf("gh repo create %s --public --add-readme --gitignore Go", repoFullName)
	}
	if description != "" {
		escapedDescription := strings.ReplaceAll(description, `'`, `'\''`)
		createCommand = fmt.Sprintf("%s --description '%s'", createCommand, escapedDescription)
	}
	return []string{
		createCommand,
		fmt.Sprintf("gh repo clone %s", repoFullName),
		fmt.Sprintf("gh vibe init %s", repoFullName),
	}
}

func InitialIssueCommands(repoFullName, prompt, workingDir string, delay int) []string {
	if prompt == "" {
		return nil
	}
	escapedPrompt := strings.ReplaceAll(prompt, `'`, `'\''`)
	issueURL := fmt.Sprintf("https://github.com/%s/issues/1", repoFullName)
	return []string{
		fmt.Sprintf("sleep %d", delay),
		fmt.Sprintf("gh issue create -R %s --title 'Initial Issue' --body '%s'", repoFullName, escapedPrompt),
		fmt.Sprintf("%s/vibe-notify/bin/vibe-notify --config %s/vibe-notify/config.yaml issue %s", workingDir, workingDir, issueURL),
	}
}

func NewVibeOpsCommand(repoFullName, repoName, workingDir string) models.PoppitCommand {
	branchName := fmt.Sprintf("%s/bootstrap-%s", models.NewProjectBranchPrefix, repoName)
	prTitle := fmt.Sprintf("Add initial VibeOps configuration for %s", repoName)
	return models.PoppitCommand{
		Repo:   repoFullName,
		Branch: fmt.Sprintf("refs/heads/%s", branchName),
		Type:   "slash-vibe-vibeops-config",
		Dir:    workingDir,
		Commands: []string{
			"git checkout main",
			"git pull",
			"make build",
			fmt.Sprintf("git checkout -b %s", branchName),
			fmt.Sprintf("./vibeops new-project %s", repoName),
			"git add source",
			fmt.Sprintf("git commit -m \"Add initial configuration for %s\"", repoName),
			fmt.Sprintf("git push -u origin %s", branchName),
			fmt.Sprintf("gh pr create --title \"%s\" --body \"%s\" --draft", prTitle, "This PR adds the initial VibeOps configuration files for the project."),
		},
	}
}

func NewPrivateProjectCommand(org, repoFullName, repoName, workingDir string) models.PoppitCommand {
	branchName := fmt.Sprintf("%s/add-project-%s", models.NewProjectBranchPrefix, repoName)
	prTitle := fmt.Sprintf("Add %s to projects.json", repoName)
	prBody := fmt.Sprintf("Automatically generated PR to add %s project configuration to projects.json", repoName)
	commitMessage := fmt.Sprintf("Add %s to projects.json", repoName)
	return models.PoppitCommand{
		Repo:   fmt.Sprintf("%s/.github-private", org),
		Branch: fmt.Sprintf("refs/heads/%s", branchName),
		Type:   "slash-vibe-github-private-pr",
		Dir:    workingDir,
		Commands: []string{
			"git checkout main",
			"git pull",
			fmt.Sprintf("git checkout -b %s", branchName),
			"git add VibeOps/projects.json",
			fmt.Sprintf("git diff-index --quiet HEAD || git commit -m \"%s\"", commitMessage),
			fmt.Sprintf("git push -u origin %s", branchName),
			fmt.Sprintf("gh pr create --title \"%s\" --body \"%s\" --draft", prTitle, prBody),
		},
	}
}
