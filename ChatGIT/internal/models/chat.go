package models

// Message represents a chat message
type Message struct {
	ID       string `json:"id"`
	Role     string `json:"role"` // "user" or "assistant"
	Content  string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	TokenCount int  `json:"tokenCount"`
}

// ChatSession represents a chat session
type ChatSession struct {
	ID       string    `json:"id"`
	RepoPath string    `json:"repoPath"`
	Messages []Message `json:"messages"`
	Created  int64     `json:"created"`
	Updated  int64     `json:"updated"`
}

// PatchProposal represents a proposed patch
type PatchProposal struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Diff        string `json:"diff"`
	Files       []string `json:"files"`
	Approved    bool   `json:"approved"`
}

// ChatConfig represents chat configuration
type ChatConfig struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	APIKey   string `json:"apiKey"`
	BaseURL  string `json:"baseUrl"`
	MaxTokens int   `json:"maxTokens"`
}

// ChatStreamToken represents a token in the stream
type ChatStreamToken struct {
	Token   string `json:"token"`
	IsEOF   bool   `json:"isEOF"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
