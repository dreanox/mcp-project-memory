package svc

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dreanox/mcp-project-memory/internal/ids"
	"github.com/dreanox/mcp-project-memory/internal/models"
	"github.com/dreanox/mcp-project-memory/internal/retrieve"
	"github.com/dreanox/mcp-project-memory/internal/store"
	"github.com/dreanox/mcp-project-memory/internal/structure"
)

// Service is the shared backend for MCP and CLI.
type Service struct {
	ProjectDir string
}

func New(workspace string) (*Service, error) {
	dir, err := store.ResolveWorkspace(workspace)
	if err != nil {
		return nil, err
	}
	pd, err := store.ProjectDir(dir)
	if err != nil {
		return nil, err
	}
	return &Service{ProjectDir: pd}, nil
}

// --- work log ---

func (s *Service) WorklogAppend(date, summary, tool string, components, tags []string, notes string) (*models.WorkLogEntry, error) {
	e := &models.WorkLogEntry{
		ID:         ids.New(),
		Date:       date,
		Summary:    summary,
		Tool:       tool,
		Components: components,
		Tags:       tags,
		Notes:      notes,
		CreatedAt:  time.Now().UTC(),
	}
	if e.Date == "" {
		e.Date = time.Now().UTC().Format("2006-01-02")
	}
	if e.Tool == "" {
		e.Tool = "manual"
	}
	return e, store.AppendWorkLog(s.ProjectDir, e)
}

func (s *Service) WorklogList(from, to string) ([]models.WorkLogEntry, error) {
	all, err := store.LoadWorkLog(s.ProjectDir)
	if err != nil {
		return nil, err
	}
	if from == "" && to == "" {
		return all, nil
	}
	var out []models.WorkLogEntry
	for _, e := range all {
		if from != "" && e.Date < from {
			continue
		}
		if to != "" && e.Date > to {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

// --- project memory ---

func (s *Service) ProjectMemoryAdd(id, typ, scope, content string, tags []string) (*models.ProjectMemory, error) {
	if id == "" {
		id = ids.New()
	}
	m := &models.ProjectMemory{
		ID:      id,
		Type:    typ,
		Scope:   scope,
		Content: content,
		Tags:    tags,
	}
	if m.Type == "" {
		m.Type = "other"
	}
	return m, store.UpsertProjectMemory(s.ProjectDir, m)
}

func (s *Service) ProjectMemoryUpdate(id, typ, scope, content string, tags []string) error {
	list, err := store.LoadProjectMemory(s.ProjectDir)
	if err != nil {
		return err
	}
	var cur *models.ProjectMemory
	for i := range list {
		if list[i].ID == id {
			cur = &list[i]
			break
		}
	}
	if cur == nil {
		return fmt.Errorf("no project memory with id %q", id)
	}
	if typ != "" {
		cur.Type = typ
	}
	if scope != "" {
		cur.Scope = scope
	}
	if content != "" {
		cur.Content = content
	}
	if tags != nil {
		cur.Tags = tags
	}
	return store.UpsertProjectMemory(s.ProjectDir, cur)
}

func (s *Service) ProjectMemoryDelete(id string) error {
	return store.DeleteProjectMemory(s.ProjectDir, id)
}

func (s *Service) ProjectMemoryList() ([]models.ProjectMemory, error) {
	return store.LoadProjectMemory(s.ProjectDir)
}

func (s *Service) MemorySearch(track, query string, limit int) (mem []models.ProjectMemory, logs []models.WorkLogEntry, notice string, err error) {
	if limit <= 0 {
		limit = 20
	}
	track = strings.ToLower(strings.TrimSpace(track))
	q := strings.TrimSpace(query)
	switch track {
	case "project_memory", "memory":
		pm, err := store.LoadProjectMemory(s.ProjectDir)
		if err != nil {
			return nil, nil, "", err
		}
		if q == "" {
			for i := 0; i < len(pm) && len(mem) < limit; i++ {
				mem = append(mem, pm[i])
			}
		} else {
			sm := retrieve.ScoreProjectMemory(q, pm)
			for i := 0; i < len(sm) && len(mem) < limit; i++ {
				if sm[i].Score > 0 {
					mem = append(mem, sm[i].Entry)
				}
			}
		}
		notice = fmt.Sprintf("project-memory matches: %d", len(mem))
		return mem, nil, notice, nil
	case "worklog", "work_log":
		all, err := store.LoadWorkLog(s.ProjectDir)
		if err != nil {
			return nil, nil, "", err
		}
		if q == "" {
			n := limit
			if n > len(all) {
				n = len(all)
			}
			for i := len(all) - n; i < len(all); i++ {
				if i >= 0 {
					logs = append(logs, all[i])
				}
			}
		} else {
			sw := retrieve.ScoreWorkLog(q, all)
			for i := 0; i < len(sw) && len(logs) < limit; i++ {
				if sw[i].Score > 0 {
					logs = append(logs, sw[i].Entry)
				}
			}
			if len(logs) == 0 {
				for i := 0; i < len(sw) && i < limit; i++ {
					logs = append(logs, sw[i].Entry)
				}
			}
		}
		notice = fmt.Sprintf("work-log matches: %d (limit %d)", len(logs), limit)
		return nil, logs, notice, nil
	case "both", "", "all":
		pm, err := store.LoadProjectMemory(s.ProjectDir)
		if err != nil {
			return nil, nil, "", err
		}
		pl, err := store.LoadWorkLog(s.ProjectDir)
		if err != nil {
			return nil, nil, "", err
		}
		if q == "" {
			for i := 0; i < len(pm) && len(mem) < limit; i++ {
				mem = append(mem, pm[i])
			}
			n := limit
			if n > len(pl) {
				n = len(pl)
			}
			for i := len(pl) - n; i < len(pl); i++ {
				if i >= 0 {
					logs = append(logs, pl[i])
				}
			}
		} else {
			sm := retrieve.ScoreProjectMemory(q, pm)
			for i := 0; i < len(sm) && len(mem) < limit; i++ {
				if sm[i].Score > 0 {
					mem = append(mem, sm[i].Entry)
				}
			}
			sw := retrieve.ScoreWorkLog(q, pl)
			for i := 0; i < len(sw) && len(logs) < limit; i++ {
				if sw[i].Score > 0 {
					logs = append(logs, sw[i].Entry)
				}
			}
		}
		notice = fmt.Sprintf("project-memory: %d, work-log: %d", len(mem), len(logs))
		return mem, logs, notice, nil
	default:
		return nil, nil, "", fmt.Errorf("unknown track %q (use project_memory, worklog, or both)", track)
	}
}

func (s *Service) GetContext(query string, topMem, topLog int, includeStructure bool) (*retrieve.ContextBundle, error) {
	return retrieve.BuildContext(s.ProjectDir, query, topMem, topLog, includeStructure)
}

func (s *Service) StructureList() ([]structure.Fragment, error) {
	return structure.ListFragments(s.ProjectDir)
}

func (s *Service) StructureShow(focus string) (*structure.View, error) {
	focus = strings.TrimSpace(focus)
	if focus == "" {
		return structure.GlobalView(s.ProjectDir)
	}
	return structure.FocusView(s.ProjectDir, focus)
}

func (s *Service) StructureMap(request string) (*structure.View, error) {
	return structure.MapForRequest(s.ProjectDir, request)
}

// FormatBundle renders ContextBundle as markdown text for tools/CLI.
func FormatBundle(b *retrieve.ContextBundle) string {
	var o strings.Builder
	o.WriteString("## notice\n")
	o.WriteString(b.Notice)
	o.WriteString("\n\n## project memory\n\n")
	for _, m := range b.ProjectMemories {
		o.WriteString(fmt.Sprintf("- **[%s]** scope=%s — %s\n  tags: %v\n", m.Type, m.Scope, m.Content, m.Tags))
	}
	if len(b.ProjectMemories) == 0 {
		o.WriteString("_(none)_\n")
	}
	o.WriteString("\n## work log\n\n")
	for _, w := range b.WorkLogEntries {
		o.WriteString(fmt.Sprintf("- **%s** [%s] %s\n", w.Date, w.Tool, w.Summary))
		if w.Notes != "" {
			o.WriteString(fmt.Sprintf("  notes: %s\n", w.Notes))
		}
	}
	if len(b.WorkLogEntries) == 0 {
		o.WriteString("_(none)_\n")
	}
	if b.StructureMarkdown != "" {
		o.WriteString("\n## structure\n\n")
		o.WriteString(b.StructureMarkdown)
		o.WriteString("\n")
	}
	return o.String()
}

// JSON helper for MCP structured output
func MustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
