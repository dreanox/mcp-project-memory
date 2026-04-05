package mcpapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/master/mcp-memory/internal/svc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTools wires all MCP tools to the given service.
func RegisterTools(server *mcp.Server, s *svc.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "worklog_append",
		Description: "Append one entry to the daily work log (tasks done in Cursor/Claude).",
	}, worklogAppend(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "worklog_list",
		Description: "List work log entries, optionally filtered by date range (YYYY-MM-DD).",
	}, worklogList(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_memory_add",
		Description: "Add durable project memory (architecture, pattern, decision, successful_task, etc.).",
	}, projectMemoryAdd(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_memory_update",
		Description: "Update an existing project memory entry by id.",
	}, projectMemoryUpdate(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_memory_delete",
		Description: "Delete a project memory entry by id.",
	}, projectMemoryDelete(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_memory_list",
		Description: "List all project memory entries.",
	}, projectMemoryList(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "memory_search",
		Description: "Search project memory and/or work log. track: project_memory | worklog | both",
	}, memorySearch(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_context",
		Description: "Ranked context bundle for a task: project memory + recent work log; optional structure. Always tell the user the notice.",
	}, getContext(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project_structure",
		Description: "Return structure markdown. Omit focus for global (index + fragment list); or pass domain slug or structure/foo.md path.",
	}, getProjectStructure(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_structure_for_request",
		Description: "Natural-language request: pick matching structure fragments, follow cross-refs, compose one view. Tell the user the notice.",
	}, getStructureForRequest(s))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "structure_list_fragments",
		Description: "List structure/*.md fragments for this project.",
	}, structureListFragments(s))
}

type toolOut struct {
	Notice string `json:"notice" jsonschema:"what was used; repeat to the user"`
	Text   string `json:"text" jsonschema:"human-readable markdown body"`
}

func textResult(notice, body string) (*mcp.CallToolResult, toolOut, error) {
	o := toolOut{Notice: notice, Text: body}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: body + "\n\n---\n**notice:** " + notice}},
	}, o, nil
}

// --- worklog_append ---

type worklogAppendIn struct {
	Date       string   `json:"date,omitempty" jsonschema:"ISO date YYYY-MM-DD, default today"`
	Summary    string   `json:"summary" jsonschema:"what was done"`
	Tool       string   `json:"tool,omitempty" jsonschema:"cursor | claude | manual"`
	Components []string `json:"components,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Notes      string   `json:"notes,omitempty"`
}

func worklogAppend(s *svc.Service) func(context.Context, *mcp.CallToolRequest, worklogAppendIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in worklogAppendIn) (*mcp.CallToolResult, toolOut, error) {
		e, err := s.WorklogAppend(in.Date, in.Summary, in.Tool, in.Components, in.Tags, in.Notes)
		if err != nil {
			return nil, toolOut{}, err
		}
		b, _ := json.MarshalIndent(e, "", "  ")
		notice := fmt.Sprintf("Appended work-log entry id=%s date=%s", e.ID, e.Date)
		return textResult(notice, string(b))
	}
}

// --- worklog_list ---

type worklogListIn struct {
	From string `json:"from,omitempty" jsonschema:"start date YYYY-MM-DD inclusive"`
	To   string `json:"to,omitempty" jsonschema:"end date YYYY-MM-DD inclusive"`
}

func worklogList(s *svc.Service) func(context.Context, *mcp.CallToolRequest, worklogListIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in worklogListIn) (*mcp.CallToolResult, toolOut, error) {
		list, err := s.WorklogList(in.From, in.To)
		if err != nil {
			return nil, toolOut{}, err
		}
		b, _ := json.MarshalIndent(list, "", "  ")
		notice := fmt.Sprintf("Listed %d work-log entries", len(list))
		return textResult(notice, string(b))
	}
}

// --- project_memory_add ---

type projectMemoryAddIn struct {
	ID      string   `json:"id,omitempty" jsonschema:"optional; generated if empty"`
	Type    string   `json:"type" jsonschema:"architecture | successful_task | pattern | decision | anti-pattern | component | other"`
	Scope   string   `json:"scope,omitempty" jsonschema:"global | feature | component"`
	Content string   `json:"content" jsonschema:"concise actionable text"`
	Tags    []string `json:"tags,omitempty"`
}

func projectMemoryAdd(s *svc.Service) func(context.Context, *mcp.CallToolRequest, projectMemoryAddIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in projectMemoryAddIn) (*mcp.CallToolResult, toolOut, error) {
		m, err := s.ProjectMemoryAdd(in.ID, in.Type, in.Scope, in.Content, in.Tags)
		if err != nil {
			return nil, toolOut{}, err
		}
		raw, _ := json.MarshalIndent(m, "", "  ")
		notice := fmt.Sprintf("Added project memory id=%s type=%s", m.ID, m.Type)
		return textResult(notice, string(raw))
	}
}

// --- project_memory_update ---

type projectMemoryUpdateIn struct {
	ID      string   `json:"id" jsonschema:"entry id"`
	Type    string   `json:"type,omitempty"`
	Scope   string   `json:"scope,omitempty"`
	Content string   `json:"content,omitempty"`
	Tags    []string `json:"tags,omitempty" jsonschema:"replace tags when provided"`
}

func projectMemoryUpdate(s *svc.Service) func(context.Context, *mcp.CallToolRequest, projectMemoryUpdateIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in projectMemoryUpdateIn) (*mcp.CallToolResult, toolOut, error) {
		if err := s.ProjectMemoryUpdate(in.ID, in.Type, in.Scope, in.Content, in.Tags); err != nil {
			return nil, toolOut{}, err
		}
		notice := fmt.Sprintf("Updated project memory id=%s", in.ID)
		return textResult(notice, notice)
	}
}

// --- project_memory_delete ---

type projectMemoryDeleteIn struct {
	ID string `json:"id" jsonschema:"entry id"`
}

func projectMemoryDelete(s *svc.Service) func(context.Context, *mcp.CallToolRequest, projectMemoryDeleteIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in projectMemoryDeleteIn) (*mcp.CallToolResult, toolOut, error) {
		if err := s.ProjectMemoryDelete(in.ID); err != nil {
			return nil, toolOut{}, err
		}
		notice := fmt.Sprintf("Deleted project memory id=%s", in.ID)
		return textResult(notice, notice)
	}
}

// --- project_memory_list ---

type emptyIn struct{}

func projectMemoryList(s *svc.Service) func(context.Context, *mcp.CallToolRequest, emptyIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in emptyIn) (*mcp.CallToolResult, toolOut, error) {
		list, err := s.ProjectMemoryList()
		if err != nil {
			return nil, toolOut{}, err
		}
		b, _ := json.MarshalIndent(list, "", "  ")
		notice := fmt.Sprintf("Listed %d project-memory entries", len(list))
		return textResult(notice, string(b))
	}
}

// --- memory_search ---

type memorySearchIn struct {
	Track string `json:"track,omitempty" jsonschema:"project_memory | worklog | both (default both)"`
	Query string `json:"query,omitempty" jsonschema:"search text"`
	Limit int    `json:"limit,omitempty" jsonschema:"max results per track, default 20"`
}

func memorySearch(s *svc.Service) func(context.Context, *mcp.CallToolRequest, memorySearchIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in memorySearchIn) (*mcp.CallToolResult, toolOut, error) {
		mem, logs, notice, err := s.MemorySearch(in.Track, in.Query, in.Limit)
		if err != nil {
			return nil, toolOut{}, err
		}
		out := map[string]any{"project_memory": mem, "work_log": logs}
		b, _ := json.MarshalIndent(out, "", "  ")
		return textResult(notice, string(b))
	}
}

// --- get_context ---

type getContextIn struct {
	Query            string `json:"query,omitempty" jsonschema:"task description to match; empty returns recent items"`
	TopMem           int    `json:"top_mem,omitempty" jsonschema:"max project memory hits, default 5"`
	TopLog           int    `json:"top_log,omitempty" jsonschema:"max work log lines, default 3"`
	IncludeStructure bool   `json:"include_structure,omitempty" jsonschema:"attach global structure index"`
}

func getContext(s *svc.Service) func(context.Context, *mcp.CallToolRequest, getContextIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in getContextIn) (*mcp.CallToolResult, toolOut, error) {
		b, err := s.GetContext(in.Query, in.TopMem, in.TopLog, in.IncludeStructure)
		if err != nil {
			return nil, toolOut{}, err
		}
		body := svc.FormatBundle(b)
		return textResult(b.Notice, body)
	}
}

// --- get_project_structure ---

type getProjectStructureIn struct {
	Focus string `json:"focus,omitempty" jsonschema:"empty for global; or domain slug or structure/foo.md"`
}

func getProjectStructure(s *svc.Service) func(context.Context, *mcp.CallToolRequest, getProjectStructureIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in getProjectStructureIn) (*mcp.CallToolResult, toolOut, error) {
		v, err := s.StructureShow(strings.TrimSpace(in.Focus))
		if err != nil {
			return nil, toolOut{}, err
		}
		return textResult(v.Notice, v.Body)
	}
}

// --- get_structure_for_request ---

type getStructureForRequestIn struct {
	Request string `json:"request" jsonschema:"what the user wants oriented in structure form"`
}

func getStructureForRequest(s *svc.Service) func(context.Context, *mcp.CallToolRequest, getStructureForRequestIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in getStructureForRequestIn) (*mcp.CallToolResult, toolOut, error) {
		v, err := s.StructureMap(in.Request)
		if err != nil {
			return nil, toolOut{}, err
		}
		return textResult(v.Notice, v.Body)
	}
}

// --- structure_list_fragments ---

func structureListFragments(s *svc.Service) func(context.Context, *mcp.CallToolRequest, emptyIn) (*mcp.CallToolResult, toolOut, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, in emptyIn) (*mcp.CallToolResult, toolOut, error) {
		frags, err := s.StructureList()
		if err != nil {
			return nil, toolOut{}, err
		}
		b, _ := json.MarshalIndent(frags, "", "  ")
		notice := fmt.Sprintf("%d structure fragment(s)", len(frags))
		return textResult(notice, string(b))
	}
}
