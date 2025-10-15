package git

import (
	"ChatGIT/internal/models"
	"context"
)

// Service provides git operations
type Service interface {
	// Repository management
	SetRepo(path string) error
	GetRepo() string
	GetRepository() interface{} // Returns the underlying repository
	Health() error
	GitVersion() string
	
	// Status and basic info
	Status() (*models.Status, error)
	
	// Commit operations
	LogSpine(ctx context.Context, limit, skip int) (*models.CommitList, error)
	LogWithGraph(ctx context.Context, options LogOptions) (*models.CommitList, error)
	GetCommit(sha string) (*models.Commit, error)
	CreateEmptyCommit(role, subject, body string, footers map[string]string) (*models.Commit, error)
	
	// File operations
	ShowFile(commitSHA, path string) (string, error)
	ListTree(commitSHA, path string) ([]string, error)
	
	// Patch operations
	ApplyPatch(ctx context.Context, diff string, allowCreate bool) error
	AddPatchCommit(role, relatesToSHA, summary, body string) (*models.Commit, error)
	
	// Branch operations
	ListBranches() ([]string, error)
	CreateBranch(name string) error
	CheckoutBranch(name string) error
	ComputeMergeBase(a, b string) (string, error)
}

// LogOptions controls log query behavior
type LogOptions struct {
	Limit    int    `json:"limit"`
	Skip     int    `json:"skip"`
	Path     string `json:"path"`
	From     string `json:"from"`
	To       string `json:"to"`
	FirstParent bool `json:"firstParent"`
	IncludeMerges bool `json:"includeMerges"`
}
