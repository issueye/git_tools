package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig holds the application configuration
type AppConfig struct {
	AI          AIConfig   `json:"ai"`
	RecentRepos []string   `json:"recentRepos"`
	Window      WindowConfig `json:"window"`
}

// WindowConfig holds window state
type WindowConfig struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

// ConfigService manages application configuration
type ConfigService struct {
	configPath string
	config     *AppConfig
}

// NewConfigService creates a new ConfigService instance
func NewConfigService() *ConfigService {
	cs := &ConfigService{
		config: &AppConfig{
			AI: AIConfig{
				Provider: ProviderOpenAI,
				BaseURL:  "https://api.openai.com/v1",
				Model:    "gpt-4",
			},
			RecentRepos: []string{},
			Window: WindowConfig{
				Width:  1200,
				Height: 800,
			},
		},
	}

	// Get config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		cs.configPath = "config.json"
	} else {
		configDir = filepath.Join(configDir, "git-ai-tools")
		os.MkdirAll(configDir, 0755)
		cs.configPath = filepath.Join(configDir, "config.json")
	}

	// Load existing config
	cs.Load()

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
func (c *ConfigService) GetConfig() *AppConfig {
	return c.config
}

// SetAIConfig updates the AI configuration
func (c *ConfigService) SetAIConfig(config AIConfig) error {
	c.config.AI = config
	return c.Save()
}

// GetAIConfig returns the AI configuration
func (c *ConfigService) GetAIConfig() AIConfig {
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
func (c *ConfigService) SetWindowConfig(config WindowConfig) error {
	c.config.Window = config
	return c.Save()
}

// GetWindowConfig returns the window configuration
func (c *ConfigService) GetWindowConfig() WindowConfig {
	return c.config.Window
}

// GetConfigPath returns the configuration file path
func (c *ConfigService) GetConfigPath() string {
	return c.configPath
}
