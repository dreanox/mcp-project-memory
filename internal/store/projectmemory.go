package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/master/mcp-memory/internal/models"
)

const projectMemoryFile = "project-memory.json"

// LoadProjectMemory reads the JSON array of project memories.
func LoadProjectMemory(projectDir string) ([]models.ProjectMemory, error) {
	path := filepath.Join(projectDir, projectMemoryFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var list []models.ProjectMemory
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// SaveProjectMemory writes the full list atomically.
func SaveProjectMemory(projectDir string, list []models.ProjectMemory) error {
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return err
	}
	if list == nil {
		list = []models.ProjectMemory{}
	}
	path := filepath.Join(projectDir, projectMemoryFile)
	tmp := path + ".tmp"
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// UpsertProjectMemory adds or updates by ID.
func UpsertProjectMemory(projectDir string, m *models.ProjectMemory) error {
	list, err := LoadProjectMemory(projectDir)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for i := range list {
		if list[i].ID == m.ID {
			m.CreatedAt = list[i].CreatedAt
			m.UpdatedAt = now
			list[i] = *m
			return SaveProjectMemory(projectDir, list)
		}
	}
	m.CreatedAt = now
	m.UpdatedAt = now
	list = append(list, *m)
	return SaveProjectMemory(projectDir, list)
}

// DeleteProjectMemory removes by ID.
func DeleteProjectMemory(projectDir, id string) error {
	list, err := LoadProjectMemory(projectDir)
	if err != nil {
		return err
	}
	var next []models.ProjectMemory
	for _, m := range list {
		if m.ID != id {
			next = append(next, m)
		}
	}
	return SaveProjectMemory(projectDir, next)
}
