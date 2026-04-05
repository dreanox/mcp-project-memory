package retrieve

import (
	"fmt"
	"strings"

	"github.com/dreanox/mcp-project-memory/internal/models"
	"github.com/dreanox/mcp-project-memory/internal/store"
	"github.com/dreanox/mcp-project-memory/internal/structure"
)

// ContextBundle is returned by get_context / pm context.
type ContextBundle struct {
	Notice            string
	ProjectMemories   []models.ProjectMemory
	WorkLogEntries    []models.WorkLogEntry
	StructureMarkdown string
}

// BuildContext loads scored slices and optional structure.
func BuildContext(projectDir, query string, topMem, topLog int, includeStructure bool) (*ContextBundle, error) {
	mem, err := store.LoadProjectMemory(projectDir)
	if err != nil {
		return nil, err
	}
	logs, err := store.LoadWorkLog(projectDir)
	if err != nil {
		return nil, err
	}
	if topMem <= 0 {
		topMem = 5
	}
	if topLog <= 0 {
		topLog = 3
	}
	q := strings.TrimSpace(query)
	var pickedM []models.ProjectMemory
	var pickedW []models.WorkLogEntry

	if q == "" {
		sm := ScoreProjectMemory("", mem)
		for i := 0; i < len(sm) && i < topMem; i++ {
			pickedM = append(pickedM, sm[i].Entry)
		}
		n := topLog
		if n > len(logs) {
			n = len(logs)
		}
		for i := len(logs) - n; i < len(logs); i++ {
			if i >= 0 {
				pickedW = append(pickedW, logs[i])
			}
		}
	} else {
		sm := ScoreProjectMemory(q, mem)
		for i := 0; i < len(sm) && len(pickedM) < topMem; i++ {
			if sm[i].Score > 0 {
				pickedM = append(pickedM, sm[i].Entry)
			}
		}
		if len(pickedM) == 0 {
			for i := 0; i < len(sm) && i < topMem && i < 3; i++ {
				pickedM = append(pickedM, sm[i].Entry)
			}
		}
		sw := ScoreWorkLog(q, logs)
		for i := 0; i < len(sw) && len(pickedW) < topLog; i++ {
			if sw[i].Score > 0 {
				pickedW = append(pickedW, sw[i].Entry)
			}
		}
	}

	b := &ContextBundle{
		ProjectMemories: pickedM,
		WorkLogEntries:  pickedW,
	}
	parts := []string{
		fmt.Sprintf("Included %d project-memory and %d work-log entries.", len(pickedM), len(pickedW)),
	}
	if includeStructure {
		g, err := structure.GlobalView(projectDir)
		if err != nil {
			return nil, err
		}
		b.StructureMarkdown = g.Body
		parts = append(parts, "Structure: "+g.Notice+".")
	}
	b.Notice = strings.Join(parts, " ")
	return b, nil
}
