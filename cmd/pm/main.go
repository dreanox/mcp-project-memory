// Command pm is the CLI for project memory (same storage as mcp-memory).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dreanox/mcp-project-memory/internal/svc"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	ws := os.Getenv("PROJECT_MEMORY_WORKSPACE")
	s, err := svc.New(ws)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch os.Args[1] {
	case "worklog":
		cmdWorklog(s, os.Args[2:])
	case "memory":
		cmdMemory(s, os.Args[2:])
	case "context":
		cmdContext(s, os.Args[2:])
	case "structure":
		cmdStructure(s, os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `pm — project memory CLI

Storage: %%USERPROFILE%%\.cursor\project-memory\{key}\  (override with PROJECT_MEMORY_DIR / PROJECT_MEMORY_ROOT / PROJECT_MEMORY_KEY)

Commands:
  pm worklog add --summary TEXT [--date YYYY-MM-DD] [--tool cursor|claude|manual] [--tags a,b] [--components a,b] [--notes TEXT]
  pm worklog list [--from DATE] [--to DATE]
  pm memory add --type TYPE --content TEXT [--scope SCOPE] [--tags a,b] [--id ID]
  pm memory search [--track project_memory|worklog|both] [--limit N] [query...]
  pm memory list
  pm memory delete ID
  pm memory update ID [--type T] [--scope S] [--content C] [--tags a,b]
  pm context [--structure] [--top-mem N] [--top-log N] <query...>
  pm structure show [focus]
  pm structure list
  pm structure map <request...>
`)
}

func cmdWorklog(s *svc.Service, args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}
	switch args[0] {
	case "add":
		fs := parseFlags(args[1:])
		summary := fs.get("summary", "")
		if summary == "" {
			fmt.Fprintln(os.Stderr, "worklog add requires --summary")
			os.Exit(1)
		}
		e, err := s.WorklogAppend(fs.get("date", ""), summary, fs.get("tool", ""), fs.getList("components"), fs.getList("tags"), fs.get("notes", ""))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		printJSON(e)
	case "list":
		fs := parseFlags(args[1:])
		list, err := s.WorklogList(fs.get("from", ""), fs.get("to", ""))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		printJSON(list)
	default:
		usage()
		os.Exit(1)
	}
}

func cmdMemory(s *svc.Service, args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}
	switch args[0] {
	case "add":
		fs := parseFlags(args[1:])
		content := fs.get("content", "")
		if content == "" {
			fmt.Fprintln(os.Stderr, "memory add requires --content")
			os.Exit(1)
		}
		m, err := s.ProjectMemoryAdd(fs.get("id", ""), fs.get("type", ""), fs.get("scope", ""), content, fs.getList("tags"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		printJSON(m)
	case "search":
		fs := parseFlags(args[1:])
		q := strings.Join(fs.rest, " ")
		mem, logs, notice, err := s.MemorySearch(fs.get("track", "both"), q, fs.getInt("limit", 20))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("notice:", notice)
		out := map[string]any{"project_memory": mem, "work_log": logs}
		printJSON(out)
	case "list":
		list, err := s.ProjectMemoryList()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		printJSON(list)
	case "delete":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "memory delete ID")
			os.Exit(1)
		}
		if err := s.ProjectMemoryDelete(args[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("deleted", args[1])
	case "update":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "memory update ID ...flags")
			os.Exit(1)
		}
		fs := parseFlags(args[2:])
		if err := s.ProjectMemoryUpdate(args[1], fs.get("type", ""), fs.get("scope", ""), fs.get("content", ""), fs.getListOptional("tags")); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("updated", args[1])
	default:
		usage()
		os.Exit(1)
	}
}

func cmdContext(s *svc.Service, args []string) {
	fs := parseFlags(args)
	q := strings.Join(fs.rest, " ")
	b, err := s.GetContext(q, fs.getInt("top-mem", 5), fs.getInt("top-log", 3), fs.has("structure"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(svc.FormatBundle(b))
}

func cmdStructure(s *svc.Service, args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}
	switch args[0] {
	case "show":
		focus := strings.TrimSpace(strings.Join(args[1:], " "))
		v, err := s.StructureShow(focus)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("notice:", v.Notice)
		fmt.Println(v.Body)
	case "list":
		frags, err := s.StructureList()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		printJSON(frags)
	case "map":
		req := strings.TrimSpace(strings.Join(args[1:], " "))
		v, err := s.StructureMap(req)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("notice:", v.Notice)
		fmt.Println(v.Body)
	default:
		usage()
		os.Exit(1)
	}
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// minimal flag parser: --key value or --key=value
type flagSet struct {
	m    map[string]string
	rest []string
}

func parseFlags(args []string) *flagSet {
	m := map[string]string{}
	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "--") {
			rest = append(rest, a)
			continue
		}
		a = strings.TrimPrefix(a, "--")
		key, val, ok := strings.Cut(a, "=")
		if ok {
			m[key] = val
			continue
		}
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			m[key] = args[i+1]
			i++
		} else {
			m[key] = "true"
		}
	}
	return &flagSet{m: m, rest: rest}
}

func (f *flagSet) get(key, def string) string {
	if v, ok := f.m[key]; ok {
		return v
	}
	return def
}

func (f *flagSet) getInt(key string, def int) int {
	s := f.get(key, "")
	if s == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

func (f *flagSet) has(key string) bool {
	_, ok := f.m[key]
	return ok
}

func (f *flagSet) getList(key string) []string {
	s := f.get(key, "")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// getListOptional returns nil if flag absent, else list (may be empty to clear)
func (f *flagSet) getListOptional(key string) []string {
	if _, ok := f.m[key]; !ok {
		return nil
	}
	return f.getList(key)
}
