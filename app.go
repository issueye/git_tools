package main

import (
	"context"
	"fmt"
	"git-ai-tools/internal/ai"
	"git-ai-tools/internal/config"
	"git-ai-tools/internal/git"
	"git-ai-tools/internal/models"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"path/filepath"
	"strings"
)

// App struct
type App struct {
	ctx          context.Context
	gitService   *git.GitService
	aiService    *ai.AIService
	configService *config.ConfigService
}

// NewApp creates a new App application struct
func NewApp(configService *config.ConfigService) *App {
	return &App{
		gitService:    git.NewGitService(),
		aiService:     ai.NewAIService(),
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

// CloneRepository clones a remote repository
func (a *App) CloneRepository(url, path, branch string) error {
	opts := models.CloneOptions{
		URL:    url,
		Path:   path,
		Branch: branch,
	}

	if err := a.gitService.Clone(opts); err != nil {
		return err
	}

	// Add to recent repos
	a.configService.AddRecentRepo(opts.Path)

	return nil
}

// GetRemotes returns all remotes in the current repository
func (a *App) GetRemotes() ([]models.Remote, error) {
	return a.gitService.GetRemotes()
}

// AddRemote adds a new remote to the current repository
func (a *App) AddRemote(name, url string) error {
	return a.gitService.AddRemote(name, url)
}

// RemoveRemote removes a remote from the current repository
func (a *App) RemoveRemote(name string) error {
	return a.gitService.RemoveRemote(name)
}

// GetCurrentRepository returns the current repository path
func (a *App) GetCurrentRepository() string {
	return a.gitService.GetCurrentPath()
}

// GetStatus returns the git status
func (a *App) GetStatus() (*models.GitStatus, error) {
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
func (a *App) GetBranches() ([]models.Branch, error) {
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
func (a *App) GetLog(limit int) ([]models.CommitInfo, error) {
	return a.gitService.GetLog(limit)
}

// ============ AI Configuration ============

// GetAIConfig returns the AI configuration
func (a *App) GetAIConfig() models.AIConfig {
	return a.configService.GetAIConfig()
}

// SetAIConfig updates the AI configuration
func (a *App) SetAIConfig(config models.AIConfig) error {
	// First set the config to the AI service
	a.aiService.SetConfig(config)

	// Then validate the new config
	if err := a.aiService.ValidateConfig(); err != nil {
		return fmt.Errorf("AI configuration validation failed: %w", err)
	}

	// Finally save to config service
	if err := a.configService.SetAIConfig(config); err != nil {
		return fmt.Errorf("failed to save AI configuration: %w", err)
	}
	return nil
}

// TestAIConnection tests the AI service connection
// If config is provided, it validates the given config without modifying internal state
// If no config is provided (detected by empty Provider field), it validates the current configuration
func (a *App) TestAIConnection(config models.AIConfig) error {
	if config.Provider != "" {
		// Validate the provided config without modifying internal state
		if err := a.aiService.ValidateConfigParam(config); err != nil {
			return fmt.Errorf("AI configuration validation failed: %w", err)
		}
		return nil
	}
	// Validate current configuration
	if err := a.aiService.ValidateConfig(); err != nil {
		return fmt.Errorf("AI configuration validation failed: %w", err)
	}
	return nil
}

// ============ Utility Functions ============

// SelectDirectory opens a directory picker dialog
func (a *App) SelectDirectory() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not initialized")
	}
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Git Repository",
	})
	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}
	return path, nil
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
	currentPath := a.gitService.GetCurrentPath()
	if currentPath == "" {
		return map[string]interface{}{
			"path":       "",
			"branch":     "",
			"hasChanges": false,
			"isRepo":     false,
		}, nil
	}

	status, err := a.gitService.GetStatus()
	if err != nil {
		// If no repository is selected, return isRepo=false
		if strings.Contains(err.Error(), "no repository selected") {
			return map[string]interface{}{
				"path":       currentPath,
				"branch":     "",
				"hasChanges": false,
				"isRepo":     false,
			}, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"path":       currentPath,
		"branch":     status.Branch,
		"hasChanges": status.HasChanges,
		"isRepo":     status.IsRepo,
	}, nil
}

// RemoveRecentRepository removes a repository from recent list
func (a *App) RemoveRecentRepository(path string) error {
	return a.configService.RemoveRecentRepo(path)
}

// Push pushes the current branch to remote
func (a *App) Push(remote string) error {
	return a.gitService.Push(remote)
}

// Pull pulls changes from remote
func (a *App) Pull(remote string, branch string) error {
	return a.gitService.Pull(remote, branch)
}

// ResetType represents the type of reset (exposed for frontend)
type ResetType = git.ResetType

const (
	ResetSoft  ResetType = git.ResetSoft
	ResetMixed ResetType = git.ResetMixed
	ResetHard  ResetType = git.ResetHard
)

// Reset resets the current branch
func (a *App) Reset(resetType ResetType, commit string) error {
	return a.gitService.Reset(resetType, commit)
}

// Revert creates a new commit that undoes changes
func (a *App) Revert(commit string, noCommit bool) error {
	return a.gitService.Revert(commit, noCommit)
}

// GetRemoteNames returns available remote names
func (a *App) GetRemoteNames() ([]string, error) {
	return a.gitService.GetRemoteNames()
}

// Tag represents a git tag (type alias)
type Tag = git.Tag

// GetTags returns all tags
func (a *App) GetTags() ([]Tag, error) {
	tags, err := a.gitService.GetTags()
	if err != nil {
		return nil, err
	}
	// Convert to app-level Tag type
	result := make([]Tag, len(tags))
	for i, t := range tags {
		result[i] = Tag(t)
	}
	return result, nil
}

// CreateTag creates a new tag
func (a *App) CreateTag(name string, message string, commit string) error {
	return a.gitService.CreateTag(name, message, commit)
}

// DeleteTag deletes a tag
func (a *App) DeleteTag(name string) error {
	return a.gitService.DeleteTag(name)
}

// CheckoutTag checks out a tag
func (a *App) CheckoutTag(name string) error {
	return a.gitService.CheckoutTag(name)
}

// MergeBranch merges a branch
func (a *App) MergeBranch(branch string, noFF bool) error {
	return a.gitService.MergeBranch(branch, noFF)
}

// DeleteBranch deletes a branch
func (a *App) DeleteBranch(name string, force bool) error {
	return a.gitService.DeleteBranch(name, force)
}

// DiffBranches compares two branches
func (a *App) DiffBranches(branch1 string, branch2 string) (string, error) {
	return a.gitService.DiffBranches(branch1, branch2)
}

// GetCommitDetail returns detailed commit info
func (a *App) GetCommitDetail(commitHash string) (map[string]interface{}, error) {
	return a.gitService.GetCommitDetail(commitHash)
}
