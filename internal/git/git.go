package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git-ai-tools/internal/models"
)

// GitService handles git operations
type GitService struct {
	currentPath string
}

// NewGitService creates a new GitService instance
func NewGitService() *GitService {
	return &GitService{}
}

// Clone clones a remote repository to the specified path
func (g *GitService) Clone(opts models.CloneOptions) error {
	if opts.URL == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if opts.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check if the destination path already exists
	if _, err := os.Stat(opts.Path); err == nil {
		// Check if it's not empty
		files, err := os.ReadDir(opts.Path)
		if err == nil && len(files) > 0 {
			return fmt.Errorf("destination path already exists and is not empty: %s", opts.Path)
		}
	}

	args := []string{"clone"}
	if opts.Branch != "" {
		args = append(args, "-b", opts.Branch)
	}
	args = append(args, opts.URL, opts.Path)

	_, err := g.runGitCommand(args...)
	if err != nil {
		return err
	}

	// Set the cloned repository as the current path
	g.currentPath = opts.Path
	return nil
}

// GetRemotes returns all remotes
func (g *GitService) GetRemotes() ([]models.Remote, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	output, err := g.runGitCommand("remote", "-v")
	if err != nil {
		return nil, err
	}

	var remotes []models.Remote
	lines := strings.Split(output, "\n")

	seen := make(map[string]bool)
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			url := parts[1]
			if !seen[name] {
				seen[name] = true
				remotes = append(remotes, models.Remote{
					Name: name,
					URL:  url,
				})
			}
		}
	}

	return remotes, nil
}

// AddRemote adds a new remote
func (g *GitService) AddRemote(name, url string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}
	if name == "" {
		return fmt.Errorf("remote name cannot be empty")
	}
	if url == "" {
		return fmt.Errorf("remote URL cannot be empty")
	}

	_, err := g.runGitCommand("remote", "add", name, url)
	return err
}

// RemoveRemote removes an existing remote
func (g *GitService) RemoveRemote(name string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}
	if name == "" {
		return fmt.Errorf("remote name cannot be empty")
	}

	_, err := g.runGitCommand("remote", "remove", name)
	return err
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
func (g *GitService) GetStatus() (*models.GitStatus, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	status := &models.GitStatus{
		IsRepo:     true,
		Staged:     []models.FileChange{},
		Unstaged:   []models.FileChange{},
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

			change := models.FileChange{
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
func (g *GitService) GetBranches() ([]models.Branch, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	output, err := g.runGitCommand("branch", "-a")
	if err != nil {
		return nil, err
	}

	var branches []models.Branch
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
			branches = append(branches, models.Branch{
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
func (g *GitService) GetLog(limit int) ([]models.CommitInfo, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	format := "%H|%s|%an|%ad"
	output, err := g.runGitCommand("log", fmt.Sprintf("-%d", limit), "--pretty=format:"+format, "--date=iso")
	if err != nil {
		return nil, err
	}

	var commits []models.CommitInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			commits = append(commits, models.CommitInfo{
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

// Push pushes the current branch to remote
func (g *GitService) Push(remote string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	args := []string{"push"}
	if remote != "" {
		args = append(args, remote)
	}

	_, err := g.runGitCommand(args...)
	return err
}

// Pull pulls changes from remote
func (g *GitService) Pull(remote string, branch string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	args := []string{"pull"}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}

	_, err := g.runGitCommand(args...)
	return err
}

// ResetType represents the type of reset
type ResetType string

const (
	ResetSoft  ResetType = "soft"
	ResetMixed ResetType = "mixed"
	ResetHard  ResetType = "hard"
)

// Reset resets the current branch to a specific commit
func (g *GitService) Reset(resetType ResetType, commit string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	args := []string{"reset", "--" + string(resetType)}
	if commit != "" {
		args = append(args, commit)
	}

	_, err := g.runGitCommand(args...)
	return err
}

// Revert creates a new commit that undoes the changes from a specific commit
func (g *GitService) Revert(commit string, noCommit bool) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	args := []string{"revert"}
	if noCommit {
		args = append(args, "--no-commit")
	}
	args = append(args, commit)

	_, err := g.runGitCommand(args...)
	return err
}

// GetRemotes returns a list of remote names
func (g *GitService) GetRemoteNames() ([]string, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	remotes, err := g.GetRemotes()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, r := range remotes {
		names = append(names, r.Name)
	}
	return names, nil
}

// Tag represents a git tag
type Tag struct {
	Name        string `json:"name"`
	CommitHash  string `json:"commitHash"`
	Message     string `json:"message"`
	IsAnnotated bool   `json:"isAnnotated"`
}

// GetTags returns all tags
func (g *GitService) GetTags() ([]Tag, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	output, err := g.runGitCommand("tag", "-l", "--format=%(refname:short)|%(objectname:short)|%(contents:subject)|%(contents:body)")
	if err != nil {
		return nil, err
	}

	var tags []Tag
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) >= 2 {
			tag := Tag{
				Name:        parts[0],
				CommitHash:  parts[1],
				IsAnnotated: len(parts) >= 3 && parts[2] != "",
			}
			if len(parts) >= 3 && parts[2] != "" {
				tag.Message = parts[2]
			}
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// CreateTag creates a new tag
func (g *GitService) CreateTag(name string, message string, commit string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if name == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	args := []string{"tag"}
	if message != "" {
		args = append(args, "-a", "-m", message)
	} else {
		args = append(args, name)
	}

	if commit != "" {
		args = append(args, commit)
	}

	_, err := g.runGitCommand(args...)
	return err
}

// DeleteTag deletes a tag
func (g *GitService) DeleteTag(name string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if name == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	_, err := g.runGitCommand("tag", "-d", name)
	return err
}

// CheckoutTag checks out a tag (creates detached HEAD)
func (g *GitService) CheckoutTag(name string) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if name == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	_, err := g.runGitCommand("checkout", name)
	return err
}

// MergeBranch merges a branch into current branch
func (g *GitService) MergeBranch(branch string, noFF bool) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	args := []string{"merge"}
	if noFF {
		args = append(args, "--no-ff")
	}
	args = append(args, branch)

	_, err := g.runGitCommand(args...)
	return err
}

// DeleteBranch deletes a branch
func (g *GitService) DeleteBranch(name string, force bool) error {
	if g.currentPath == "" {
		return fmt.Errorf("no repository selected")
	}

	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	args := []string{"branch"}
	if force {
		args = append(args, "-D")
	} else {
		args = append(args, "-d")
	}
	args = append(args, name)

	_, err := g.runGitCommand(args...)
	return err
}

// DiffBranches compares two branches and returns the diff
func (g *GitService) DiffBranches(branch1 string, branch2 string) (string, error) {
	if g.currentPath == "" {
		return "", fmt.Errorf("no repository selected")
	}

	output, err := g.runGitCommand("diff", branch1+"..."+branch2)
	return output, err
}

// GetCommitDetail returns detailed information about a commit
func (g *GitService) GetCommitDetail(commitHash string) (map[string]interface{}, error) {
	if g.currentPath == "" {
		return nil, fmt.Errorf("no repository selected")
	}

	// Get commit info
	output, err := g.runGitCommand("log", "-1", "--format=%H|%s|%an|%ad|%ae", "--date=iso", commitHash)
	if err != nil {
		return nil, fmt.Errorf("commit not found: %w", err)
	}

	parts := strings.SplitN(output, "|", 5)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid commit format")
	}

	result := map[string]interface{}{
		"hash":    parts[0],
		"message": parts[1],
		"author":  parts[2],
		"date":    parts[3],
	}

	// Get changed files
	filesOutput, _ := g.runGitCommand("show", "--stat", "--format=", commitHash)
	result["files"] = filesOutput

	return result, nil
}
