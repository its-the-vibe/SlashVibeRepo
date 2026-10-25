package slack

import (
	"github.com/its-the-vibe/SlashVibeRepo/internal/models"
	slackapi "github.com/slack-go/slack"
)

func NewRepoModal(repoName string, templateRepos []string) slackapi.ModalViewRequest {
	repoNameInput := slackapi.NewPlainTextInputBlockElement(
		slackapi.NewTextBlockObject(slackapi.PlainTextType, "my-awesome-repo", false, false),
		"repo_name_input",
	)
	if repoName != "" {
		repoNameInput = repoNameInput.WithInitialValue(repoName)
	}

	repoNameBlock := slackapi.NewInputBlock(
		"repo-name",
		slackapi.NewTextBlockObject(slackapi.PlainTextType, "Repository Name", false, false),
		slackapi.NewTextBlockObject(slackapi.PlainTextType, "Letters, numbers, hyphens only (no spaces)", false, false),
		repoNameInput,
	)
	repoDescInput := slackapi.NewPlainTextInputBlockElement(
		slackapi.NewTextBlockObject(slackapi.PlainTextType, "A short description of this project", false, false),
		"repo_desc_input",
	)
	repoDescBlock := slackapi.NewInputBlock(
		"repo-description",
		slackapi.NewTextBlockObject(slackapi.PlainTextType, "Repository Description", false, false),
		nil,
		repoDescInput,
	)
	repoDescBlock.Optional = true

	aiPromptInput := slackapi.NewPlainTextInputBlockElement(
		slackapi.NewTextBlockObject(slackapi.PlainTextType, "A simple Go service", false, false),
		"ai_prompt_input",
	).WithMultiline(true)
	aiPromptBlock := slackapi.NewInputBlock(
		"ai-prompt",
		slackapi.NewTextBlockObject(slackapi.PlainTextType, "Copilot Issue Prompt", false, false),
		slackapi.NewTextBlockObject(slackapi.PlainTextType, "Describe what Copilot should generate as the first issue", false, false),
		aiPromptInput,
	)
	aiPromptBlock.Optional = true

	vibeOpsCheckbox := slackapi.NewCheckboxGroupsBlockElement(
		"vibeops_config_checkbox",
		slackapi.NewOptionBlockObject(
			"create_vibeops_config",
			slackapi.NewTextBlockObject(slackapi.PlainTextType, "Create sample VibeOps configuration files", false, false),
			slackapi.NewTextBlockObject(slackapi.PlainTextType, "Generate initial VibeOps configuration and create a draft PR", false, false),
		),
	)
	vibeOpsCheckboxBlock := slackapi.NewInputBlock(
		"vibeops-config",
		slackapi.NewTextBlockObject(slackapi.PlainTextType, "VibeOps Configuration", false, false),
		nil,
		vibeOpsCheckbox,
	)
	vibeOpsCheckboxBlock.Optional = true

	blockSet := []slackapi.Block{repoNameBlock, repoDescBlock, aiPromptBlock, vibeOpsCheckboxBlock}
	if len(templateRepos) > 0 {
		templateOptions := []*slackapi.OptionBlockObject{
			slackapi.NewOptionBlockObject(
				"none",
				slackapi.NewTextBlockObject(slackapi.PlainTextType, "No Template", false, false),
				nil,
			),
		}
		for _, tmpl := range templateRepos {
			templateOptions = append(templateOptions, slackapi.NewOptionBlockObject(
				tmpl,
				slackapi.NewTextBlockObject(slackapi.PlainTextType, tmpl, false, false),
				nil,
			))
		}

		templateSelect := slackapi.NewOptionsSelectBlockElement(
			slackapi.OptTypeStatic,
			slackapi.NewTextBlockObject(slackapi.PlainTextType, "Select a template repository", false, false),
			"template_repo_select",
			templateOptions...,
		)
		templateSelect.InitialOption = templateOptions[0]
		templateBlock := slackapi.NewInputBlock(
			"template-repo",
			slackapi.NewTextBlockObject(slackapi.PlainTextType, "Template Repository", false, false),
			slackapi.NewTextBlockObject(slackapi.PlainTextType, "Optionally start from a template repository", false, false),
			templateSelect,
		)
		templateBlock.Optional = true
		blockSet = append(blockSet, templateBlock)
	}

	return slackapi.ModalViewRequest{
		Type:       slackapi.VTModal,
		CallbackID: models.NewRepoModalCallbackID,
		Title:      &slackapi.TextBlockObject{Type: slackapi.PlainTextType, Text: "New Repo"},
		Close:      &slackapi.TextBlockObject{Type: slackapi.PlainTextType, Text: "Cancel"},
		Submit:     &slackapi.TextBlockObject{Type: slackapi.PlainTextType, Text: "Submit"},
		Blocks:     slackapi.Blocks{BlockSet: blockSet},
	}
}
