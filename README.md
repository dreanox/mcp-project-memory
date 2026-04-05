# Project Memory (MCP + CLI)

Go implementation of the design in [`context.md`](context.md): **work log** (JSONL) + **project memory** (JSON) + optional **structure** Markdown, stored under the user’s **Cursor home** (not inside the git repo).

## Build

```bash
go build -o mcp-memory ./cmd/mcp-memory
go build -o pm ./cmd/pm
```

On Windows, outputs are written to `bin/` if you use the commands above from the repo root.

## Storage layout

Default base: `%USERPROFILE%\.cursor\project-memory\` (macOS/Linux: `~/.cursor/project-memory/`).

Per workspace folder:

`{slug}-{8-hex}` derived from the **absolute** workspace path (e.g. `mcpmemory-a1b2c3d4`).

Inside that directory:

- `work-log.jsonl`
- `project-memory.json`
- `project-structure.md` (optional)
- `structure/*.md` (optional fragments)

### Environment

| Variable | Purpose |
|----------|---------|
| `PROJECT_MEMORY_WORKSPACE` | Absolute path to the open project (MCP/CLI resolve storage key from this). |
| `PROJECT_MEMORY_ROOT` | Override base dir instead of `~/.cursor/project-memory`. |
| `PROJECT_MEMORY_KEY` | Override the `{project-key}` folder name. |
| `PROJECT_MEMORY_DIR` | Use this directory **as** the project folder (full path to where JSON/MD live). |

## Cursor MCP

Add a server entry (User or Project MCP settings) pointing at the built binary. Set `cwd` to your repository root so the default working directory is correct, **or** set `PROJECT_MEMORY_WORKSPACE` in `env`.

Example (adjust paths):

```json
{
  "mcpServers": {
    "project-memory": {
      "command": "C:\\Users\\you\\Documents\\Work\\MCPMemory\\bin\\mcp-memory.exe",
      "cwd": "C:\\Users\\you\\Documents\\Work\\YourRepo"
    }
  }
}
```

Optional `env`:

```json
"env": {
  "PROJECT_MEMORY_WORKSPACE": "C:\\Users\\you\\Documents\\Work\\YourRepo"
}
```

## Claude Desktop

Same `command` / `env` idea in `claude_desktop_config.json` under `mcpServers`.

## CLI `pm`

```text
pm worklog add --summary "..." [--date YYYY-MM-DD] [--tool cursor|claude|manual] [--tags a,b]
pm worklog list [--from DATE] [--to DATE]
pm memory add --type architecture --content "..." [--scope global] [--tags a,b]
pm memory search [--track project_memory|worklog|both] [--limit 20] <query>
pm memory list
pm memory delete <id>
pm memory update <id> [--type T] [--content C] [--tags a,b]
pm context [--structure] [--top-mem 5] [--top-log 3] <query>
pm structure show [focus]
pm structure list
pm structure map "<natural language request>"
```

## MCP tools

`worklog_append`, `worklog_list`, `project_memory_add`, `project_memory_update`, `project_memory_delete`, `project_memory_list`, `memory_search`, `get_context`, `get_project_structure`, `get_structure_for_request`, `structure_list_fragments`.

Tool results include a **`notice`** string: repeat it to the user when you use this server (see server `Instructions`).

## License

MIT (same as dependencies; add a `LICENSE` file if you publish the repo).
