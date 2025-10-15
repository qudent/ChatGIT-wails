package git

import (
	"ChatGIT/internal/logger"
	"ChatGIT/internal/models"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/filemode"
)

type service struct {
	repo   *git.Repository
	repoPath string
	logger *logger.Logger
}

// New creates a new GitService instance
func New(l *logger.Logger) Service {
	return &service{
		logger: l,
	}
}

// SetRepo sets the git repository path and opens it
func (s *service) SetRepo(path string) error {
	s.logger.Info("Setting repository path: %s", path)
	
	// Validate path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("repository path does not exist: %s", path)
	}
	
	// Open repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}
	
	s.repo = repo
	s.repoPath = path
	
	s.logger.Info("Successfully opened repository: %s", path)
	return nil
}

// GetRepo returns the current repository path
func (s *service) GetRepo() string {
	return s.repoPath
}

// GetRepository returns the underlying git repository
func (s *service) GetRepository() interface{} {
	return s.repo
}

// Health checks if the repository is accessible
func (s *service) Health() error {
	if s.repo == nil {
		return fmt.Errorf("no repository set")
	}
	
	// Try to get a reference to verify repository is valid
	_, err := s.repo.Head()
	return err
}

// GitVersion returns git version info (using go-git)
func (s *service) GitVersion() string {
	return "go-git/v5.12.0"
}

// Status returns repository status
func (s *service) Status() (*models.Status, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository set")
	}
	
	// Get worktree
	worktree, err := s.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}
	
	// Get status
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	
	// Get head reference
	head, err := s.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get head: %w", err)
	}
	
	// Get branch name
	branch := head.Name().Short()
	if branch == "" {
		branch = "HEAD"
	}
	
	// Count changes
	staged := 0
	unstaged := 0
	untracked := 0
	
	for file, fileStatus := range status {
		s.logger.Debug("File status: %s = %v", file, fileStatus)
		
		if fileStatus.Staging != git.Unmodified {
			staged++
		}
		if fileStatus.Worktree != git.Unmodified {
			unstaged++
		}
		if fileStatus.Worktree == git.Untracked {
			untracked++
		}
	}
	
	isClean := status.IsClean()
	
	return &models.Status{
		Branch:    branch,
		Head:      head.Hash().String(),
		Ahead:     0, // TODO: implement ahead/behind calculation
		Behind:    0, // TODO: implement ahead/behind calculation
		Staged:    staged,
		Unstaged:  unstaged,
		Untracked: untracked,
		IsClean:   isClean,
	}, nil
}

// LogSpine gets the main commit spine (first-parent)
func (s *service) LogSpine(ctx context.Context, limit, skip int) (*models.CommitList, error) {
	options := LogOptions{
		Limit:      limit,
		Skip:       skip,
		FirstParent: true,
	}
	return s.LogWithGraph(ctx, options)
}

// LogWithGraph gets commit log with graph information
func (s *service) LogWithGraph(ctx context.Context, options LogOptions) (*models.CommitList, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository set")
	}
	
	// Get head reference
	head, err := s.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get head: %w", err)
	}
	
	// Get commit iterator
	var commitIter object.CommitIter
	if options.From != "" && options.To != "" {
		// Range query
		toHash := plumbing.NewHash(options.To)
		commitIter, err = s.repo.Log(&git.LogOptions{
			From:  toHash,
			Order: git.LogOrderCommitterTime,
		})
	} else {
		// Default log from HEAD
		commitIter, err = s.repo.Log(&git.LogOptions{
			From:  head.Hash(),
			Order: git.LogOrderCommitterTime,
		})
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get commit iterator: %w", err)
	}
	defer commitIter.Close()
	
	// Collect commits
	var commits []models.Commit
	total := 0
	count := 0
	
	err = commitIter.ForEach(func(c *object.Commit) error {
		total++
		
		// Skip commits if needed
		if count < options.Skip {
			return nil
		}
		
		// Limit reached
		if options.Limit > 0 && count >= options.Skip+options.Limit {
			return nil
		}
		
		// Skip merges if not included
		if !options.IncludeMerges && len(c.ParentHashes) > 1 {
			return nil
		}
		
		// Convert to model
		commit := s.convertCommit(c)
		commits = append(commits, commit)
		count++
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}
	
	hasMore := options.Limit > 0 && total > options.Skip+options.Limit
	
	return &models.CommitList{
		Commits: commits,
		Total:   total,
		HasMore: hasMore,
	}, nil
}

// GetCommit gets a specific commit by SHA
func (s *service) GetCommit(sha string) (*models.Commit, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository set")
	}
	
	hash := plumbing.NewHash(sha)
	commit, err := s.repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}
	
	modelCommit := s.convertCommit(commit)
	return &modelCommit, nil
}

// CreateEmptyCommit creates an empty commit with metadata
func (s *service) CreateEmptyCommit(role, subject, body string, footers map[string]string) (*models.Commit, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository set")
	}
	
	// Get worktree
	worktree, err := s.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}
	
	// Build commit message
	message := fmt.Sprintf("%s%s", role, subject)
	if body != "" {
		message += fmt.Sprintf("\n\n%s", body)
	}
	
	// Add footers
	for key, value := range footers {
		message += fmt.Sprintf("\n\n%s: %s", key, value)
	}
	
	// Create empty commit
	hash, err := worktree.Commit(message, &git.CommitOptions{
		AllowEmptyCommits: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}
	
	// Get the created commit
	return s.GetCommit(hash.String())
}

// ShowFile shows file content at a specific commit
func (s *service) ShowFile(commitSHA, path string) (string, error) {
	if s.repo == nil {
		return "", fmt.Errorf("no repository set")
	}
	
	hash := plumbing.NewHash(commitSHA)
	commit, err := s.repo.CommitObject(hash)
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}
	
	// Get tree
	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get tree: %w", err)
	}
	
	// Get file
	file, err := tree.File(path)
	if err != nil {
		return "", fmt.Errorf("failed to get file: %w", err)
	}
	
	contents, err := file.Contents()
	if err != nil {
		return "", fmt.Errorf("failed to get file contents: %w", err)
	}
	
	return contents, nil
}

// ListTree lists files/directories at a specific commit path
func (s *service) ListTree(commitSHA, path string) ([]string, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository set")
	}
	
	hash := plumbing.NewHash(commitSHA)
	commit, err := s.repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}
	
	// Get tree
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}
	
	// If path is specified, find subtree
	if path != "" && path != "." {
		entry, err := tree.FindEntry(path)
		if err != nil {
			return nil, fmt.Errorf("failed to find entry: %w", err)
		}
		
		if entry.Mode != filemode.Dir && entry.Mode != filemode.Submodule {
			return []string{path}, nil
		}
		
		// For now, return just the file name if we can't get subtree
		// In a full implementation, we'd need to use a different approach
		// tree.FindTree doesn't exist in go-git v5
		return []string{path}, nil
	}
	
	// List entries
	var names []string
	err = tree.Files().ForEach(func(f *object.File) error {
		names = append(names, f.Name)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	
	// Add directories
	err = tree.Files().ForEach(func(d *object.File) error {
		// This is simplified - in reality we'd handle directories separately
		return nil
	})
	
	return names, err
}

// ApplyPatch applies a unified diff patch
func (s *service) ApplyPatch(ctx context.Context, diff string, allowCreate bool) error {
	// For simplicity, this is a stub implementation
	// In production, you'd use git apply or a patch library
	s.logger.Info("Applying patch (stub implementation)")
	return fmt.Errorf("patch application not yet implemented")
}

// AddPatchCommit creates a commit for an applied patch
func (s *service) AddPatchCommit(role, relatesToSHA, summary, body string) (*models.Commit, error) {
	footers := make(map[string]string)
	if relatesToSHA != "" {
		footers["Relates-To"] = relatesToSHA
	}
	
	return s.CreateEmptyCommit(role, summary, body, footers)
}

// ListBranches returns list of all branches
func (s *service) ListBranches() ([]string, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("no repository set")
	}
	
	branches, err := s.repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}
	
	var names []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		names = append(names, ref.Name().Short())
		return nil
	})
	
	return names, err
}

// CreateBranch creates a new branch
func (s *service) CreateBranch(name string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository set")
	}
	
	// Get head reference
	head, err := s.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get head: %w", err)
	}
	
	// Create branch reference
	branchRef := plumbing.NewBranchReferenceName(name)
	err = s.repo.Storer.SetReference(plumbing.NewHashReference(branchRef, head.Hash()))
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	
	return nil
}

// CheckoutBranch switches to a branch
func (s *service) CheckoutBranch(name string) error {
	if s.repo == nil {
		return fmt.Errorf("no repository set")
	}
	
	// Get worktree
	worktree, err := s.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	
	// Checkout branch
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(name),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}
	
	return nil
}

// ComputeMergeBase finds the merge base between two commits
func (s *service) ComputeMergeBase(a, b string) (string, error) {
	if s.repo == nil {
		return "", fmt.Errorf("no repository set")
	}
	
	// This is a simplified implementation
	// In production, you'd use go-git's merge base functionality
	hashA := plumbing.NewHash(a)
	hashB := plumbing.NewHash(b)
	
	_, err := s.repo.CommitObject(hashA)
	if err != nil {
		return "", fmt.Errorf("failed to get commit A: %w", err)
	}
	
	commitB, err := s.repo.CommitObject(hashB)
	if err != nil {
		return "", fmt.Errorf("failed to get commit B: %w", err)
	}
	
	// Simple common ancestor detection (first parent chain)
	// This is a placeholder - real implementation would be more complex
	commitIter, err := s.repo.Log(&git.LogOptions{
		From: hashA,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get log: %w", err)
	}
	defer commitIter.Close()
	
	var mergeBase string
	err = commitIter.ForEach(func(c *object.Commit) error {
		if c.Hash == commitB.Hash {
			mergeBase = c.Hash.String()
			return fmt.Errorf("found")
		}
		return nil
	})
	
	if err != nil && err.Error() != "found" {
		return "", fmt.Errorf("failed to find merge base: %w", err)
	}
	
	if mergeBase == "" {
		return "", fmt.Errorf("no merge base found")
	}
	
	return mergeBase, nil
}

// convertCommit converts a go-git commit to our model
func (s *service) convertCommit(c *object.Commit) models.Commit {
	// Get parents
	var parents []string
	for _, parent := range c.ParentHashes {
		parents = append(parents, parent.String())
	}
	
	// Get author
	author := c.Author.Name
	if c.Author.Email != "" {
		author += fmt.Sprintf(" <%s>", c.Author.Email)
	}
	
	// Parse role and type from subject with fallbacks
	role := ""
	subject := c.Message
	
	// Check for role prefix
	if strings.HasPrefix(subject, "USER:") {
		role = "USER:"
		subject = subject[5:] // Remove "USER:" prefix
	} else if strings.HasPrefix(subject, "AGENT:") {
		role = "AGENT:"
		subject = subject[6:] // Remove "AGENT:" prefix
	}
	
	// Extract type (try to get first word, fallback to empty)
	typee := ""
	if strings.Contains(subject, " ") {
		parts := strings.SplitN(subject, " ", 2)
		if len(parts) > 1 {
			typee = parts[0] // First word as type
		}
	} else if subject != "" {
		typee = subject
	}
	
	// Parse message to get subject, body, and footers
	messageParts := s.parseMessage(c.Message)
	if messageParts.Subject != "" {
		subject = messageParts.Subject
	}
	
	// Build refs list
	var refs []string
	// Add branch heads that point to this commit
	branches, _ := s.repo.Branches()
	branches.ForEach(func(ref *plumbing.Reference) error {
		if ref.Hash() == c.Hash {
			refs = append(refs, ref.Name().Short())
		}
		return nil
	})
	
	return models.Commit{
		SHA:       c.Hash.String(),
		Parents:   parents,
		Author:    author,
		Subject:   subject,
		Body:      messageParts.Body,
		Refs:      refs,
		Role:      role,
		Type:      typee,
		RelatesTo: messageParts.RelatesTo,
		Timestamp: c.Author.When,
	}
}

// messageParts represents parsed commit message
type messageParts struct {
	Subject   string
	Body      string
	RelatesTo string
}

// parseMessage parses a commit message into subject, body, and footers
func (s *service) parseMessage(message string) messageParts {
	// Split on empty line to separate subject from body
	parts := strings.SplitN(message, "\n\n", 2)
	result := messageParts{
		Subject: strings.TrimSpace(parts[0]),
	}
	
	if len(parts) > 1 {
		// Check for footers in the body
		body := parts[1]
		footerLines := strings.Split(body, "\n")
		var bodyLines []string
		
		for _, line := range footerLines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Relates-To:") {
				result.RelatesTo = strings.TrimSpace(strings.TrimPrefix(line, "Relates-To:"))
			} else if strings.Contains(line, ":") && bodyLines == nil {
				// This looks like a footer, so we're in the footer section
			} else {
				bodyLines = append(bodyLines, line)
			}
		}
		
		result.Body = strings.Join(bodyLines, "\n")
	}
	
	return result
}
