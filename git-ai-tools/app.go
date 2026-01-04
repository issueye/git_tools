package main

import (
	"context"
	"fmt"
	"git-ai-tools/services"
	"os"
	"path/filepath"
)

// App struct
type App struct {
	ctx          context.Context
	gitService   *services.GitService
	aiService    *services.AIService
	configService *services.ConfigService
}

// NewApp creates a new App application struct
func NewApp(configService *services.ConfigService) *App {
	return &App{
		gitService:    services.NewGitService(),
		aiService:     services.NewAIService(),
		configService: configService,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load AI config
	if aiConfig := a.configService.GetAIConfig(); aiConfig.APIKey != "" {
		a.aiService.SetConfig(aiConfig)
	}
}

// ============ Repository Operations ============

// SelectRepository selects a git repository
func (a *App) SelectRepository(path string) error {
	if err := a.gitService.SetPath(path); err != nil {
		return err
	}

	// Add to recent repos
	a.configService.AddRecentRepo(path)

	return nil
}

// GetCurrentRepository returns the current repository path
func (a *App) GetCurrentRepository() string {
	return a.gitService.GetCurrentPath()
}

// GetStatus returns the git status
func (a *App) GetStatus() (*services.GitStatus, error) {
	return a.gitService.GetStatus()
}

// GetRecentRepositories returns recent repositories
func (a *App) GetRecentRepositories() []string {
	return a.configService.GetRecentRepos()
}

// ============ Stage Operations ============

// StageFiles stages the given files
func (a *App) StageFiles(files []string) error {
	return a.gitService.StageFiles(files)
}

// StageAll stages all changes
func (a *App) StageAll() error {
	return a.gitService.StageFiles([]string{"."})
}

// UnstageFiles unstages the given files
func (a *App) UnstageFiles(files []string) error {
	return a.gitService.UnstageFiles(files)
}

// UnstageAll unstages all changes
func (a *App) UnstageAll() error {
	return a.gitService.UnstageFiles([]string{"."})
}

// DiscardChanges discards changes to the given file
func (a *App) DiscardChanges(filePath string) error {
	return a.gitService.DiscardChanges(filePath)
}

// ============ Commit Operations ============

// Commit creates a commit with the given message
func (a *App) Commit(message string) error {
	return a.gitService.Commit(message)
}

// GenerateCommitMessage generates a commit message using AI
func (a *App) GenerateCommitMessage() (string, error) {
	status, err := a.gitService.GetStatus()
	if err != nil {
		return "", err
	}

	// Get diff of staged changes
	diff := ""
	for _, file := range status.Staged {
		fileDiff, err := a.gitService.GetDiff(file.Path, true)
		if err != nil {
			continue
		}
		diff += fmt.Sprintf("\n=== %s ===\n%s\n", file.Path, fileDiff)
	}

	if diff == "" {
		return "", fmt.Errorf("no staged changes to generate commit message for")
	}

	return a.aiService.GenerateCommitMessage(diff)
}

// ============ Branch Operations ============

// GetBranches returns all branches
func (a *App) GetBranches() ([]services.Branch, error) {
	return a.gitService.GetBranches()
}

// CheckoutBranch switches to the given branch
func (a *App) CheckoutBranch(branch string) error {
	return a.gitService.CheckoutBranch(branch)
}

// CreateBranch creates a new branch
func (a *App) CreateBranch(branch string, checkout bool) error {
	return a.gitService.CreateBranch(branch, checkout)
}

// ============ Diff Operations ============

// GetDiff returns the diff for the given file
func (a *App) GetDiff(filePath string, staged bool) (string, error) {
	return a.gitService.GetDiff(filePath, staged)
}

// ============ History Operations ============

// GetLog returns commit history
func (a *App) GetLog(limit int) ([]services.CommitInfo, error) {
	return a.gitService.GetLog(limit)
}

// ============ AI Configuration ============

// GetAIConfig returns the AI configuration
func (a *App) GetAIConfig() services.AIConfig {
	return a.configService.GetAIConfig()
}

// SetAIConfig updates the AI configuration
func (a *App) SetAIConfig(config services.AIConfig) error {
	// First set the config to the AI service
	a.aiService.SetConfig(config)

	// Then validate the new config
	if err := a.aiService.ValidateConfig(); err != nil {
		return err
	}

	// Finally save to config service
	return a.configService.SetAIConfig(config)
}

// TestAIConnection tests the AI service connection
func (a *App) TestAIConnection() error {
	return a.aiService.ValidateConfig()
}

// TestAIConnectionWithConfig tests the AI service connection with the given configuration
// without modifying the internal state
func (a *App) TestAIConnectionWithConfig(config services.AIConfig) error {
	return a.aiService.ValidateConfigParam(config)
}

// ============ Utility Functions ============

// SelectDirectory opens a directory picker dialog
func (a *App) SelectDirectory() (string, error) {
	// This is a placeholder - actual implementation would use runtime dialogs
	// For now, return current directory or home directory
	cwd, err := os.Getwd()
	if err != nil {
		home, _ := os.UserHomeDir()
		return home, nil
	}
	return cwd, nil
}

// IsValidGitRepository checks if a path is a valid git repository
func (a *App) IsValidGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}
	return false
}

// OpenRepositoryInTerminal opens the repository in terminal (placeholder)
func (a *App) OpenRepositoryInTerminal() error {
	// Placeholder - actual implementation would open terminal
	return nil
}

// OpenFileInEditor opens a file in editor (placeholder)
func (a *App) OpenFileInEditor(filePath string) error {
	// Placeholder - actual implementation would open file
	return nil
}

// GetRepositoryInfo returns repository information
func (a *App) GetRepositoryInfo() (map[string]interface{}, error) {
	status, err := a.gitService.GetStatus()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"path":     a.gitService.GetCurrentPath(),
		"branch":   status.Branch,
		"hasChanges": status.HasChanges,
		"isRepo":   status.IsRepo,
	}, nil
}

// RemoveRecentRepository removes a repository from recent list
func (a *App) RemoveRecentRepository(path string) error {
	return a.configService.RemoveRecentRepo(path)
}
