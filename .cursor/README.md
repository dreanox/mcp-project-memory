# Cursor integration (this repository)

## Example rule: `rules/project-memory-mcp.mdc`

This file is a **template** you can reuse in **any project** where you enable the project-memory MCP.

### Use in another project (your app / library)

1. Install and register **mcp-memory** in Cursor (see the root [README.md](../README.md) → *Cursor MCP*).
2. Copy this file into that project:

   ```text
   from:  mcp-project-memory/.cursor/rules/project-memory-mcp.mdc
   to:    your-repo/.cursor/rules/project-memory-mcp.mdc
   ```

3. Restart Cursor or reload the window if rules do not pick up immediately.
4. Optional: set `alwaysApply: false` and add `globs` in the frontmatter if you only want the rule when certain files are open.

### Use only in this repo

If you are **developing** mcp-project-memory and find the rule noisy, set in the frontmatter:

```yaml
alwaysApply: false
```

and optionally scope with `globs` (e.g. `README.md` only), or disable the MCP entry while coding.

## What the rule does

It does **not** install the MCP. It tells the agent **when** to call `get_context`, structure tools, `worklog_append`, and `project_memory_add`, and to **surface the `notice`** to the user—aligned with the server’s design.
