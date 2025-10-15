package settings

import (
	"ChatGIT/internal/logger"
	"ChatGIT/internal/models"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type service struct {
	logger *logger.Logger
}

// New creates a new SettingsService instance
func New(l *logger.Logger) Service {
	return &service{
		logger: l,
	}
}

// GetChatConfig retrieves chat configuration from environment variables
func (s *service) GetChatConfig(ctx context.Context) (*models.ChatConfig, error) {
	s.logger.Debug("Loading chat configuration from environment")
	
	config := &models.ChatConfig{
		Provider:    s.getEnv("CHAT_PROVIDER", "openai"),
		Model:       s.getEnv("CHAT_MODEL", "gpt-4"),
		APIKey:      s.getEnv("OPENAI_API_KEY", ""), // Use standard OpenAI env var
		BaseURL:     s.getEnv("CHAT_BASE_URL", ""),
		MaxTokens:   s.getEnvInt("CHAT_MAX_TOKENS", 4000),
	}
	
	s.logger.Debug("Loaded chat config: provider=%s, model=%s", config.Provider, config.Model)
	return config, nil
}

// SetChatConfig updates chat configuration (mainly for validation, since we use env vars)
func (s *service) SetChatConfig(ctx context.Context, config *models.ChatConfig) error {
	s.logger.Info("Setting chat configuration: provider=%s, model=%s", config.Provider, config.Model)
	
	// Validate the config
	if err := s.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Note: We don't actually set environment variables here
	// In production, you might want to update a config file or use a different storage method
	s.logger.Info("Chat configuration validated successfully")
	return nil
}

// GetSettings retrieves general application settings
func (s *service) GetSettings(ctx context.Context) (*models.Settings, error) {
	s.logger.Debug("Loading general settings")
	
	settings := &models.Settings{
		Provider:    s.getEnv("CHAT_PROVIDER", "openai"),
		Model:       s.getEnv("CHAT_MODEL", "gpt-4"),
		MaxTokens:   s.getEnvInt("CHAT_MAX_TOKENS", 4000),
		Temperature: s.getEnvFloat("CHAT_TEMPERATURE", 0.7),
	}
	
	s.logger.Debug("Loaded settings: provider=%s, model=%s, max_tokens=%d", 
		settings.Provider, settings.Model, settings.MaxTokens)
	return settings, nil
}

// SetSettings updates general application settings
func (s *service) SetSettings(ctx context.Context, settings *models.Settings) error {
	s.logger.Info("Setting general settings")
	
	// Validate settings
	if settings.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	
	if settings.Model == "" {
		return fmt.Errorf("model is required")
	}
	
	if settings.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}
	
	if settings.Temperature < 0 || settings.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	
	// Note: We don't actually set environment variables here
	// In production, you might want to update a config file
	s.logger.Info("General settings validated successfully")
	return nil
}

// ListAvailableProviders returns list of supported chat providers
func (s *service) ListAvailableProviders() []string {
	return []string{
		"openai",
		"anthropic", 
		"local",
	}
}

// ListAvailableModels returns list of available models for a provider
func (s *service) ListAvailableModels(provider string) []string {
	switch strings.ToLower(provider) {
	case "openai":
		return []string{
			"gpt-4",
			"gpt-4-turbo",
			"gpt-3.5-turbo",
			"gpt-3.5-turbo-16k",
		}
	case "anthropic":
		return []string{
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
		}
	case "local":
		return []string{
			"llama-2-7b",
			"llama-2-13b",
			"llama-2-70b",
			"mistral-7b",
		}
	default:
		return []string{}
	}
}

// ValidateConfig validates chat configuration
func (s *service) ValidateConfig(config *models.ChatConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// Validate provider
	providers := s.ListAvailableProviders()
	providerValid := false
	for _, provider := range providers {
		if config.Provider == provider {
			providerValid = true
			break
		}
	}
	if !providerValid {
		return fmt.Errorf("invalid provider '%s', must be one of: %v", config.Provider, providers)
	}
	
	// Validate model
	models := s.ListAvailableModels(config.Provider)
	modelValid := false
	for _, model := range models {
		if config.Model == model {
			modelValid = true
			break
		}
	}
	if !modelValid {
		return fmt.Errorf("invalid model '%s' for provider '%s', must be one of: %v", 
			config.Model, config.Provider, models)
	}
	
	// Validate API key for cloud providers
	if config.Provider != "local" && config.APIKey == "" {
		// Check if API key is in environment
		envKey := ""
		switch config.Provider {
		case "openai":
			envKey = "OPENAI_API_KEY"
		case "anthropic":
			envKey = "ANTHROPIC_API_KEY"
		}
		
		if envKey != "" && os.Getenv(envKey) == "" {
			return fmt.Errorf("API key is required for provider '%s'. Set %s environment variable.", 
				config.Provider, envKey)
		}
	}
	
	// Validate max tokens
	if config.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}
	if config.MaxTokens > 32000 {
		return fmt.Errorf("max_tokens cannot exceed 32000")
	}
	
	return nil
}

// Helper functions for environment variable parsing
func (s *service) getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (s *service) getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (s *service) getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// GetEnvVarDoc returns documentation for available environment variables
func (s *service) GetEnvVarDoc() string {
	return `
ChatGIT Environment Variables:

Core Settings:
  LOG_LEVEL           - Log level (debug, info, warn, error, fatal) [default: info]
  LOG_FORMAT          - Log format (text, json) [default: text]
  SESSION_PATH        - Session storage path [default: ~/.chatgit/sessions]
  REPOS_PATH          - Repositories metadata path [default: ~/.chatgit/repos]

Chat Configuration:
  CHAT_PROVIDER       - Chat provider (openai, anthropic, local) [default: openai]
  CHAT_MODEL          - Chat model [default: gpt-4 for openai]
  CHAT_BASE_URL       - Custom base URL for chat API
  CHAT_MAX_TOKENS     - Maximum tokens for chat responses [default: 4000]
  CHAT_TEMPERATURE    - Chat temperature (0.0-2.0) [default: 0.7]

Provider-specific:
  OPENAI_API_KEY      - OpenAI API key
  ANTHROPIC_API_KEY   - Anthropic API key

Examples:
  export CHAT_PROVIDER=openai
  export CHAT_MODEL=gpt-4-turbo
  export OPENAI_API_KEY=your_api_key_here
  export CHAT_MAX_TOKENS=8000
`
}
