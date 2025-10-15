package session

import (
	"ChatGIT/internal/models"
	"context"
)

// Service provides session persistence
type Service interface {
	// UI state management
	SaveUIState(ctx context.Context, state *models.UIState) error
	GetUIState(ctx context.Context) (*models.UIState, error)
	
	// Repo metadata
	SaveRepoMetadata(ctx context.Context, metadata *models.RepoMetadata) error
	GetRepoMetadata(ctx context.Context, path string) (*models.RepoMetadata, error)
	ListRecentRepos(ctx context.Context, limit int) ([]*models.RepoMetadata, error)
	
	// Session cleanup
	Cleanup(ctx context.Context) error
}
