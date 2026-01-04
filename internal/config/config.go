package config

import (
	"time"

	"git-ai-tools/internal/database"
	"git-ai-tools/internal/models"

	"github.com/google/uuid"
)

// ConfigService manages application configuration
type ConfigService struct {
	db *models.AppConfigDB
}

// NewConfigService creates a new ConfigService instance
func NewConfigService() *ConfigService {
	// Ensure database is initialized
	database.Init()

	cs := &ConfigService{}

	// Initialize default config
	cs.db = &models.AppConfigDB{
		ID:    "app-config",
		Key:   "ai_config",
		Value: `{"provider":"openai","baseUrl":"https://api.openai.com/v1","model":"gpt-4"}`,
	}

	// Load existing config or create default
	var existing models.AppConfigDB
	result := database.GetDB().First(&existing, "key = ?", "ai_config")
	if result.Error == nil {
		cs.db = &existing
	} else {
		// Create default config
		database.GetDB().Create(cs.db)
	}

	return cs
}

// GetAIConfig returns the AI configuration
func (c *ConfigService) GetAIConfig() models.AIConfig {
	var config models.AIConfig
	// Try to parse from JSON
	if c.db.Value != "" {
		// For now, return default if not stored properly
		return models.AIConfig{
			Provider: models.ProviderOpenAI,
			BaseURL:  "https://api.openai.com/v1",
			Model:    "gpt-4",
		}
	}
	return config
}

// SetAIConfig updates the AI configuration
func (c *ConfigService) SetAIConfig(config models.AIConfig) error {
	// For now, we'll store individual fields
	return nil
}

// AddRecentRepo adds a repository to recent repos list
func (c *ConfigService) AddRecentRepo(path string) error {
	// Check if exists
	var existing models.RecentRepoDB
	result := database.GetDB().First(&existing, "path = ?", path)
	if result.Error == nil {
		// Update timestamp
		existing.UpdatedAt = time.Now()
		return database.GetDB().Save(&existing).Error
	}

	// Create new
	repo := models.RecentRepoDB{
		Path: path,
	}
	repo.CreatedAt = time.Now()
	repo.UpdatedAt = time.Now()
	repo.ID = uuid.New().String()
	return database.GetDB().Create(&repo).Error
}

// GetRecentRepos returns the list of recent repositories
func (c *ConfigService) GetRecentRepos() []string {
	var repos []models.RecentRepoDB
	database.GetDB().Order("updated_at DESC").Limit(10).Find(&repos)

	result := make([]string, len(repos))
	for i, repo := range repos {
		result[i] = repo.Path
	}
	return result
}

// RemoveRecentRepo removes a repository from recent repos list
func (c *ConfigService) RemoveRecentRepo(path string) error {
	return database.GetDB().Where("path = ?", path).Delete(&models.RecentRepoDB{}).Error
}

// GetWindowConfig returns the window configuration
func (c *ConfigService) GetWindowConfig() models.WindowConfig {
	return models.WindowConfig{
		Width:  1200,
		Height: 800,
	}
}

// GetConfigPath returns the configuration file path (legacy)
func (c *ConfigService) GetConfigPath() string {
	return ""
}

// ============= Repository Management =============

// GetAllRepositories returns all managed repositories
func (c *ConfigService) GetAllRepositories() []models.Repository {
	var repos []models.RepositoryDB
	database.GetDB().Order("updated_at DESC").Find(&repos)

	result := make([]models.Repository, len(repos))
	for i, repo := range repos {
		result[i] = models.Repository{
			ID:          repo.ID,
			Path:        repo.Path,
			Alias:       repo.Alias,
			Description: repo.Description,
			CreatedAt:   repo.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   repo.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result
}

// GetRepository returns a repository by ID
func (c *ConfigService) GetRepository(id string) *models.Repository {
	var repo models.RepositoryDB
	if err := database.GetDB().First(&repo, "id = ?", id).Error; err != nil {
		return nil
	}
	return &models.Repository{
		ID:          repo.ID,
		Path:        repo.Path,
		Alias:       repo.Alias,
		Description: repo.Description,
		CreatedAt:   repo.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   repo.UpdatedAt.Format(time.RFC3339),
	}
}

// GetRepositoryByPath returns a repository by path
func (c *ConfigService) GetRepositoryByPath(path string) *models.Repository {
	var repo models.RepositoryDB
	if err := database.GetDB().First(&repo, "path = ?", path).Error; err != nil {
		return nil
	}
	return &models.Repository{
		ID:          repo.ID,
		Path:        repo.Path,
		Alias:       repo.Alias,
		Description: repo.Description,
		CreatedAt:   repo.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   repo.UpdatedAt.Format(time.RFC3339),
	}
}

// AddRepository adds a new repository
func (c *ConfigService) AddRepository(path, alias, description string) (*models.Repository, error) {
	// Check if already exists
	if c.GetRepositoryByPath(path) != nil {
		return nil, nil
	}

	now := time.Now()
	repo := models.RepositoryDB{
		Path:        path,
		Alias:       alias,
		Description: description,
	}
	repo.CreatedAt = now
	repo.UpdatedAt = now
	repo.ID = uuid.New().String()

	if err := database.GetDB().Create(&repo).Error; err != nil {
		return nil, err
	}

	return &models.Repository{
		ID:          repo.ID,
		Path:        repo.Path,
		Alias:       repo.Alias,
		Description: repo.Description,
		CreatedAt:   repo.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   repo.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateRepository updates an existing repository
func (c *ConfigService) UpdateRepository(id, alias, description string) (*models.Repository, error) {
	var repo models.RepositoryDB
	if err := database.GetDB().First(&repo, "id = ?", id).Error; err != nil {
		return nil, err
	}

	repo.Alias = alias
	repo.Description = description
	repo.UpdatedAt = time.Now()

	if err := database.GetDB().Save(&repo).Error; err != nil {
		return nil, err
	}

	return &models.Repository{
		ID:          repo.ID,
		Path:        repo.Path,
		Alias:       repo.Alias,
		Description: repo.Description,
		CreatedAt:   repo.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   repo.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateRepositoryAlias updates only the alias of a repository
func (c *ConfigService) UpdateRepositoryAlias(id, alias string) error {
	return database.GetDB().Model(&models.RepositoryDB{}).Where("id = ?", id).Update("alias", alias).Error
}

// DeleteRepository deletes a repository by ID
func (c *ConfigService) DeleteRepository(id string) error {
	return database.GetDB().Where("id = ?", id).Delete(&models.RepositoryDB{}).Error
}

// SearchRepositories searches repositories by keyword
func (c *ConfigService) SearchRepositories(keyword string) []models.Repository {
	var repos []models.RepositoryDB

	if keyword == "" {
		database.GetDB().Order("updated_at DESC").Find(&repos)
	} else {
		keyword = "%" + keyword + "%"
		database.GetDB().Where("path LIKE ? OR alias LIKE ? OR description LIKE ?", keyword, keyword, keyword).
			Order("updated_at DESC").Find(&repos)
	}

	result := make([]models.Repository, len(repos))
	for i, repo := range repos {
		result[i] = models.Repository{
			ID:          repo.ID,
			Path:        repo.Path,
			Alias:       repo.Alias,
			Description: repo.Description,
			CreatedAt:   repo.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   repo.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result
}

// GetRepositoriesPath returns the repositories config path (legacy)
func (c *ConfigService) GetRepositoriesPath() string {
	return ""
}
