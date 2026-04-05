package models

import "time"

// WorkLogEntry is one append-only work-log line (JSONL).
type WorkLogEntry struct {
	ID         string    `json:"id"`
	Date       string    `json:"date"`
	Summary    string    `json:"summary"`
	Tool       string    `json:"tool"`
	Components []string  `json:"components,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ProjectMemory is stable project knowledge.
type ProjectMemory struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Scope     string    `json:"scope"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}
