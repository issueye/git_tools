package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"git-ai-tools/internal/models"

	"github.com/google/uuid"
)

// TemplateService manages prompts and custom commands
type TemplateService struct {
	promptsPath  string
	commandsPath string
	prompts      *models.PromptsConfig
	commands     *models.CommandsConfig
}

// NewTemplateService creates a new TemplateService instance
func NewTemplateService() *TemplateService {
	ts := &TemplateService{
		prompts:  &models.PromptsConfig{Prompts: []models.Prompt{}},
		commands: &models.CommandsConfig{Commands: []models.Command{}},
	}

	// Get config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		ts.promptsPath = "prompts.json"
		ts.commandsPath = "commands.json"
	} else {
		configDir = filepath.Join(configDir, "git-ai-tools")
		os.MkdirAll(configDir, 0755)
		ts.promptsPath = filepath.Join(configDir, "prompts.json")
		ts.commandsPath = filepath.Join(configDir, "commands.json")
	}

	// Load existing data
	ts.LoadPrompts()
	ts.LoadCommands()

	// Create default prompts if none exist
	if len(ts.prompts.Prompts) == 0 {
		ts.createDefaultPrompts()
	}

	return ts
}

// ============= Prompt Operations =============

// LoadPrompts loads prompts from file
func (ts *TemplateService) LoadPrompts() error {
	data, err := os.ReadFile(ts.promptsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read prompts file: %w", err)
	}

	if err := json.Unmarshal(data, ts.prompts); err != nil {
		return fmt.Errorf("failed to parse prompts file: %w", err)
	}

	return nil
}

// SavePrompts saves prompts to file
func (ts *TemplateService) SavePrompts() error {
	data, err := json.MarshalIndent(ts.prompts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal prompts: %w", err)
	}

	if err := os.WriteFile(ts.promptsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write prompts file: %w", err)
	}

	return nil
}

// GetPrompts returns all prompts
func (ts *TemplateService) GetPrompts() []models.Prompt {
	return ts.prompts.Prompts
}

// GetPrompt returns a prompt by ID
func (ts *TemplateService) GetPrompt(id string) *models.Prompt {
	for i := range ts.prompts.Prompts {
		if ts.prompts.Prompts[i].ID == id {
			return &ts.prompts.Prompts[i]
		}
	}
	return nil
}

// GetDefaultPrompt returns the default prompt
func (ts *TemplateService) GetDefaultPrompt() *models.Prompt {
	for i := range ts.prompts.Prompts {
		if ts.prompts.Prompts[i].IsDefault {
			return &ts.prompts.Prompts[i]
		}
	}
	// Return first prompt if no default set
	if len(ts.prompts.Prompts) > 0 {
		return &ts.prompts.Prompts[0]
	}
	return nil
}

// CreatePrompt creates a new prompt
func (ts *TemplateService) CreatePrompt(name, description, template string, isDefault bool) (*models.Prompt, error) {
	now := time.Now().Format(time.RFC3339)

	prompt := models.Prompt{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Template:    template,
		IsDefault:   isDefault,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// If this is default, unset other defaults
	if isDefault {
		for i := range ts.prompts.Prompts {
			ts.prompts.Prompts[i].IsDefault = false
		}
	}

	ts.prompts.Prompts = append(ts.prompts.Prompts, prompt)

	if err := ts.SavePrompts(); err != nil {
		return nil, err
	}

	return &prompt, nil
}

// UpdatePrompt updates an existing prompt
func (ts *TemplateService) UpdatePrompt(id, name, description, template string, isDefault bool) (*models.Prompt, error) {
	for i := range ts.prompts.Prompts {
		if ts.prompts.Prompts[i].ID == id {
			// If setting as default, unset other defaults
			if isDefault {
				for j := range ts.prompts.Prompts {
					ts.prompts.Prompts[j].IsDefault = false
				}
			}

			ts.prompts.Prompts[i].Name = name
			ts.prompts.Prompts[i].Description = description
			ts.prompts.Prompts[i].Template = template
			ts.prompts.Prompts[i].IsDefault = isDefault
			ts.prompts.Prompts[i].UpdatedAt = time.Now().Format(time.RFC3339)

			if err := ts.SavePrompts(); err != nil {
				return nil, err
			}

			return &ts.prompts.Prompts[i], nil
		}
	}

	return nil, fmt.Errorf("prompt not found: %s", id)
}

// DeletePrompt deletes a prompt
func (ts *TemplateService) DeletePrompt(id string) error {
	for i := range ts.prompts.Prompts {
		if ts.prompts.Prompts[i].ID == id {
			ts.prompts.Prompts = append(ts.prompts.Prompts[:i], ts.prompts.Prompts[i+1:]...)
			return ts.SavePrompts()
		}
	}

	return fmt.Errorf("prompt not found: %s", id)
}

// SetDefaultPrompt sets a prompt as the default
func (ts *TemplateService) SetDefaultPrompt(id string) error {
	for i := range ts.prompts.Prompts {
		ts.prompts.Prompts[i].IsDefault = (ts.prompts.Prompts[i].ID == id)
		ts.prompts.Prompts[i].UpdatedAt = time.Now().Format(time.RFC3339)
	}
	return ts.SavePrompts()
}

// createDefaultPrompts creates default prompt templates
func (ts *TemplateService) createDefaultPrompts() {
	defaultPrompts := []models.Prompt{
		{
			ID:          uuid.New().String(),
			Name:        "生成提交信息",
			Description: "根据git diff生成标准化的提交信息",
			Template:    "请根据以下git diff生成一个提交信息。\n\n提交信息应该遵循 Conventional Commits规范：\n- feat: 新功能\n- fix: 修复bug\n- docs: 文档更新\n- style: 代码格式调整\n- refactor: 重构\n- test: 测试\n- chore: 构建过程或辅助工具的更改\n\ndiff:\n{{.Diff}}\n\n请只返回提交信息，不要其他解释。",
			IsDefault:   true,
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          uuid.New().String(),
			Name:        "分析代码变更",
			Description: "分析代码变更并提供审查建议",
			Template:    "请分析以下代码变更并提供审查建议：\n\n{{.Diff}}\n\n请从以下方面进行分析：\n1. 变更的目的\n2. 潜在的问题或风险\n3. 代码质量建议\n4. 安全性考虑",
			IsDefault:   false,
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	ts.prompts.Prompts = defaultPrompts
	ts.SavePrompts()
}

// ============= Command Operations =============

// LoadCommands loads commands from file
func (ts *TemplateService) LoadCommands() error {
	data, err := os.ReadFile(ts.commandsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read commands file: %w", err)
	}

	if err := json.Unmarshal(data, ts.commands); err != nil {
		return fmt.Errorf("failed to parse commands file: %w", err)
	}

	return nil
}

// SaveCommands saves commands to file
func (ts *TemplateService) SaveCommands() error {
	data, err := json.MarshalIndent(ts.commands, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}

	if err := os.WriteFile(ts.commandsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write commands file: %w", err)
	}

	return nil
}

// GetCommands returns all commands
func (ts *TemplateService) GetCommands() []models.Command {
	return ts.commands.Commands
}

// GetCommand returns a command by ID
func (ts *TemplateService) GetCommand(id string) *models.Command {
	for i := range ts.commands.Commands {
		if ts.commands.Commands[i].ID == id {
			return &ts.commands.Commands[i]
		}
	}
	return nil
}

// GetCommandsByCategory returns commands filtered by category
func (ts *TemplateService) GetCommandsByCategory(category string) []models.Command {
	var result []models.Command
	for i := range ts.commands.Commands {
		if ts.commands.Commands[i].Category == category {
			result = append(result, ts.commands.Commands[i])
		}
	}
	return result
}

// GetCategories returns all unique categories
func (ts *TemplateService) GetCategories() []string {
	categorySet := make(map[string]struct{})
	for i := range ts.commands.Commands {
		if ts.commands.Commands[i].Category != "" {
			categorySet[ts.commands.Commands[i].Category] = struct{}{}
		}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	return categories
}

// CreateCommand creates a new command
func (ts *TemplateService) CreateCommand(name, description, command, category string) (*models.Command, error) {
	now := time.Now().Format(time.RFC3339)

	if category == "" {
		category = "自定义"
	}

	cmd := models.Command{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Command:     command,
		Category:    category,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	ts.commands.Commands = append(ts.commands.Commands, cmd)

	if err := ts.SaveCommands(); err != nil {
		return nil, err
	}

	return &cmd, nil
}

// UpdateCommand updates an existing command
func (ts *TemplateService) UpdateCommand(id, name, description, command, category string) (*models.Command, error) {
	for i := range ts.commands.Commands {
		if ts.commands.Commands[i].ID == id {
			if category == "" {
				category = "自定义"
			}

			ts.commands.Commands[i].Name = name
			ts.commands.Commands[i].Description = description
			ts.commands.Commands[i].Command = command
			ts.commands.Commands[i].Category = category
			ts.commands.Commands[i].UpdatedAt = time.Now().Format(time.RFC3339)

			if err := ts.SaveCommands(); err != nil {
				return nil, err
			}

			return &ts.commands.Commands[i], nil
		}
	}

	return nil, fmt.Errorf("command not found: %s", id)
}

// DeleteCommand deletes a command
func (ts *TemplateService) DeleteCommand(id string) error {
	for i := range ts.commands.Commands {
		if ts.commands.Commands[i].ID == id {
			ts.commands.Commands = append(ts.commands.Commands[:i], ts.commands.Commands[i+1:]...)
			return ts.SaveCommands()
		}
	}

	return fmt.Errorf("command not found: %s", id)
}

// GetPromptsPath returns the prompts config path
func (ts *TemplateService) GetPromptsPath() string {
	return ts.promptsPath
}

// GetCommandsPath returns the commands config path
func (ts *TemplateService) GetCommandsPath() string {
	return ts.commandsPath
}
