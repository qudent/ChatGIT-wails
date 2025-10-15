package diff

import (
	"ChatGIT/internal/logger"
	"ChatGIT/internal/models"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type service struct {
	repo   *git.Repository
	logger *logger.Logger
}

// New creates a new DiffService instance
func New(l *logger.Logger) Service {
	return &service{
		logger: l,
	}
}

// SetRepository sets the git repository for diff operations
func (s *service) SetRepository(repo interface{}) {
	if gitRepo, ok := repo.(*git.Repository); ok {
		s.repo = gitRepo
	}
}

// DiffCommit creates a diff for a single commit
func (s *service) DiffCommit(ctx context.Context, commitSHA string) (*models.Diff, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository set")
	}
	
	s.logger.Debug("DiffCommit: %s", commitSHA)
	
	// Get the commit
	hash := plumbing.NewHash(commitSHA)
	commit, err := s.repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}
	
	// Get parent commit
	var parent *object.Commit
	if len(commit.ParentHashes) > 0 {
		parent, err = s.repo.CommitObject(commit.ParentHashes[0])
		if err != nil {
			return nil, fmt.Errorf("failed to get parent commit: %w", err)
		}
	}
	
	// Create diff
	diffResult, err := s.createDiff(parent, commit)
	if err != nil {
		return nil, fmt.Errorf("failed to create diff: %w", err)
	}
	
	return diffResult, nil
}

// DiffRange creates a diff between two commits
func (s *service) DiffRange(ctx context.Context, base, head string) (*models.Diff, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository set")
	}
	
	s.logger.Debug("DiffRange: %s..%s", base, head)
	
	// Get commits
	baseHash := plumbing.NewHash(base)
	headHash := plumbing.NewHash(head)
	
	baseCommit, err := s.repo.CommitObject(baseHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get base commit: %w", err)
	}
	
	headCommit, err := s.repo.CommitObject(headHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get head commit: %w", err)
	}
	
	// Create diff
	diffResult, err := s.createDiff(baseCommit, headCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to create diff: %w", err)
	}
	
	return diffResult, nil
}

// DiffWorkingTree creates a diff of working tree changes
func (s *service) DiffWorkingTree(ctx context.Context) (*models.Diff, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository set")
	}
	
	s.logger.Debug("DiffWorkingTree")
	
	// Get worktree
	worktree, err := s.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}
	
	// Get head reference
	head, err := s.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get head: %w", err)
	}
	
	// Get status to see what files have changed
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	
	// Generate diff content for each changed file
	var diffContent strings.Builder
	var files []string
	
	for file, fileStatus := range status {
		if fileStatus.Worktree == git.Unmodified {
			continue
		}
		
		files = append(files, file)
		
		// For now, just note the file change
		// Full diff generation would require more complex implementation
		diffContent.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", file, file))
		diffContent.WriteString(fmt.Sprintf("File status: Changed\n"))
		diffContent.WriteString("\n")
	}
	
	// Parse hunks from the diff content
	hunks, err := s.ParseHunks(diffContent.String())
	if err != nil {
		s.logger.Warn("Failed to parse hunks: %v", err)
		hunks = []models.Hunk{}
	}
	
	return &models.Diff{
		From:    head.Hash().String(),
		To:      "working-tree",
		Content: diffContent.String(),
		Hunks:   hunks,
		Files:   files,
	}, nil
}

// ParseHunks parses diff hunks from diff content
func (s *service) ParseHunks(diffContent string) ([]models.Hunk, error) {
	var hunks []models.Hunk
	
	// Regex to match hunk headers
	// Format: @@ -start,lines +start,lines @@
	hunkRegex := regexp.MustCompile(`@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	
	lines := strings.Split(diffContent, "\n")
	
	for _, line := range lines {
		if !strings.HasPrefix(line, "@@") || !strings.Contains(line, "@@") {
			continue
		}
		
		matches := hunkRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		
		// Parse line numbers
		oldStart := s.parseInt(matches[1])
		oldLines := s.parseInt(matches[2])
		newStart := s.parseInt(matches[3])
		newLines := s.parseInt(matches[4])
		
		// Default values if parsing failed
		if oldLines == 0 {
			oldLines = 1
		}
		if newLines == 0 {
			newLines = 1
		}
		
		hunks = append(hunks, models.Hunk{
			OldStart: oldStart,
			OldLines: oldLines,
			NewStart: newStart,
			NewLines: newLines,
			Content:  line,
		})
	}
	
	s.logger.Debug("Parsed %d hunks from diff", len(hunks))
	return hunks, nil
}

// GetModifiedFiles extracts modified file paths from diff content
func (s *service) GetModifiedFiles(diffContent string) ([]string, error) {
	var files []string
	
	// Simple regex to match file lines in diff
	// Look for lines starting with "diff --git" or "+++ " or "--- "
	lines := strings.Split(diffContent, "\n")
	
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			// Extract file from "diff --git a/file b/file"
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				// Remove "a/" prefix
				file := strings.TrimPrefix(parts[2], "a/")
				files = append(files, file)
			}
		} else if strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "--- ") {
			// Extract file from "+++ b/file" or "--- a/file"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				file := strings.TrimPrefix(parts[1], "a/")
				file = strings.TrimPrefix(file, "b/")
				
				// Avoid duplicates
				found := false
				for _, existing := range files {
					if existing == file {
						found = true
						break
					}
				}
				if !found {
					files = append(files, file)
				}
			}
		}
	}
	
	s.logger.Debug("Found %d modified files in diff", len(files))
	return files, nil
}

// createDiff creates a diff between two commits
func (s *service) createDiff(base, head *object.Commit) (*models.Diff, error) {
	var from, to string
	var files []string
	
	if base == nil {
		from = "null"
		to = head.Hash.String()
	} else {
		from = base.Hash.String()
		to = head.Hash.String()
	}
	
	// Get patches
	patch, err := base.Patch(head)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch: %w", err)
	}
	
	// Convert to string
	diffContent := patch.String()
	
	// Parse hunks
	hunks, err := s.ParseHunks(diffContent)
	if err != nil {
		s.logger.Warn("Failed to parse hunks: %v", err)
		hunks = []models.Hunk{}
	}
	
	// Get modified files
	files, err = s.GetModifiedFiles(diffContent)
	if err != nil {
		s.logger.Warn("Failed to get modified files: %v", err)
		files = []string{}
	}
	
	return &models.Diff{
		From:    from,
		To:      to,
		Content: diffContent,
		Hunks:   hunks,
		Files:   files,
	}, nil
}

// parseInt safely parses an integer from string
func (s *service) parseInt(str string) int {
	if str == "" {
		return 0
	}
	
	var result int
	fmt.Sscanf(str, "%d", &result)
	return result
}

// formatPatch creates a unified diff format patch
func (s *service) formatPatch(patch object.Patch) string {
	// This is a simplified implementation
	// In production, you'd want to properly format the patch
	return patch.String()
}

// getFileChanges extracts file changes from a patch
func (s *service) getFileChanges(patch object.Patch) []string {
	var files []string
	
	patch.FilePatches()
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		if from != nil {
			files = append(files, from.Path())
		}
		if to != nil {
			files = append(files, to.Path())
		}
	}
	
	// Remove duplicates
	unique := make(map[string]bool)
	var result []string
	for _, file := range files {
		if !unique[file] {
			unique[file] = true
			result = append(result, file)
		}
	}
	
	return result
}
