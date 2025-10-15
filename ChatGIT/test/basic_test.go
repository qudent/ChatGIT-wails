// Basic test file to demonstrate the backend implementation
package test

import (
	"ChatGIT/internal/config"
	"ChatGIT/internal/logger"
	"ChatGIT/internal/services/chat"
	"ChatGIT/internal/services/diff"
	"ChatGIT/internal/services/fs"
	"ChatGIT/internal/services/git"
	"ChatGIT/internal/services/session"
	"ChatGIT/internal/services/settings"
	"testing"
)

func TestServiceInitialization(t *testing.T) {
	// Test that all services can be initialized without errors
	
	// Initialize configuration and logger
	cfg := config.Load()
	logger := logger.New(cfg.GetLogLevel(), cfg.GetLogFormat())
	
	// Test git service
	gitService := git.New(logger)
	if gitService == nil {
		t.Fatal("Failed to initialize GitService")
	}
	
	// Test fs service
	fsService := fs.New(logger)
	if fsService == nil {
		t.Fatal("Failed to initialize FSService")
	}
	
	// Test diff service
	diffService := diff.New(logger)
	if diffService == nil {
		t.Fatal("Failed to initialize DiffService")
	}
	
	// Test session service
	sessionService := session.New(cfg, logger)
	if sessionService == nil {
		t.Fatal("Failed to initialize SessionService")
	}
	
	// Test settings service
	settingsService := settings.New(logger)
	if settingsService == nil {
		t.Fatal("Failed to initialize SettingsService")
	}
	
	// Test chat service with mock provider
	chatProvider := chat.NewOpenAIProvider(logger, "", "")
	chatService := chat.New(logger, chatProvider)
	if chatService == nil {
		t.Fatal("Failed to initialize ChatService")
	}
	
	t.Log("All services initialized successfully")
}

func TestSettingsService(t *testing.T) {
	cfg := config.Load()
	logger := logger.New(cfg.GetLogLevel(), cfg.GetLogFormat())
	settingsService := settings.New(logger)
	
	// Test getting settings
	settings, err := settingsService.GetSettings(nil)
	if err != nil {
		t.Fatalf("Failed to get settings: %v", err)
	}
	
	if settings.Provider == "" {
		t.Error("Provider should not be empty")
	}
	
	// Test getting available providers
	providers := settingsService.ListAvailableProviders()
	if len(providers) == 0 {
		t.Error("Should have at least one available provider")
	}
	
	// Test getting models for OpenAI
	models := settingsService.ListAvailableModels("openai")
	if len(models) == 0 {
		t.Error("Should have at least one model for OpenAI provider")
	}
	
	t.Logf("Settings: provider=%s, model=%s", settings.Provider, settings.Model)
	t.Logf("Available providers: %v", providers)
	t.Logf("OpenAI models: %v", models)
}

func TestFSService(t *testing.T) {
	cfg := config.Load()
	logger := logger.New(cfg.GetLogLevel(), cfg.GetLogFormat())
	fsService := fs.New(logger)
	
	// Test path validation with current directory
	testPath := "."
	
	fsService.SetRepoRoot(testPath)
	
	repoRoot := fsService.GetRepoRoot()
	if repoRoot == "" {
		t.Error("Repo root should not be empty")
	}
	
	t.Logf("FS Service initialized with repo root: %s", repoRoot)
	
	// Test path validation 
	if err := fsService.ValidatePath("./test/basic_test.go"); err != nil {
		t.Errorf("Valid path should pass validation: %v", err)
	}
	
	// Test unsafe path validation
	if err := fsService.ValidatePath("/etc/passwd"); err == nil {
		t.Error("Unsafe path should fail validation")
	}
}
