# Project Memory (MCP + CLI)

Go **work log** (JSONL) + **project memory** (JSON) + optional **structure** Markdown, stored under the user’s **Cursor home** (not inside the git repo).

**Repository:** [https://github.com/dreanox/mcp-project-memory](https://github.com/dreanox/mcp-project-memory)

## Prerequisites: install Go

This project is written in **Go**. You must install the Go toolchain on your machine before building. The version must match **`go.mod`** (currently **1.25.x** or newer in the same line).

Official downloads and docs: [https://go.dev/dl/](https://go.dev/dl/) · [https://go.dev/doc/install](https://go.dev/doc/install)

After installation, open a **new** terminal and check:

```bash
go version
```

You should see something like `go version go1.25.x ...`.

### Windows

1. Open [https://go.dev/dl/](https://go.dev/dl/) and download the **Windows** installer (`.msi`) for the latest **1.25** (or newer) release.
2. Run the installer, accept the defaults (it usually installs to `C:\Program Files\Go` and can add Go to your `PATH`).
3. Close and reopen **PowerShell** or **Command Prompt**, then run `go version`.
4. If `go` is not recognized, add `C:\Program Files\Go\bin` to your user **PATH** (Settings → System → About → Advanced system settings → Environment Variables).

### macOS

**Option A — Installer:** download the **macOS** `.pkg` from [https://go.dev/dl/](https://go.dev/dl/), run it, then open a new Terminal and run `go version`.

**Option B — Homebrew:**

```bash
brew install go
go version
```

(If you need exactly 1.25 and Homebrew lags, use the `.pkg` from go.dev.)

### Linux

**Option A — Tarball (works on most distributions):**

1. Download `go1.25.x.linux-amd64.tar.gz` (or `arm64` on ARM) from [https://go.dev/dl/](https://go.dev/dl/).
2. Remove any old install, then extract to `/usr/local` (needs `sudo`):

   ```bash
   sudo rm -rf /usr/local/go
   sudo tar -C /usr/local -xzf go1.25.x.linux-amd64.tar.gz
   ```

3. Add Go to your `PATH` (bash example — add to `~/.bashrc` or `~/.profile`):

   ```bash
   export PATH=$PATH:/usr/local/go/bin
   ```

4. Reload the shell or `source ~/.bashrc`, then `go version`.

**Option B — Package manager:** e.g. `sudo apt install golang-go` (Debian/Ubuntu) — versions vary; if the package is too old, use **Option A** or a [backports](https://go.dev/doc/install) / snap / official tarball.

---

## Build

With Go installed, get the code and compile.

**Clone (recommended):**

```bash
git clone https://github.com/dreanox/mcp-project-memory.git
cd mcp-project-memory
```

From the repository root:

### Windows (PowerShell)

```powershell
cd path\to\mcp-project-memory
mkdir bin -Force | Out-Null
go build -o bin/mcp-memory.exe ./cmd/mcp-memory
go build -o bin/pm.exe ./cmd/pm
```

### Windows (Command Prompt)

```bat
cd path\to\mcp-project-memory
mkdir bin 2>nul
go build -o bin\mcp-memory.exe .\cmd\mcp-memory
go build -o bin\pm.exe .\cmd\pm
```

### macOS / Linux (bash or zsh)

```bash
cd path/to/mcp-project-memory
mkdir -p bin
go build -o bin/mcp-memory ./cmd/mcp-memory
go build -o bin/pm ./cmd/pm
chmod +x bin/mcp-memory bin/pm   # optional
```

**Install binaries without keeping the repo** (downloads the module from GitHub; requires Go and a compatible version):

```bash
go install github.com/dreanox/mcp-project-memory/cmd/mcp-memory@latest
go install github.com/dreanox/mcp-project-memory/cmd/pm@latest
```

The executables go to **`$GOPATH/bin`** or **`$GOBIN`** — that directory must be on your **`PATH`**, or use the **full path** to `mcp-memory` (and `mcp-memory.exe` on Windows) in Cursor’s MCP `command` field.

**From a local checkout** (same as above but uses your tree):

```bash
go install ./cmd/mcp-memory
go install ./cmd/pm
```

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

### If I “install again”, do I lose memory from other projects?

**No — not from rebuilding or re-enabling the MCP.**

- **`go build` / `go install` again** only updates the **executable**. It does **not** touch `~/.cursor/project-memory/` (or whatever `PROJECT_MEMORY_ROOT` points to).
- **Editing Cursor MCP JSON** (path to the binary, `cwd`, `env`) does **not** delete stored JSON/MD files.
- **Each codebase** keeps its own subdirectory `{project-key}/` (derived from the **absolute** workspace path). Other projects’ folders are left alone when you use or reinstall the server for one repo.

**When you might see an empty memory (old data still on disk):**

- **`PROJECT_MEMORY_KEY` or `PROJECT_MEMORY_DIR` changed** → the server reads/writes a **different** folder; previous data remains in the old folder until you delete it manually.
- **Same project opened from a new path** (e.g. moved folder, new clone) → a **new** `{slug}-{hash}` key; the old key’s files are unchanged under `.cursor/project-memory/`.

**What actually deletes history:** manually removing `~/.cursor/project-memory` or individual `{project-key}/` folders (or wiping the machine).

## Cursor MCP

**Important workflow**

1. **Install once per machine** — Install Go, then build or `go install` the **`mcp-memory`** binary (see [Build](#build)). You do **not** need to rebuild for every repo; the same binary can serve all projects.
2. **Turn it on per project** — In Cursor, register the MCP server for each workspace (or use **User** MCP with `PROJECT_MEMORY_WORKSPACE` / per-project `mcp.json`) so **`cwd`** or **`env`** points at **that** project’s root. Memory files then live under `~/.cursor/project-memory/{key}/` **per project path**, not inside the git repo.
3. **Rule per project** — Copy [`.cursor/rules/project-memory-mcp.mdc`](.cursor/rules/project-memory-mcp.mdc) into **each** repo’s `.cursor/rules/` and **adapt** it (tone, when to call tools, `alwaysApply` / `globs`) to how that team works. The rule is only guidance text; tuning it is expected.

---

1. Build or `go install` the **`mcp-memory`** binary (see [Build](#build)).
2. In **Cursor → Settings → MCP** (or project `.cursor/mcp.json`), add a server whose **`command`** is the **absolute path** to that binary.
3. Set **`cwd`** to the **root of the project** where you want memory scoped (the open repo), **or** set **`PROJECT_MEMORY_WORKSPACE`** in `env` to that path.

Cursor does **not** install this server from the GitHub URL by itself: someone (you or the assistant) must place the binary and the JSON config as below.

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

### Cursor rule (example — when to use the tools)

The MCP only **exposes** tools; it does not force the model to call them. To get consistent behavior, add a **project rule** in **each** repository:

- **In this repo:** [`.cursor/rules/project-memory-mcp.mdc`](.cursor/rules/project-memory-mcp.mdc) (already present).
- **Elsewhere:** copy that file to `your-project/.cursor/rules/project-memory-mcp.mdc`, then **edit** it for that project’s needs (which tools, how often, transparency wording). Same install binary, new rule file per codebase if behaviors differ.

Details: [`.cursor/README.md`](.cursor/README.md).

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

## Feedback

Complaints, suggestions, or any other comments are welcome by email: **[drakgengard@gmail.com](mailto:drakgengard@gmail.com)**.

## License

MIT — see [`LICENSE`](LICENSE).
