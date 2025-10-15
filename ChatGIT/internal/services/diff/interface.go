package diff

import (
	"ChatGIT/internal/models"
	"context"
)

// Service provides diff operations
type Service interface {
	// Repository setup
	SetRepository(repo interface{})
	
	// Diff operations
	DiffCommit(ctx context.Context, commitSHA string) (*models.Diff, error)
	DiffRange(ctx context.Context, base, head string) (*models.Diff, error)
	DiffWorkingTree(ctx context.Context) (*models.Diff, error)
	
	// Diff utilities
	ParseHunks(diffContent string) ([]models.Hunk, error)
	GetModifiedFiles(diffContent string) ([]string, error)
}
