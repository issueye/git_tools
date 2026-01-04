package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitStatus represents the status of a git repository
type GitStatus struct {
	Branch       string            `json:"branch"`
	Staged       []FileChange      `json:"staged"`
	Unstaged     []FileChange      `json:"unstaged"`
	Untracked    []string          `json:"untracked"`
	IsRepo       bool              `json:"isRepo"`
	HasChanges   bool              `json:"hasChanges"`
}

// FileChange represents a changed file
type FileChange struct {
	Path     string `json:"path"`
	Status   string `json:"status"`
	 additions int   `json:"additions"`
	 deletions int   `json:"deletions"`
}

// Branch represents a git branch
type Branch struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"isCurrent"`
}

// CommitInfo represents a git commit
type CommitInfo struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Date    string `json:"date"`
}

// GitService handles git operations
type GitService struct {
	currentPath string
}

// NewGitService creates a new GitService instance
func NewGitService() *GitService {
	return &GitService{}
}

// SetPath sets the current working directory
func (g *GitService) SetPath(path string) error {
	// Check if it's a valid directory
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", path)
	}

	// Check if it's a git repository
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository: %s", path)
	}

	g.currentPath = path
	return nil
}

// GetCurrentPath returns the current path
func (g *GitService) GetCurrentPath() string {
	return g.currentPath
}

// GetStatus returns the current git status
func (g *GitService) GetStatus() (*GitStatus, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	status := &GitStatus{
		IsRepo:     true,
		Staged:     []FileChange{},
		Unstaged:   []FileChange{},
		Untracked:  []string{},
	}

	// Get current branch
	branch, err := g.runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		status.Branch = strings.TrimSpace(branch)
	}

	// Get branch status (ahead/behind)
	branchStatus, _ := g.runGitCommand("status", "-sb")
	if branchStatus != "" {
		status.Branch = strings.Fields(branchStatus)[0]
	}

	// Get status in porcelain format
	output, err := g.runGitCommand("status", "--porcelain=v1")
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	if output == "" {
		status.HasChanges = false
		return status, nil
	}

	status.HasChanges = true

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		if len(line) >= 3 {
			statusCode := line[:2]
			filePath := line[3:]

			// Handle renamed files
			if strings.Contains(filePath, "->") {
				parts := strings.Split(filePath, "->")
				filePath = strings.TrimSpace(parts[len(parts)-1])
			}

			change := FileChange{
				Path:   filePath,
				Status: getStatusDescription(statusCode),
			}

			switch statusCode[0] {
			case 'M', 'A', 'R', 'C':
				status.Staged = append(status.Staged, change)
			}

			if statusCode[0] == '?' {
				status.Untracked = append(status.Untracked, filePath)
			}

			if statusCode[1] == 'M' || (statusCode[0] == '?' && statusCode[1] == '?') {
				if statusCode[0] != '?' {
					status.Unstaged = append(status.Unstaged, change)
				}
			}
		}
	}

	return status, nil
}

// StageFiles stages the given files
func (g *GitService) StageFiles(files []string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if len(files) == 0 {
		return nil
	}

	args := append([]string{"add"}, files...)
	_, err := g.runGitCommand(args...)
	return err
}

// UnstageFiles unstages the given files
func (g *GitService) UnstageFiles(files []string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if len(files) == 0 {
		return nil
	}

	args := append([]string{"reset"}, files...)
	_, err := g.runGitCommand(args...)
	return err
}

// Commit creates a commit with the given message
func (g *GitService) Commit(message string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	_, err := g.runGitCommand("commit", "-m", message)
	return err
}

// GetBranches returns all branches
func (g *GitService) GetBranches() ([]Branch, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	output, err := g.runGitCommand("branch", "-a")
	if err != nil {
		return nil, err
	}

	var branches []Branch
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isCurrent := strings.HasPrefix(line, "*")
		name := strings.TrimPrefix(line, "*")
		name = strings.TrimSpace(name)
		name = strings.TrimPrefix(name, "remotes/")
		name = strings.TrimSpace(name)

		if name != "" && !strings.HasPrefix(name, "HEAD ->") {
			branches = append(branches, Branch{
				Name:      name,
				IsCurrent: isCurrent,
			})
		}
	}

	return branches, nil
}

// CheckoutBranch switches to the given branch
func (g *GitService) CheckoutBranch(branch string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	_, err := g.runGitCommand("checkout", branch)
	return err
}

// CreateBranch creates a new branch
func (g *GitService) CreateBranch(branch string, checkout bool) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	if checkout {
		_, err := g.runGitCommand("checkout", "-b", branch)
		return err
	}

	_, err := g.runGitCommand("branch", branch)
	return err
}

// GetDiff returns the diff for the given file
func (g *GitService) GetDiff(filePath string, staged bool) (string, error) {
	if g.currentPath == "" {
		return "", fmt.Errorf("no repository selected")
	}

	var args []string
	if staged {
		args = []string{"diff", "--staged", filePath}
	} else {
		args = []string{"diff", filePath}
	}

	return g.runGitCommand(args...)
}

// GetLog returns commit history
func (g *GitService) GetLog(limit int) ([]CommitInfo, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	format := "%H|%s|%an|%ad"
	output, err := g.runGitCommand("log", fmt.Sprintf("-%d", limit), "--pretty=format:"+format, "--date=iso")
	if err != nil {
		return nil, err
	}

	var commits []CommitInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0][:7],
				Message: parts[1],
				Author:  parts[2],
				Date:    parts[3],
			})
		}
	}

	return commits, nil
}

// DiscardChanges discards changes to the given file
func (g *GitService) DiscardChanges(filePath string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	_, err := g.runGitCommand("checkout", "--", filePath)
	return err
}

// runGitCommand executes a git command in the current directory
func (g *GitService) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if g.currentPath != "" {
		cmd.Dir = g.currentPath
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w\n%s", strings.Join(args, " "), err, string(output))
	}

	return strings.TrimSuffix(string(output), "\n"), nil
}

// getStatusDescription returns a human-readable status description
func getStatusDescription(code string) string {
	switch code {
	case "M ":
		return "Staged"
	case " M":
		return "Modified"
	case "MM":
		return "Modified (staged and unstaged)"
	case "A ":
		return "Added"
	case " D":
		return "Deleted"
	case "D ":
		return "Deleted (staged)"
	case "R ":
		return "Renamed"
	case "C ":
		return "Copied"
	case "??":
		return "Untracked"
	case "!!":
		return "Ignored"
	default:
		return "Unknown"
	}
}
