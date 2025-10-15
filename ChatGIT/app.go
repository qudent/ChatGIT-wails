package main

import (
	"ChatGIT/internal/config"
	"ChatGIT/internal/logger"
	"ChatGIT/internal/services/chat"
	"ChatGIT/internal/services/diff"
	"ChatGIT/internal/services/fs"
	"ChatGIT/internal/services/git"
	"ChatGIT/internal/services/session"
	"ChatGIT/internal/services/settings"
	"ChatGIT/internal/models"
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
	
	// Configuration and logging
	cfg    *config.Config
	logger *logger.Logger
	
	// Services
	gitService      git.Service
	fsService       fs.Service
	diffService     diff.Service
	chatService     chat.Service
	settingsService settings.Service
	sessionService  session.Service
	
	// Chat provider
	chatProvider chat.Provider
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Initialize configuration
	a.cfg = config.Load()
	
	// Initialize logger
	a.logger = logger.New(a.cfg.GetLogLevel(), a.cfg.GetLogFormat())
	a.logger.Info("ChatGIT application starting up")
	
	// Initialize services
	a.initializeServices()
	
	a.logger.Info("ChatGIT application ready")
}

// initializeServices initializes all services
func (a *App) initializeServices() {
	// Initialize git service
	a.gitService = git.New(a.logger)
	
	// Initialize file system service
	a.fsService = fs.New(a.logger)
	
	// Initialize diff service
	a.diffService = diff.New(a.logger)
	
	// Initialize settings service
	a.settingsService = settings.New(a.logger)
	
	// Initialize session service
	a.sessionService = session.New(a.cfg, a.logger)
	
	// Initialize chat provider
	chatConfig, _ := a.settingsService.GetChatConfig(context.Background())
	if chatConfig != nil {
		a.chatProvider = chat.NewOpenAIProvider(a.logger, chatConfig.APIKey, chatConfig.BaseURL)
	} else {
		a.chatProvider = chat.NewOpenAIProvider(a.logger, "", "")
	}
	
	// Initialize chat service
	a.chatService = chat.New(a.logger, a.chatProvider)
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	a.logger.Debug("Frontend is ready")
}

// beforeClose is called when the application is about to quit
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	a.logger.Info("Application shutting down")
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	a.logger.Info("Application shutdown complete")
}

// Git operations

// SetGitRepository sets the current git repository
func (a *App) SetGitRepository(path string) error {
	a.logger.Info("Setting git repository: %s", path)
	
	// Set repository in git service
	if err := a.gitService.SetRepo(path); err != nil {
		return fmt.Errorf("failed to set git repository: %w", err)
	}
	
	// Update file system service
	a.fsService.SetRepoRoot(path)
	
	// Update diff service with repository
	a.diffService.SetRepository(a.gitService.GetRepository())
	
	// Save to session state
	if state, err := a.sessionService.GetUIState(context.Background()); err == nil {
		state.CurrentRepo = path
		_ = a.sessionService.SaveUIState(context.Background(), state)
	}
	
	// Save repo metadata
	metadata, _ := a.sessionService.GetRepoMetadata(context.Background(), path)
	metadata.Path = path
	metadata.LastUsed = 0 // Will be updated by session service
	_ = a.sessionService.SaveRepoMetadata(context.Background(), metadata)
	
	return nil
}

// GetGitStatus returns current git repository status
func (a *App) GetGitStatus() (*models.Status, error) {
	return a.gitService.Status()
}

// GetGitLog returns commit log
func (a *App) GetGitLog(limit, skip int) (*models.CommitList, error) {
	return a.gitService.LogSpine(context.Background(), limit, skip)
}

// CreateGitCommit creates a new git commit
func (a *App) CreateGitCommit(role, subject, body string) (*models.Commit, error) {
	return a.gitService.CreateEmptyCommit(role, subject, body, nil)
}

// File operations

// SafeReadFile safely reads a file
func (a *App) SafeReadFile(path string) (string, error) {
	data, err := a.fsService.SafeRead(context.Background(), path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SafeWriteFile safely writes a file
func (a *App) SafeWriteFile(path, content string) error {
	return a.fsService.SafeWrite(context.Background(), path, []byte(content))
}

// ListFiles lists files in a directory
func (a *App) ListFiles(path string) ([]string, error) {
	return a.fsService.ListFiles(context.Background(), path)
}

// Diff operations

// GetDiffCommit gets diff for a commit
func (a *App) GetDiffCommit(commitSHA string) (*models.Diff, error) {
	return a.diffService.DiffCommit(context.Background(), commitSHA)
}

// GetDiffWorkingTree gets working tree diff
func (a *App) GetDiffWorkingTree() (*models.Diff, error) {
	return a.diffService.DiffWorkingTree(context.Background())
}

// Chat operations

// StartChatMessage starts a chat message
func (a *App) StartChatMessage(role, text string) (<-chan models.ChatStreamToken, error) {
	return a.chatService.StartMessage(context.Background(), role, text)
}

// ProposePatch proposes a patch
func (a *App) ProposePatch(description string) (*models.PatchProposal, error) {
	return a.chatService.ProposePatch(context.Background(), description)
}

// Settings operations

// GetSettings returns current settings
func (a *App) GetSettings() (*models.Settings, error) {
	return a.settingsService.GetSettings(context.Background())
}

// UpdateSettings updates settings
func (a *App) UpdateSettings(settings *models.Settings) error {
	return a.settingsService.SetSettings(context.Background(), settings)
}

// GetAvailableProviders returns available chat providers
func (a *App) GetAvailableProviders() []string {
	return a.settingsService.ListAvailableProviders()
}

// GetAvailableModels returns available models for a provider
func (a *App) GetAvailableModels(provider string) []string {
	return a.settingsService.ListAvailableModels(provider)
}

// Session operations

// GetRecentRepos returns recent repositories
func (a *App) GetRecentRepos(limit int) ([]*models.RepoMetadata, error) {
	return a.sessionService.ListRecentRepos(context.Background(), limit)
}

// GetUIState returns UI state
func (a *App) GetUIState() (*models.UIState, error) {
	return a.sessionService.GetUIState(context.Background())
}

// SaveUIState saves UI state
func (a *App) SaveUIState(state *models.UIState) error {
	return a.sessionService.SaveUIState(context.Background(), state)
}

// Utility methods

// SelectDirectory shows directory selection dialog
func (a *App) SelectDirectory() (string, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Git Repository",
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

// ShowError shows an error dialog
func (a *App) ShowError(title, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   title,
		Message: message,
	})
}

// ShowInfo shows an info dialog
func (a *App) ShowInfo(title, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
}

// Greet returns a greeting for the given name (legacy method)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, Welcome to ChatGIT!", name)
}
