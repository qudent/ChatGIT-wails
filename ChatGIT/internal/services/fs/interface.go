package fs

import (
	"context"
)

// Service provides file system operations with safety checks
type Service interface {
	// Basic file operations
	SafeRead(ctx context.Context, path string) ([]byte, error)
	SafeWrite(ctx context.Context, path string, data []byte) error
	Exists(ctx context.Context, path string) (bool, error)
	
	// Directory operations
	ListFiles(ctx context.Context, path string) ([]string, error)
	CreateDir(ctx context.Context, path string) error
	
	// Validation
	ValidatePath(path string) error
	GetRepoRoot() string
	SetRepoRoot(path string)
}
