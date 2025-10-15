package session

import (
	"ChatGIT/internal/config"
	"ChatGIT/internal/logger"
	"ChatGIT/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type service struct {
	config    *config.Config
	logger    *logger.Logger
	sessionDir string
	mu        sync.RWMutex
	
	// In-memory cache
	uiState     *models.UIState
	repoMetadata map[string]*models.RepoMetadata
}

// New creates a new SessionService instance
func New(cfg *config.Config, l *logger.Logger) Service {
	sessionDir := cfg.GetSessionPath()
	
	// Ensure session directory exists
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		l.Error("Failed to create session directory: %v", err)
	} else {
		l.Debug("Session directory: %s", sessionDir)
	}
	
	return &service{
		config:       cfg,
		logger:       l,
		sessionDir:   sessionDir,
		repoMetadata: make(map[string]*models.RepoMetadata),
	}
}

// SaveUIState saves the UI state to disk
func (s *service) SaveUIState(ctx context.Context, state *models.UIState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.logger.Debug("Saving UI state")
	
	// Update cache
	s.uiState = state
	
	// Save to file
	filename := s.getUIStateFilename()
	return s.saveJSON(filename, state)
}

// GetUIState retrieves the UI state from disk
func (s *service) GetUIState(ctx context.Context) (*models.UIState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return from cache if available
	if s.uiState != nil {
		return s.uiState, nil
	}
	
	s.logger.Debug("Loading UI state from disk")
	
	// Load from file
	filename := s.getUIStateFilename()
	state := &models.UIState{
		OpenTabs:    []string{},
		SplitLayout: true,
		Theme:       "light",
	}
	
	if err := s.loadJSON(filename, state); err != nil {
		if !os.IsNotExist(err) {
			s.logger.Warn("Failed to load UI state: %v", err)
		}
		// Use default values
	}
	
	// Cache it
	s.uiState = state
	return state, nil
}

// SaveRepoMetadata saves repository metadata
func (s *service) SaveRepoMetadata(ctx context.Context, metadata *models.RepoMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.logger.Debug("Saving repo metadata: %s", metadata.Path)
	
	// Update cache
	s.repoMetadata[metadata.Path] = metadata
	metadata.LastUsed = time.Now().Unix()
	
	// Save to file
	filename := s.getRepoMetadataFilename(metadata.Path)
	return s.saveJSON(filename, metadata)
}

// GetRepoMetadata retrieves repository metadata
func (s *service) GetRepoMetadata(ctx context.Context, path string) (*models.RepoMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return from cache if available
	if metadata, exists := s.repoMetadata[path]; exists {
		return metadata, nil
	}
	
	s.logger.Debug("Loading repo metadata: %s", path)
	
	// Load from file
	filename := s.getRepoMetadataFilename(path)
	metadata := &models.RepoMetadata{
		Path:     path,
		Name:     filepath.Base(path),
		Tags:     []string{},
		Favorite: false,
	}
	
	if err := s.loadJSON(filename, metadata); err != nil {
		if !os.IsNotExist(err) {
			s.logger.Warn("Failed to load repo metadata: %v", err)
		}
	}
	
	// Cache it
	s.repoMetadata[path] = metadata
	return metadata, nil
}

// ListRecentRepos returns list of recent repositories
func (s *service) ListRecentRepos(ctx context.Context, limit int) ([]*models.RepoMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	s.logger.Debug("Listing recent repos, limit: %d", limit)
	
	// Ensure we have loaded all metadata
	if err := s.loadAllRepoMetadata(); err != nil {
		s.logger.Warn("Failed to load all repo metadata: %v", err)
	}
	
	// Convert to slice and sort by last used
	var repos []*models.RepoMetadata
	for _, metadata := range s.repoMetadata {
		repos = append(repos, metadata)
	}
	
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].LastUsed > repos[j].LastUsed
	})
	
	// Apply limit
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}
	
	s.logger.Debug("Found %d recent repos", len(repos))
	return repos, nil
}

// Cleanup removes old or invalid session data
func (s *service) Cleanup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.logger.Info("Starting session cleanup")
	
	// Get all files in session directory
	files, err := filepath.Glob(filepath.Join(s.sessionDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list session files: %w", err)
	}
	
	removed := 0
	cutoff := time.Now().AddDate(0, 0, -30) // 30 days ago
	
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		
		// Remove files older than 30 days
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				s.logger.Warn("Failed to remove old session file %s: %v", file, err)
			} else {
				s.logger.Debug("Removed old session file: %s", file)
				removed++
			}
		}
	}
	
	s.logger.Info("Cleanup completed, removed %d files", removed)
	return nil
}

// Helper methods

func (s *service) getUIStateFilename() string {
	return filepath.Join(s.sessionDir, "ui_state.json")
}

func (s *service) getRepoMetadataFilename(repoPath string) string {
	// Use a safe filename based on repo path
	hash := s.hashPath(repoPath)
	return filepath.Join(s.sessionDir, fmt.Sprintf("repo_%s.json", hash))
}

func (s *service) hashPath(path string) string {
	// Simple hash to create valid filename
	// In production, use proper hash function like SHA256
	result := ""
	for _, c := range path {
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'):
			result += string(c)
		case c == '/' || c == '\\':
			result += "_"
		default:
			result += fmt.Sprintf("%x", c)
		}
	}
	// Truncate if too long
	if len(result) > 50 {
		return result[:50]
	}
	return result
}

func (s *service) saveJSON(filename string, data interface{}) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal to JSON with pretty formatting
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	// Write to temporary file first
	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Clean up
		return fmt.Errorf("failed to rename file: %w", err)
	}
	
	return nil
}

func (s *service) loadJSON(filename string, data interface{}) error {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	
	if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return nil
}

func (s *service) loadAllRepoMetadata() error {
	// Get all repo metadata files
	files, err := filepath.Glob(filepath.Join(s.sessionDir, "repo_*.json"))
	if err != nil {
		return fmt.Errorf("failed to list repo files: %w", err)
	}
	
	for _, file := range files {
		var metadata models.RepoMetadata
		if err := s.loadJSON(file, &metadata); err != nil {
			s.logger.Warn("Failed to load repo metadata from %s: %v", file, err)
			continue
		}
		s.repoMetadata[metadata.Path] = &metadata
	}
	
	return nil
}

// GetReposByTag returns repositories filtered by tag
func (s *service) GetReposByTag(ctx context.Context, tag string) ([]*models.RepoMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if err := s.loadAllRepoMetadata(); err != nil {
		return nil, fmt.Errorf("failed to load repo metadata: %w", err)
	}
	
	var repos []*models.RepoMetadata
	for _, metadata := range s.repoMetadata {
		for _, repoTag := range metadata.Tags {
			if repoTag == tag {
				repos = append(repos, metadata)
				break
			}
		}
	}
	
	return repos, nil
}

// GetFavoriteRepos returns favorite repositories
func (s *service) GetFavoriteRepos(ctx context.Context) ([]*models.RepoMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if err := s.loadAllRepoMetadata(); err != nil {
		return nil, fmt.Errorf("failed to load repo metadata: %w", err)
	}
	
	var repos []*models.RepoMetadata
	for _, metadata := range s.repoMetadata {
		if metadata.Favorite {
			repos = append(repos, metadata)
		}
	}
	
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].LastUsed > repos[j].LastUsed
	})
	
	return repos, nil
}
