package chat

import (
	"ChatGIT/internal/logger"
	"ChatGIT/internal/models"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type service struct {
	logger   *logger.Logger
	provider Provider
	config   *models.ChatConfig
}

// New creates a new ChatService instance
func New(l *logger.Logger, provider Provider) Service {
	return &service{
		logger:   l,
		provider: provider,
		config:   getDefaultConfig(),
	}
}

// StartMessage initiates a chat message and returns a streaming channel
func (s *service) StartMessage(ctx context.Context, role, text string) (<-chan models.ChatStreamToken, error) {
	s.logger.Info("Starting message: role=%s, length=%d", role, len(text))
	
	// Create a user message
	userMessage := models.Message{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   text,
		Timestamp: time.Now().Unix(),
	}
	
	// Start streaming response
	tokenChan, err := s.provider.StreamChat(ctx, []models.Message{userMessage}, s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to start chat stream: %w", err)
	}
	
	return tokenChan, nil
}

// OnUserMessage handles a user message
func (s *service) OnUserMessage(ctx context.Context, message string) error {
	s.logger.Info("Handling user message: length=%d", len(message))
	
	// This would create a USER: commit via GitService
	// For now, this is a placeholder
	s.logger.Debug("Creating USER commit for message")
	
	return nil
}

// OnAgentFinal handles the final agent response
func (s *service) OnAgentFinal(ctx context.Context, message string) error {
	s.logger.Info("Handling agent final message: length=%d", len(message))
	
	// This would create an AGENT: commit via GitService
	// For now, this is a placeholder
	s.logger.Debug("Creating AGENT commit for message")
	
	return nil
}

// ProposePatch generates a patch proposal
func (s *service) ProposePatch(ctx context.Context, description string) (*models.PatchProposal, error) {
	s.logger.Info("Generating patch proposal: %s", description)
	
	// Generate patch using provider
	patchContent, err := s.provider.GeneratePatch(ctx, "", description)
	if err != nil {
		return nil, fmt.Errorf("failed to generate patch: %w", err)
	}
	
	// Create proposal
	proposal := &models.PatchProposal{
		ID:          uuid.New().String(),
		Description: description,
		Diff:        patchContent,
		Files:       []string{}, // TODO: parse files from diff
		Approved:    false,
	}
	
	s.logger.Info("Generated patch proposal: %s", proposal.ID)
	return proposal, nil
}

// ApplyProposedPatch applies an approved patch
func (s *service) ApplyProposedPatch(ctx context.Context, proposalID string) error {
	s.logger.Info("Applying approved patch: %s", proposalID)
	
	// This would apply the patch via GitService
	// For now, this is a placeholder
	s.logger.Debug("Applying patch via GitService")
	
	return nil
}

// GetSession retrieves a chat session
func (s *service) GetSession(ctx context.Context, sessionID string) (*models.ChatSession, error) {
	s.logger.Debug("Getting session: %s", sessionID)
	
	// This would retrieve from session store
	// For now, return empty session
	session := &models.ChatSession{
		ID:       sessionID,
		RepoPath: "",
		Messages: []models.Message{},
		Created:  time.Now().Unix(),
		Updated:  time.Now().Unix(),
	}
	
	return session, nil
}

// CreateSession creates a new chat session
func (s *service) CreateSession(ctx context.Context, repoPath string) (*models.ChatSession, error) {
	s.logger.Info("Creating session for repo: %s", repoPath)
	
	session := &models.ChatSession{
		ID:       uuid.New().String(),
		RepoPath: repoPath,
		Messages: []models.Message{},
		Created:  time.Now().Unix(),
		Updated:  time.Now().Unix(),
	}
	
	// This would save to session store
	s.logger.Info("Created session: %s", session.ID)
	return session, nil
}

// GetConfig returns current chat configuration
func (s *service) GetConfig() *models.ChatConfig {
	return s.config
}

// UpdateConfig updates chat configuration
func (s *service) UpdateConfig(config *models.ChatConfig) error {
	s.logger.Info("Updating chat config: provider=%s, model=%s", config.Provider, config.Model)
	
	// Validate config
	if err := s.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	
	s.config = config
	s.logger.Info("Updated chat config successfully")
	return nil
}

// validateConfig validates chat configuration
func (s *service) validateConfig(config *models.ChatConfig) error {
	if config.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	
	if config.Model == "" {
		return fmt.Errorf("model is required")
	}
	
	// Check if provider is supported
	supportedProviders := []string{"openai", "anthropic", "local"}
	supported := false
	for _, provider := range supportedProviders {
		if config.Provider == provider {
			supported = true
			break
		}
	}
	
	if !supported {
		return fmt.Errorf("unsupported provider: %s", config.Provider)
	}
	
	return nil
}

// getDefaultConfig returns default chat configuration
func getDefaultConfig() *models.ChatConfig {
	return &models.ChatConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		APIKey:      "",
		BaseURL:     "",
		MaxTokens:   4000,
	}
}

// estimateTokens estimates token count for a message (simplified)
func (s *service) estimateTokens(text string) int {
	// Simple estimation: roughly 4 characters per token
	// In production, use a proper tokenizer
	return len(text) / 4
}

// formatMessage formats a message for the provider
func (s *service) formatMessage(msg models.Message) string {
	return fmt.Sprintf("%s: %s", strings.ToUpper(msg.Role), msg.Content)
}
