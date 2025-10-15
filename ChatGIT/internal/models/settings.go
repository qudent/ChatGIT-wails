package models

// Settings represents application settings
type Settings struct {
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	MaxTokens   int    `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
}

// UIState represents UI state to persist
type UIState struct {
	CurrentRepo string `json:"currentRepo"`
	OpenTabs    []string `json:"openTabs"`
	SplitLayout bool   `json:"splitLayout"`
	Theme       string `json:"theme"`
}

// RepoMetadata stores repository-specific metadata
type RepoMetadata struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	LastUsed int64     `json:"lastUsed"`
	Favorite bool      `json:"favorite"`
	Tags     []string  `json:"tags"`
}
