package settings

import (
	"ChatGIT/internal/models"
	"context"
)

// Service provides settings management
type Service interface {
	// Chat settings
	GetChatConfig(ctx context.Context) (*models.ChatConfig, error)
	SetChatConfig(ctx context.Context, config *models.ChatConfig) error
	
	// Application settings
	GetSettings(ctx context.Context) (*models.Settings, error)
	SetSettings(ctx context.Context, settings *models.Settings) error
	
	// Environment management
	ListAvailableProviders() []string
	ListAvailableModels(provider string) []string
	ValidateConfig(config *models.ChatConfig) error
}
