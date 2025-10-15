package models

import "time"

// Commit represents a git commit with additional metadata
type Commit struct {
	SHA       string    `json:"sha"`
	Parents   []string  `json:"parents"`
	Author    string    `json:"author"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Refs      []string  `json:"refs"`
	Role      string    `json:"role"`      // USER: or AGENT:
	Type      string    `json:"type"`      // extracted from subject
	RelatesTo string    `json:"relatesTo"` // from Relates-To footer
	Timestamp time.Time `json:"timestamp"`
}

// CommitList represents a paginated list of commits
type CommitList struct {
	Commits []Commit `json:"commits"`
	Total   int      `json:"total"`
	HasMore bool     `json:"hasMore"`
}

// Status represents git repository status
type Status struct {
	Branch     string `json:"branch"`
	Head       string `json:"head"`
	Ahead      int    `json:"ahead"`
	Behind     int    `json:"behind"`
	Staged     int    `json:"staged"`
	Unstaged   int    `json:"unstaged"`
	Untracked  int    `json:"untracked"`
	IsClean    bool   `json:"isClean"`
}

// Diff represents a unified diff
type Diff struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	Content string   `json:"content"`
	Hunks   []Hunk   `json:"hunks"`
	Files   []string `json:"files"`
}

// Hunk represents a diff hunk
type Hunk struct {
	OldStart int    `json:"oldStart"`
	OldLines int    `json:"oldLines"`
	NewStart int    `json:"newStart"`
	NewLines int    `json:"newLines"`
	Content  string `json:"content"`
}
