package structure

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	GlobalFile = "project-structure.md"
	FragDir    = "structure"
)

var linkRe = regexp.MustCompile(`(?i)structure[/\\][a-z0-9_.-]+\.md`)

// Fragment describes one structure file.
type Fragment struct {
	Name string // e.g. api (basename without .md)
	Path string // relative to project dir: structure/api.md
}

// ListFragments returns structure/*.md basenames.
func ListFragments(projectDir string) ([]Fragment, error) {
	dir := filepath.Join(projectDir, FragDir)
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Fragment
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".md") {
			base := strings.TrimSuffix(name, filepath.Ext(name))
			out = append(out, Fragment{
				Name: base,
				Path: filepath.Join(FragDir, name),
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// View is a structure response with notice.
type View struct {
	Notice string
	Body   string
}

// GlobalView returns project-structure.md body plus TOC of fragments.
func GlobalView(projectDir string) (*View, error) {
	mainPath := filepath.Join(projectDir, GlobalFile)
	main, _ := os.ReadFile(mainPath)
	frags, err := ListFragments(projectDir)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	if len(main) > 0 {
		b.WriteString(string(main))
		b.WriteString("\n\n")
	} else {
		b.WriteString("*(No project-structure.md yet — create one in your project-memory folder.)*\n\n")
	}
	if len(frags) > 0 {
		b.WriteString("## Structure fragments\n\n")
		for _, f := range frags {
			b.WriteString("- `")
			b.WriteString(f.Path)
			b.WriteString("` (")
			b.WriteString(f.Name)
			b.WriteString(")\n")
		}
	}
	notice := "global: project-structure.md"
	if len(frags) > 0 {
		notice += fmtFragList(frags)
	}
	return &View{Notice: notice, Body: strings.TrimSpace(b.String())}, nil
}

func fmtFragList(frags []Fragment) string {
	if len(frags) == 0 {
		return ""
	}
	var names []string
	for _, f := range frags {
		names = append(names, f.Path)
	}
	return " + listed " + strings.Join(names, ", ")
}

// FocusView resolves slug or path to a single fragment or global file section.
func FocusView(projectDir, focus string) (*View, error) {
	focus = strings.TrimSpace(focus)
	if focus == "" {
		return GlobalView(projectDir)
	}
	// normalize: strip leading structure/
	focus = strings.TrimPrefix(focus, "/")
	if strings.HasPrefix(strings.ToLower(focus), "structure/") || strings.HasPrefix(strings.ToLower(focus), `structure\`) {
		p := filepath.Join(projectDir, filepath.FromSlash(focus))
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		return &View{
			Notice: "single-fragment: " + focus,
			Body:   string(data),
		}, nil
	}
	// slug -> structure/{slug}.md
	p := filepath.Join(projectDir, FragDir, focus+".md")
	if data, err := os.ReadFile(p); err == nil {
		return &View{
			Notice: "single-fragment: structure/" + focus + ".md",
			Body:   string(data),
		}, nil
	}
	// try global file if name matches
	if strings.EqualFold(focus, "global") || strings.EqualFold(focus, "index") {
		return GlobalView(projectDir)
	}
	return nil, os.ErrNotExist
}

// MapForRequest scores fragments + global file by keyword overlap; follows structure/*.md links.
func MapForRequest(projectDir, request string) (*View, error) {
	toks := tokenize(request)
	if len(toks) == 0 {
		v, err := GlobalView(projectDir)
		if err != nil {
			return nil, err
		}
		v.Notice = "global (vague request): " + v.Notice
		return v, nil
	}
	type scored struct {
		relPath string
		score   int
		body    string
	}
	var files []scored
	addFile := func(rel string) {
		p := filepath.Join(projectDir, filepath.FromSlash(rel))
		data, err := os.ReadFile(p)
		if err != nil {
			return
		}
		body := string(data)
		low := strings.ToLower(body + " " + rel)
		s := 0
		for _, t := range toks {
			if strings.Contains(low, t) {
				s++
			}
		}
		files = append(files, scored{relPath: rel, score: s, body: body})
	}
	if _, err := os.Stat(filepath.Join(projectDir, GlobalFile)); err == nil {
		addFile(GlobalFile)
	}
	frags, _ := ListFragments(projectDir)
	for _, f := range frags {
		addFile(f.Path)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].score > files[j].score })

	var picked []scored
	seen := map[string]bool{}
	var follow func(s scored)
	follow = func(s scored) {
		if seen[s.relPath] {
			return
		}
		seen[s.relPath] = true
		picked = append(picked, s)
		for _, m := range linkRe.FindAllString(s.body, -1) {
			rel := strings.ReplaceAll(m, `\`, `/`)
			p := filepath.Join(projectDir, filepath.FromSlash(rel))
			if data, err := os.ReadFile(p); err == nil {
				follow(scored{relPath: rel, score: 0, body: string(data)})
			}
		}
	}

	if len(files) == 0 {
		v, err := GlobalView(projectDir)
		if err != nil {
			return nil, err
		}
		v.Notice = "global (no structure files): " + v.Notice
		return v, nil
	}
	var positives []scored
	for _, f := range files {
		if f.score > 0 {
			positives = append(positives, f)
		}
	}
	if len(positives) == 0 {
		follow(files[0])
	} else {
		for i := 0; i < len(positives) && i < 3; i++ {
			follow(positives[i])
		}
	}
	var b strings.Builder
	var names []string
	for _, p := range picked {
		names = append(names, p.relPath)
		b.WriteString("### ")
		b.WriteString(p.relPath)
		b.WriteString("\n\n")
		b.WriteString(p.body)
		b.WriteString("\n\n")
	}
	kind := "composed"
	if len(picked) == 1 {
		kind = "single-fragment"
	}
	return &View{
		Notice: kind + ": " + strings.Join(names, ", "),
		Body:   strings.TrimSpace(b.String()),
	}, nil
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	var toks []string
	cur := strings.Builder{}
	flush := func() {
		if cur.Len() >= 2 {
			toks = append(toks, cur.String())
		}
		cur.Reset()
	}
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return toks
}
