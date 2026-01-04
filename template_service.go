package main

import (
	"time"

	"git-ai-tools/internal/database"
	"git-ai-tools/internal/models"

	"github.com/google/uuid"
)

// TemplateService manages prompts and custom commands
type TemplateService struct{}

// NewTemplateService creates a new TemplateService instance
func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

// ============= Prompt Operations =============

// GetPrompts returns all prompts
func (ts *TemplateService) GetPrompts() []models.Prompt {
	var prompts []models.PromptDB
	database.GetDB().Order("created_at DESC").Find(&prompts)

	result := make([]models.Prompt, len(prompts))
	for i, p := range prompts {
		result[i] = models.Prompt{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Template:    p.Template,
			IsDefault:   p.IsDefault,
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result
}

// GetPrompt returns a prompt by ID
func (ts *TemplateService) GetPrompt(id string) *models.Prompt {
	var p models.PromptDB
	if err := database.GetDB().First(&p, "id = ?", id).Error; err != nil {
		return nil
	}
	return &models.Prompt{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Template:    p.Template,
		IsDefault:   p.IsDefault,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

// GetDefaultPrompt returns the default prompt
func (ts *TemplateService) GetDefaultPrompt() *models.Prompt {
	var p models.PromptDB
	if err := database.GetDB().First(&p, "is_default = ?", true).Error; err != nil {
		// Return first prompt if no default set
		var prompts []models.PromptDB
		database.GetDB().Order("created_at DESC").First(&prompts)
		if len(prompts) > 0 {
			p = prompts[0]
		} else {
			// Create default prompts if none exist
			ts.createDefaultPrompts()
			database.GetDB().Order("created_at DESC").First(&p)
		}
	}
	if p.ID == "" {
		return nil
	}
	return &models.Prompt{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Template:    p.Template,
		IsDefault:   p.IsDefault,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

// CreatePrompt creates a new prompt
func (ts *TemplateService) CreatePrompt(name, description, template string, isDefault bool) (*models.Prompt, error) {
	now := time.Now()

	// If this is default, unset other defaults
	if isDefault {
		database.GetDB().Model(&models.PromptDB{}).Update("is_default", false)
	}

	prompt := models.PromptDB{
		Name:        name,
		Description: description,
		Template:    template,
		IsDefault:   isDefault,
	}
	prompt.CreatedAt = now
	prompt.UpdatedAt = now
	prompt.ID = uuid.New().String()

	if err := database.GetDB().Create(&prompt).Error; err != nil {
		return nil, err
	}

	return &models.Prompt{
		ID:          prompt.ID,
		Name:        prompt.Name,
		Description: prompt.Description,
		Template:    prompt.Template,
		IsDefault:   prompt.IsDefault,
		CreatedAt:   prompt.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   prompt.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdatePrompt updates an existing prompt
func (ts *TemplateService) UpdatePrompt(id, name, description, template string, isDefault bool) (*models.Prompt, error) {
	var p models.PromptDB
	if err := database.GetDB().First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// If setting as default, unset other defaults
	if isDefault {
		database.GetDB().Model(&models.PromptDB{}).Where("id != ?", id).Update("is_default", false)
	}

	p.Name = name
	p.Description = description
	p.Template = template
	p.IsDefault = isDefault
	p.UpdatedAt = time.Now()

	if err := database.GetDB().Save(&p).Error; err != nil {
		return nil, err
	}

	return &models.Prompt{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Template:    p.Template,
		IsDefault:   p.IsDefault,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeletePrompt deletes a prompt
func (ts *TemplateService) DeletePrompt(id string) error {
	return database.GetDB().Where("id = ?", id).Delete(&models.PromptDB{}).Error
}

// SetDefaultPrompt sets a prompt as the default
func (ts *TemplateService) SetDefaultPrompt(id string) error {
	// Unset all defaults
	database.GetDB().Model(&models.PromptDB{}).Update("is_default", false)
	// Set new default
	return database.GetDB().Model(&models.PromptDB{}).Where("id = ?", id).Update("is_default", true).Error
}

// createDefaultPrompts creates default prompt templates
func (ts *TemplateService) createDefaultPrompts() {
	now := time.Now()

	defaultPrompts := []models.PromptDB{
		{
			Name:        "生成提交信息",
			Description: "根据git diff生成标准化的提交信息",
			Template:    "请根据以下git diff生成一个提交信息。\n\n提交信息应该遵循 Conventional Commits规范：\n- feat: 新功能\n- fix: 修复bug\n- docs: 文档更新\n- style: 代码格式调整\n- refactor: 重构\n- test: 测试\n- chore: 构建过程或辅助工具的更改\n\ndiff:\n{{.Diff}}\n\n请只返回提交信息，不要其他解释。",
			IsDefault:   true,
		},
		{
			Name:        "分析代码变更",
			Description: "分析代码变更并提供审查建议",
			Template:    "请分析以下代码变更并提供审查建议：\n\n{{.Diff}}\n\n请从以下方面进行分析：\n1. 变更的目的\n2. 潜在的问题或风险\n3. 代码质量建议\n4. 安全性考虑",
			IsDefault:   false,
		},
	}

	for i := range defaultPrompts {
		defaultPrompts[i].CreatedAt = now
		defaultPrompts[i].UpdatedAt = now
		defaultPrompts[i].ID = uuid.New().String()
	}

	database.GetDB().Create(&defaultPrompts)
}

// ============= Command Operations =============

// GetCommands returns all commands
func (ts *TemplateService) GetCommands() []models.Command {
	var commands []models.CommandDB
	database.GetDB().Order("created_at DESC").Find(&commands)

	result := make([]models.Command, len(commands))
	for i, c := range commands {
		result[i] = models.Command{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Command:     c.Command,
			Category:    c.Category,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result
}

// GetCommand returns a command by ID
func (ts *TemplateService) GetCommand(id string) *models.Command {
	var c models.CommandDB
	if err := database.GetDB().First(&c, "id = ?", id).Error; err != nil {
		return nil
	}
	return &models.Command{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Command:     c.Command,
		Category:    c.Category,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

// GetCommandsByCategory returns commands filtered by category
func (ts *TemplateService) GetCommandsByCategory(category string) []models.Command {
	var commands []models.CommandDB
	database.GetDB().Where("category = ?", category).Order("created_at DESC").Find(&commands)

	result := make([]models.Command, len(commands))
	for i, c := range commands {
		result[i] = models.Command{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Command:     c.Command,
			Category:    c.Category,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result
}

// GetCategories returns all unique categories
func (ts *TemplateService) GetCategories() []string {
	var categories []string
	database.GetDB().Model(&models.CommandDB{}).Distinct("category").Pluck("category", &categories)
	return categories
}

// CreateCommand creates a new command
func (ts *TemplateService) CreateCommand(name, description, command, category string) (*models.Command, error) {
	now := time.Now()

	if category == "" {
		category = "自定义"
	}

	cmd := models.CommandDB{
		Name:        name,
		Description: description,
		Command:     command,
		Category:    category,
	}
	cmd.CreatedAt = now
	cmd.UpdatedAt = now
	cmd.ID = uuid.New().String()

	if err := database.GetDB().Create(&cmd).Error; err != nil {
		return nil, err
	}

	return &models.Command{
		ID:          cmd.ID,
		Name:        cmd.Name,
		Description: cmd.Description,
		Command:     cmd.Command,
		Category:    cmd.Category,
		CreatedAt:   cmd.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   cmd.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateCommand updates an existing command
func (ts *TemplateService) UpdateCommand(id, name, description, command, category string) (*models.Command, error) {
	var c models.CommandDB
	if err := database.GetDB().First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if category == "" {
		category = "自定义"
	}

	c.Name = name
	c.Description = description
	c.Command = command
	c.Category = category
	c.UpdatedAt = time.Now()

	if err := database.GetDB().Save(&c).Error; err != nil {
		return nil, err
	}

	return &models.Command{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Command:     c.Command,
		Category:    c.Category,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteCommand deletes a command
func (ts *TemplateService) DeleteCommand(id string) error {
	return database.GetDB().Where("id = ?", id).Delete(&models.CommandDB{}).Error
}
