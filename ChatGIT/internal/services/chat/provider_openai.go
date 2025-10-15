package chat

import (
	"ChatGIT/internal/logger"
	"ChatGIT/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIProvider implements Provider interface for OpenAI API
type OpenAIProvider struct {
	logger   *logger.Logger
	client   *http.Client
	apiKey   string
	baseURL  string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(l *logger.Logger, apiKey, baseURL string) Provider {
	return &OpenAIProvider{
		logger:  l,
		client:  &http.Client{Timeout: 30 * time.Second},
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// StreamChat streams chat completion from OpenAI
func (p *OpenAIProvider) StreamChat(ctx context.Context, messages []models.Message, config *models.ChatConfig) (<-chan models.ChatStreamToken, error) {
	tokenChan := make(chan models.ChatStreamToken, 10)
	
	if config.APIKey == "" {
		close(tokenChan)
		return tokenChan, fmt.Errorf("API key is required")
	}
	
	// Convert messages to OpenAI format
	openaiMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}
	
	// Prepare request
	requestBody := map[string]interface{}{
		"model":    config.Model,
		"messages": openaiMessages,
		"stream":   true,
		"max_tokens": config.MaxTokens,
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		close(tokenChan)
		return tokenChan, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	url := p.getAPIEndpoint("/chat/completions")
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		close(tokenChan)
		return tokenChan, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	
	// Make request in goroutine
	go func() {
		defer close(tokenChan)
		
		resp, err := p.client.Do(req)
		if err != nil {
			tokenChan <- models.ChatStreamToken{
				Error: fmt.Sprintf("Request failed: %v", err),
				IsEOF: true,
			}
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			tokenChan <- models.ChatStreamToken{
				Error: fmt.Sprintf("API error: %d - %s", resp.StatusCode, string(body)),
				IsEOF: true,
			}
			return
		}
		
		// Stream response
		
		for {
			// Read line by line for SSE format
			line, err := readLine(resp.Body)
			if err != nil {
				if err == io.EOF {
					tokenChan <- models.ChatStreamToken{IsEOF: true}
				} else {
					tokenChan <- models.ChatStreamToken{
						Error: fmt.Sprintf("Read error: %v", err),
						IsEOF: true,
					}
				}
				return
			}
			
			// Skip empty lines and "data: [DONE]" messages
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				tokenChan <- models.ChatStreamToken{IsEOF: true}
				return
			}
			
			// Parse JSON
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(data), &response); err != nil {
				p.logger.Warn("Failed to parse SSE data: %v", err)
				continue
			}
			
			// Extract token
			if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							tokenChan <- models.ChatStreamToken{
								Token: content,
								IsEOF: false,
							}
						}
					}
				}
			}
		}
	}()
	
	return tokenChan, nil
}

// GeneratePatch generates a code patch using OpenAI
func (p *OpenAIProvider) GeneratePatch(ctx context.Context, context, request string) (string, error) {
	if context == "" {
		return "", fmt.Errorf("context is required for patch generation")
	}
	
	// Build prompt for patch generation
	prompt := fmt.Sprintf(`
Given the following context and request, generate a unified diff patch:

Context:
%s

Request:
%s

Generate only the unified diff patch in the format:
--- a/file
+++ b/file
@@ -line,count +line,count @@
 ... content ...

Do not include any explanations or additional text.
`, context, request)
	
	// Create request for non-streaming completion
	requestBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream": false,
		"max_tokens": 2000,
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Make HTTP request
	url := p.getAPIEndpoint("/chat/completions")
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Extract content
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return strings.TrimSpace(content), nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("no content in response")
}

// getAPIEndpoint returns the full API endpoint
func (p *OpenAIProvider) getAPIEndpoint(path string) string {
	if p.baseURL != "" {
		return p.baseURL + path
	}
	return "https://api.openai.com/v1" + path
}

// readLine reads a line from the response body
func readLine(r io.Reader) (string, error) {
	var line []byte
	buf := make([]byte, 1)
	
	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF && len(line) > 0 {
				return string(line), nil
			}
			return "", err
		}
		
		if n == 0 {
			continue
		}
		
		if buf[0] == '\n' {
			return string(line), nil
		}
		
		if buf[0] != '\r' {
			line = append(line, buf[0])
		}
	}
}
