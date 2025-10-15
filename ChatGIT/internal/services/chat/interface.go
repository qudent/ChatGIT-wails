package chat

import (
	"ChatGIT/internal/models"
	"context"
)

// Service provides chat functionality
type Service interface {
	// Chat operations
	StartMessage(ctx context.Context, role, text string) (<-chan models.ChatStreamToken, error)
	OnUserMessage(ctx context.Context, message string) error
	OnAgentFinal(ctx context.Context, message string) error
	
	// Patch operations
	ProposePatch(ctx context.Context, description string) (*models.PatchProposal, error)
	ApplyProposedPatch(ctx context.Context, proposalID string) error
	
	// Session management
	GetSession(ctx context.Context, sessionID string) (*models.ChatSession, error)
	CreateSession(ctx context.Context, repoPath string) (*models.ChatSession, error)
	
	// Configuration
	GetConfig() *models.ChatConfig
	UpdateConfig(config *models.ChatConfig) error
}

// Provider represents an AI chat provider interface
type Provider interface {
	StreamChat(ctx context.Context, messages []models.Message, config *models.ChatConfig) (<-chan models.ChatStreamToken, error)
	GeneratePatch(ctx context.Context, context string, request string) (string, error)
}
