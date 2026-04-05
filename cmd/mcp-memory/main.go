// Command mcp-memory runs the Project Memory MCP server (stdio).
package main

import (
	"context"
	"log"
	"os"

	"github.com/master/mcp-memory/internal/mcpapp"
	"github.com/master/mcp-memory/internal/svc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverInstructions = `Project Memory: structured work log + project memory + optional architecture markdown stored under the user's ~/.cursor/project-memory/{project-key}/ (not in the git repo).

Always tell the user when you used this server's tools and repeat the **notice** field from tool results—especially for get_context, get_project_structure, and get_structure_for_request.

Workspace root: set env PROJECT_MEMORY_WORKSPACE to the opened project, or the server uses the process working directory (Cursor usually sets cwd to the workspace).`

func main() {
	log.SetOutput(os.Stderr)
	ws := os.Getenv("PROJECT_MEMORY_WORKSPACE")
	if ws == "" {
		var err error
		ws, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}
	s, err := svc.New(ws)
	if err != nil {
		log.Fatal(err)
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "project-memory", Version: "0.1.0"}, &mcp.ServerOptions{
		Instructions: serverInstructions,
	})
	mcpapp.RegisterTools(server, s)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
