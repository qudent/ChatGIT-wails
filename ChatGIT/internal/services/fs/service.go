package fs

import (
	"ChatGIT/internal/logger"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type service struct {
	repoRoot string
	logger   *logger.Logger
	
	// Allowed file extensions for writes
	allowedExtensions map[string]bool
}

// New creates a new FSService instance
func New(l *logger.Logger) Service {
	allowed := make(map[string]bool)
	extensions := []string{
		".txt", ".md", ".go", ".js", ".ts", ".tsx", ".jsx", 
		".json", ".yaml", ".yml", ".xml", ".html", ".css",
		".py", ".rs", ".c", ".cpp", ".h", ".hpp",
		".sh", ".bat", ".ps1", ".dockerfile", ".gitignore",
	}
	
	for _, ext := range extensions {
		allowed[ext] = true
	}
	
	return &service{
		logger:           l,
		allowedExtensions: allowed,
	}
}

// SafeRead safely reads a file within the repository
func (s *service) SafeRead(ctx context.Context, path string) ([]byte, error) {
	s.logger.Debug("SafeRead: %s", path)
	
	// Validate path
	if err := s.ValidatePath(path); err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}
	
	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", path)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	
	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("path is not a regular file: %s", path)
	}
	
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	s.logger.Debug("Successfully read %d bytes from %s", len(data), path)
	return data, nil
}

// SafeWrite safely writes data to a file within the repository
func (s *service) SafeWrite(ctx context.Context, path string, data []byte) error {
	s.logger.Debug("SafeWrite: %s (%d bytes)", path, len(data))
	
	// Validate path
	if err := s.ValidatePath(path); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}
	
	// Check file extension for writes
	ext := strings.ToLower(filepath.Ext(path))
	if !s.allowedExtensions[ext] {
		return fmt.Errorf("file extension %s is not allowed for writes", ext)
	}
	
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := s.CreateDir(ctx, dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write file
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	s.logger.Debug("Successfully wrote %d bytes to %s", len(data), path)
	return nil
}

// Exists checks if a path exists
func (s *service) Exists(ctx context.Context, path string) (bool, error) {
	// Validate path
	if err := s.ValidatePath(path); err != nil {
		return false, fmt.Errorf("path validation failed: %w", err)
	}
	
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat path: %w", err)
	}
	
	return true, nil
}

// ListFiles lists files in a directory within the repository
func (s *service) ListFiles(ctx context.Context, path string) ([]string, error) {
	s.logger.Debug("ListFiles: %s", path)
	
	// Validate path
	if err := s.ValidatePath(path); err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}
	
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // Empty directory
		}
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}
	
	// Check if it's a directory
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}
	
	// Read directory
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	
	var files []string
	for _, entry := range entries {
		// Skip hidden files and directories starting with .
		if strings.HasPrefix(entry.Name(), ".") && entry.Name() != ".gitignore" {
			continue
		}
		
		fullPath := filepath.Join(path, entry.Name())
		
		// Validate full path (in case of symlinks)
		if err := s.ValidatePath(fullPath); err != nil {
			s.logger.Warn("Skipping invalid path: %s", fullPath)
			continue
		}
		
		files = append(files, entry.Name())
	}
	
	s.logger.Debug("Listed %d files in %s", len(files), path)
	return files, nil
}

// CreateDir creates a directory within the repository
func (s *service) CreateDir(ctx context.Context, path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	
	s.logger.Debug("CreateDir: %s", path)
	
	// Validate path
	if err := s.ValidatePath(path); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}
	
	// Check if directory already exists
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return nil // Directory already exists
		}
		return fmt.Errorf("path exists but is not a directory: %s", path)
	}
	
	// Create directory with proper permissions
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	s.logger.Debug("Successfully created directory: %s", path)
	return nil
}

// ValidatePath validates that a path is safe to access
func (s *service) ValidatePath(path string) error {
	if s.repoRoot == "" {
		return fmt.Errorf("repository root not set")
	}
	
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	
	// Clean path to remove any .. or . elements
	cleanPath := filepath.Clean(path)
	
	// Make path absolute
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Make sure repoRoot is absolute too
	absRepoRoot, err := filepath.Abs(s.repoRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute repo root: %w", err)
	}
	
	// Check if path is within repository root
	if !strings.HasPrefix(absPath, absRepoRoot) {
		return fmt.Errorf("path %s is outside repository root %s", absPath, absRepoRoot)
	}
	
	// Check for symlinks that might escape the repository
	if err := s.validateSymlinkSafety(absPath, absRepoRoot); err != nil {
		return fmt.Errorf("symlink validation failed: %w", err)
	}
	
	return nil
}

// validateSymlinkSafety checks that symlinks don't escape the repository
func (s *service) validateSymlinkSafety(path, repoRoot string) error {
	// Walk up the path checking each component
	current := path
	for {
		// Get directory
		dir := filepath.Dir(current)
		if dir == current {
			break // Reached root
		}
		
		// Stat the path to check if it's a symlink
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				current = dir
				continue
			}
			return fmt.Errorf("failed to stat path component %s: %w", current, err)
		}
		
		// If it's a symlink, resolve it and check
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(current)
			if err != nil {
				return fmt.Errorf("failed to read symlink %s: %w", current, err)
			}
			
			// Make target absolute
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(current), target)
			}
			
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("failed to get absolute symlink target: %w", err)
			}
			
			// Check if symlink target is outside repo
			if !strings.HasPrefix(absTarget, repoRoot) {
				return fmt.Errorf("symlink %s -> %s points outside repository", current, absTarget)
			}
			
			// Continue validation from the target
			return s.validateSymlinkSafety(absTarget, repoRoot)
		}
		
		current = dir
		
		// If we've reached the repo root, we're done
		if current == repoRoot {
			break
		}
	}
	
	return nil
}

// GetRepoRoot returns the current repository root
func (s *service) GetRepoRoot() string {
	return s.repoRoot
}

// SetRepoRoot sets the repository root for path validation
func (s *service) SetRepoRoot(path string) {
	s.logger.Info("Setting repository root: %s", path)
	
	// Validate path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		s.logger.Error("Repository root does not exist: %s", path)
		return
	}
	
	// Make path absolute and clean
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		s.logger.Error("Failed to get absolute path: %v", err)
		return
	}
	
	s.repoRoot = absPath
	s.logger.Info("Repository root set to: %s", s.repoRoot)
}

// isExecutable checks if a file is executable on Unix systems
func (s *service) isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	
	// Check executable bit
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return stat.Mode&0111 != 0
	}
	
	return false
}
