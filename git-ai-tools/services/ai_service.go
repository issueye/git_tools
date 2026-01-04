package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

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

// AIService handles AI operations for generating commit messages
type AIService struct {
	config AIConfig
	client *http.Client
}

// NewAIService creates a new AIService instance
func NewAIService() *AIService {
	return &AIService{
		client: &http.Client{},
		config: AIConfig{
			Provider: ProviderOpenAI,
			BaseURL:  "https://api.openai.com/v1",
			Model:    "gpt-4",
		},
	}
}

// SetConfig updates the AI service configuration
func (a *AIService) SetConfig(config AIConfig) {
	a.config = config
}

// GetConfig returns the current AI configuration
func (a *AIService) GetConfig() AIConfig {
	return a.config
}

// GenerateCommitMessage generates a commit message based on git diff
func (a *AIService) GenerateCommitMessage(diff string) (string, error) {
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("diff is empty")
	}

	if a.config.APIKey == "" && a.config.Provider != ProviderOllama {
		return "", fmt.Errorf("API key is required for %s", a.config.Provider)
	}

	switch a.config.Provider {
	case ProviderOpenAI:
		return a.generateWithOpenAI(diff)
	case ProviderClaude:
		return a.generateWithClaude(diff)
	case ProviderOllama:
		return a.generateWithOllama(diff)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", a.config.Provider)
	}
}

// generateWithOpenAI generates commit message using OpenAI API
func (a *AIService) generateWithOpenAI(diff string) (string, error) {
	baseURL := a.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	requestBody := map[string]interface{}{
		"model": a.getModel(),
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You are a helpful assistant that generates concise and clear git commit messages following the Conventional Commits specification.

Analyze the git diff and generate a commit message that:
1. Starts with a type (feat, fix, docs, style, refactor, test, chore, etc.)
2. Followed by a short description (max 72 characters)
3. Optionally includes a more detailed body paragraph
4. Uses imperative mood ("add" not "added" or "adds")
5. Is clear and specific about what changed

Return ONLY the commit message, no explanations or additional text.`,
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Generate a commit message for this diff:\n\n%s", diff),
			},
		},
		"temperature": 0.3,
		"max_tokens":  200,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content, _ := message["content"].(string)

	return strings.TrimSpace(content), nil
}

// generateWithClaude generates commit message using Claude API
func (a *AIService) generateWithClaude(diff string) (string, error) {
	baseURL := a.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	requestBody := map[string]interface{}{
		"model": a.getModel(),
		"max_tokens": 200,
		"system": `You are a helpful assistant that generates concise and clear git commit messages following the Conventional Commits specification.

Analyze the git diff and generate a commit message that:
1. Starts with a type (feat, fix, docs, style, refactor, test, chore, etc.)
2. Followed by a short description (max 72 characters)
3. Optionally includes a more detailed body paragraph
4. Uses imperative mood ("add" not "added" or "adds")
5. Is clear and specific about what changed

Return ONLY the commit message, no explanations or additional text.`,
		"messages": []map[string]string{
			{
				"role": "user",
				"content": fmt.Sprintf("Generate a commit message for this diff:\n\n%s", diff),
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	content, ok := response["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	text := content[0].(map[string]interface{})["text"].(string)
	return strings.TrimSpace(text), nil
}

// generateWithOllama generates commit message using local Ollama
func (a *AIService) generateWithOllama(diff string) (string, error) {
	baseURL := a.config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model := a.getModel()
	if model == "" {
		model = "llama2"
	}

	requestBody := map[string]interface{}{
		"model": model,
		"prompt": fmt.Sprintf(`You are a helpful assistant that generates concise and clear git commit messages following the Conventional Commits specification.

Generate a commit message for this diff. Return ONLY the commit message, no explanations.

Diff:
%s

Commit message:`, diff),
		"stream": false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	content, ok := response["response"].(string)
	if !ok {
		return "", fmt.Errorf("no response in output")
	}

	return strings.TrimSpace(content), nil
}

// getModel returns the model to use, with defaults for each provider
func (a *AIService) getModel() string {
	if a.config.Model != "" {
		return a.config.Model
	}

	switch a.config.Provider {
	case ProviderOpenAI:
		return "gpt-4"
	case ProviderClaude:
		return "claude-3-sonnet-20240229"
	case ProviderOllama:
		return "llama2"
	default:
		return "gpt-4"
	}
}

// ValidateConfig checks if the current configuration is valid
func (a *AIService) ValidateConfig() error {
	switch a.config.Provider {
	case ProviderOpenAI, ProviderClaude:
		if a.config.APIKey == "" {
			return fmt.Errorf("API key is required for %s", a.config.Provider)
		}
	case ProviderOllama:
		// Ollama doesn't require API key
	}

	if a.config.Provider == "" {
		return fmt.Errorf("provider must be specified")
	}

	return nil
}

// ValidateConfigParam validates the given AI configuration without modifying internal state
func (a *AIService) ValidateConfigParam(config AIConfig) error {
	switch config.Provider {
	case ProviderOpenAI, ProviderClaude:
		if config.APIKey == "" {
			return fmt.Errorf("API key is required for %s", config.Provider)
		}
	case ProviderOllama:
		// Ollama doesn't require API key
	}

	if config.Provider == "" {
		return fmt.Errorf("provider must be specified")
	}

	return nil
}
