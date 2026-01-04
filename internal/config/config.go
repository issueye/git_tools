package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"git-ai-tools/internal/models"

	"github.com/google/uuid"
)

// ConfigService manages application configuration
type ConfigService struct {
	configPath      string
	reposPath       string
	config          *models.AppConfig
	repositories    *models.RepositoriesConfig
}

// NewConfigService creates a new ConfigService instance
func NewConfigService() *ConfigService {
	cs := &ConfigService{
		config: &models.AppConfig{
			AI: models.AIConfig{
				Provider: models.ProviderOpenAI,
				BaseURL:  "https://api.openai.com/v1",
				Model:    "gpt-4",
			},
			RecentRepos: []string{},
			Window: models.WindowConfig{
				Width:  1200,
				Height: 800,
			},
		},
		repositories: &models.RepositoriesConfig{
			Repositories: []models.Repository{},
		},
	}

	// Get config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		cs.configPath = "config.json"
		cs.reposPath = "repositories.json"
	} else {
		configDir = filepath.Join(configDir, "git-ai-tools")
		os.MkdirAll(configDir, 0755)
		cs.configPath = filepath.Join(configDir, "config.json")
		cs.reposPath = filepath.Join(configDir, "repositories.json")
	}

	// Load existing config
	cs.Load()
	// Load repositories
	cs.LoadRepositories()

	return cs
}

// Load loads configuration from file
func (c *ConfigService) Load() error {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use defaults
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, c.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// Save saves configuration to file
func (c *ConfigService) Save() error {
	data, err := json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func (c *ConfigService) GetConfig() *models.AppConfig {
	return c.config
}

// SetAIConfig updates the AI configuration
func (c *ConfigService) SetAIConfig(config models.AIConfig) error {
	c.config.AI = config
	return c.Save()
}

// GetAIConfig returns the AI configuration
func (c *ConfigService) GetAIConfig() models.AIConfig {
	return c.config.AI
}

// AddRecentRepo adds a repository to recent repos list
func (c *ConfigService) AddRecentRepo(path string) error {
	// Remove if already exists
	for i, repo := range c.config.RecentRepos {
		if repo == path {
			c.config.RecentRepos = append(c.config.RecentRepos[:i], c.config.RecentRepos[i+1:]...)
			break
		}
	}

	// Add to front
	c.config.RecentRepos = append([]string{path}, c.config.RecentRepos...)

	// Keep only last 10
	if len(c.config.RecentRepos) > 10 {
		c.config.RecentRepos = c.config.RecentRepos[:10]
	}

	return c.Save()
}

// GetRecentRepos returns the list of recent repositories
func (c *ConfigService) GetRecentRepos() []string {
	return c.config.RecentRepos
}

// RemoveRecentRepo removes a repository from recent repos list
func (c *ConfigService) RemoveRecentRepo(path string) error {
	for i, repo := range c.config.RecentRepos {
		if repo == path {
			c.config.RecentRepos = append(c.config.RecentRepos[:i], c.config.RecentRepos[i+1:]...)
			return c.Save()
		}
	}
	return nil
}

// SetWindowConfig updates the window configuration
func (c *ConfigService) SetWindowConfig(config models.WindowConfig) error {
	c.config.Window = config
	return c.Save()
}

// GetWindowConfig returns the window configuration
func (c *ConfigService) GetWindowConfig() models.WindowConfig {
	return c.config.Window
}

// GetConfigPath returns the configuration file path
func (c *ConfigService) GetConfigPath() string {
	return c.configPath
}

// ============= Repository Management =============

// LoadRepositories loads repositories from file
func (c *ConfigService) LoadRepositories() error {
	data, err := os.ReadFile(c.reposPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read repositories file: %w", err)
	}

	if err := json.Unmarshal(data, c.repositories); err != nil {
		return fmt.Errorf("failed to parse repositories file: %w", err)
	}

	return nil
}

// SaveRepositories saves repositories to file
func (c *ConfigService) SaveRepositories() error {
	data, err := json.MarshalIndent(c.repositories, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal repositories: %w", err)
	}

	if err := os.WriteFile(c.reposPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write repositories file: %w", err)
	}

	return nil
}

// GetAllRepositories returns all managed repositories
func (c *ConfigService) GetAllRepositories() []models.Repository {
	return c.repositories.Repositories
}

// GetRepository returns a repository by ID
func (c *ConfigService) GetRepository(id string) *models.Repository {
	for i := range c.repositories.Repositories {
		if c.repositories.Repositories[i].ID == id {
			return &c.repositories.Repositories[i]
		}
	}
	return nil
}

// GetRepositoryByPath returns a repository by path
func (c *ConfigService) GetRepositoryByPath(path string) *models.Repository {
	for i := range c.repositories.Repositories {
		if c.repositories.Repositories[i].Path == path {
			return &c.repositories.Repositories[i]
		}
	}
	return nil
}

// AddRepository adds a new repository
func (c *ConfigService) AddRepository(path, alias, description string) (*models.Repository, error) {
	// Check if already exists
	if c.GetRepositoryByPath(path) != nil {
		return nil, fmt.Errorf("repository already exists: %s", path)
	}

	now := time.Now().Format(time.RFC3339)
	repo := models.Repository{
		ID:          uuid.New().String(),
		Path:        path,
		Alias:       alias,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	c.repositories.Repositories = append(c.repositories.Repositories, repo)

	if err := c.SaveRepositories(); err != nil {
		return nil, err
	}

	return &repo, nil
}

// UpdateRepository updates an existing repository
func (c *ConfigService) UpdateRepository(id, alias, description string) (*models.Repository, error) {
	for i := range c.repositories.Repositories {
		if c.repositories.Repositories[i].ID == id {
			c.repositories.Repositories[i].Alias = alias
			c.repositories.Repositories[i].Description = description
			c.repositories.Repositories[i].UpdatedAt = time.Now().Format(time.RFC3339)

			if err := c.SaveRepositories(); err != nil {
				return nil, err
			}

			return &c.repositories.Repositories[i], nil
		}
	}

	return nil, fmt.Errorf("repository not found: %s", id)
}

// UpdateRepositoryAlias updates only the alias of a repository
func (c *ConfigService) UpdateRepositoryAlias(id, alias string) error {
	for i := range c.repositories.Repositories {
		if c.repositories.Repositories[i].ID == id {
			c.repositories.Repositories[i].Alias = alias
			c.repositories.Repositories[i].UpdatedAt = time.Now().Format(time.RFC3339)
			return c.SaveRepositories()
		}
	}
	return fmt.Errorf("repository not found: %s", id)
}

// DeleteRepository deletes a repository by ID
func (c *ConfigService) DeleteRepository(id string) error {
	for i := range c.repositories.Repositories {
		if c.repositories.Repositories[i].ID == id {
			c.repositories.Repositories = append(c.repositories.Repositories[:i], c.repositories.Repositories[i+1:]...)
			return c.SaveRepositories()
		}
	}
	return fmt.Errorf("repository not found: %s", id)
}

// SearchRepositories searches repositories by keyword
func (c *ConfigService) SearchRepositories(keyword string) []models.Repository {
	if keyword == "" {
		return c.repositories.Repositories
	}

	var result []models.Repository
	keyword = keyword
	for _, repo := range c.repositories.Repositories {
		// Search in path, alias, and description
		if contains(repo.Path, keyword) ||
			contains(repo.Alias, keyword) ||
			contains(repo.Description, keyword) {
			result = append(result, repo)
		}
	}
	return result
}

// contains checks if a string contains another string (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		// Simple case-insensitive comparison
		if ca >= 'A' && ca <= 'Z' {
			ca = ca - 'A' + 'a'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb = cb - 'A' + 'a'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// GetRepositoriesPath returns the repositories config path
func (c *ConfigService) GetRepositoriesPath() string {
	return c.reposPath
}
