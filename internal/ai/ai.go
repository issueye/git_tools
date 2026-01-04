package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"git-ai-tools/internal/models"
)

// AIService handles AI operations for generating commit messages
type AIService struct {
	config models.AIConfig
	client *http.Client
}

// NewAIService creates a new AIService instance
func NewAIService() *AIService {
	return &AIService{
		client: &http.Client{},
		config: models.AIConfig{
			Provider: models.ProviderOpenAI,
			BaseURL:  "https://api.openai.com/v1",
			Model:    "gpt-4",
		},
	}
}

// SetConfig updates the AI service configuration
func (a *AIService) SetConfig(config models.AIConfig) {
	a.config = config
}

// GetConfig returns the current AI configuration
func (a *AIService) GetConfig() models.AIConfig {
	return a.config
}

// GenerateCommitMessage generates a commit message based on git diff
func (a *AIService) GenerateCommitMessage(diff string) (string, error) {
	if strings.TrimSpace(diff) == "" {
		return "", fmt.Errorf("diff is empty")
	}

	if a.config.APIKey == "" && a.config.Provider != models.ProviderOllama {
		return "", fmt.Errorf("API key is required for %s", a.config.Provider)
	}

	switch a.config.Provider {
	case models.ProviderOpenAI:
		return a.generateWithOpenAI(diff)
	case models.ProviderClaude:
		return a.generateWithClaude(diff)
	case models.ProviderOllama:
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
				"content": `你是一个专业的 git 提交信息助手，擅长生成简洁清晰的提交信息，遵循 Conventional Commits 规范。

分析 git diff 并生成提交信息，要求：
1. 使用中文编写提交信息
2. 以类型开头（feat, fix, docs, style, refactor, test, chore 等）
3. 后面跟简短的描述（不超过 50 字）
4. 如有必要，添加更详细的正文说明
5. 使用祈使句（用"添加"而非"已添加"）
6. 明确具体地说明变更内容

只返回提交信息本身，不要有其他解释。`,
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("请为以下 diff 生成一个中文的 git 提交信息：\n\n%s", diff),
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
		"model":     a.getModel(),
		"max_tokens": 200,
		"system": `你是一个专业的 git 提交信息助手，擅长生成简洁清晰的提交信息，遵循 Conventional Commits 规范。

分析 git diff 并生成提交信息，要求：
1. 使用中文编写提交信息
2. 以类型开头（feat, fix, docs, style, refactor, test, chore 等）
3. 后面跟简短的描述（不超过 50 字）
4. 如有必要，添加更详细的正文说明
5. 使用祈使句（用"添加"而非"已添加"）
6. 明确具体地说明变更内容

只返回提交信息本身，不要有其他解释。`,
		"messages": []map[string]string{
			{
				"role": "user",
				"content": fmt.Sprintf("请为以下 diff 生成一个中文的 git 提交信息：\n\n%s", diff),
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
		"prompt": fmt.Sprintf(`你是一个专业的 git 提交信息助手，擅长生成简洁清晰的提交信息，遵循 Conventional Commits 规范。

分析 git diff 并生成中文提交信息。要求：
1. 以类型开头（feat, fix, docs, style, refactor, test, chore 等）
2. 后面跟简短的描述（不超过 50 字）
3. 使用祈使句（用"添加"而非"已添加"）
4. 只返回提交信息本身，不要有其他解释

Diff:
%s

提交信息：`, diff),
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

	respContent, ok := response["response"].(string)
	if !ok {
		return "", fmt.Errorf("no response in output")
	}

	return strings.TrimSpace(respContent), nil
}

// getModel returns the model to use, with defaults for each provider
func (a *AIService) getModel() string {
	if a.config.Model != "" {
		return a.config.Model
	}

	switch a.config.Provider {
	case models.ProviderOpenAI:
		return "gpt-4"
	case models.ProviderClaude:
		return "claude-3-sonnet-20240229"
	case models.ProviderOllama:
		return "llama2"
	default:
		return "gpt-4"
	}
}

// ValidateConfig checks if the current configuration is valid
func (a *AIService) ValidateConfig() error {
	switch a.config.Provider {
	case models.ProviderOpenAI, models.ProviderClaude:
		if a.config.APIKey == "" {
			return fmt.Errorf("API key is required for %s", a.config.Provider)
		}
	case models.ProviderOllama:
		// Ollama doesn't require API key
	}

	if a.config.Provider == "" {
		return fmt.Errorf("provider must be specified")
	}

	return nil
}

// ValidateConfigParam validates the given AI configuration without modifying internal state
func (a *AIService) ValidateConfigParam(config models.AIConfig) error {
	switch config.Provider {
	case models.ProviderOpenAI, models.ProviderClaude:
		if config.APIKey == "" {
			return fmt.Errorf("API key is required for %s", config.Provider)
		}
	case models.ProviderOllama:
		// Ollama doesn't require API key
	}

	if config.Provider == "" {
		return fmt.Errorf("provider must be specified")
	}

	return nil
}
