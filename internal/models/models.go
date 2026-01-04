package models

// AIProvider represents the AI service provider
type AIProvider string

const (
	ProviderOpenAI AIProvider = "openai"
	ProviderClaude AIProvider = "claude"
	ProviderOllama AIProvider = "ollama"
)

// AIConfig holds AI service configuration
type AIConfig struct {
	Provider AIProvider `json:"provider"`
	APIKey   string     `json:"apiKey"`
	BaseURL  string     `json:"baseUrl"`
	Model    string     `json:"model"`
}

// AppConfig holds the application configuration
type AppConfig struct {
	AI          AIConfig     `json:"ai"`
	RecentRepos []string     `json:"recentRepos"`
	Window      WindowConfig `json:"window"`
}

// WindowConfig holds window state
type WindowConfig struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

// GitStatus represents the status of a git repository
type GitStatus struct {
	Branch     string       `json:"branch"`
	Staged     []FileChange `json:"staged"`
	Unstaged   []FileChange `json:"unstaged"`
	Untracked  []string     `json:"untracked"`
	IsRepo     bool         `json:"isRepo"`
	HasChanges bool         `json:"hasChanges"`
}

// FileChange represents a changed file
type FileChange struct {
	Path     string `json:"path"`
	Status   string `json:"status"`
	Additions int   `json:"additions"`
	Deletions int   `json:"deletions"`
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

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	URL    string `json:"url"`
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

// Remote represents a git remote
type Remote struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
