package retrieve

import (
	"sort"
	"strings"

	"github.com/dreanox/mcp-project-memory/internal/models"
)

// ScoredMemory pairs an entry with its relevance score.
type ScoredMemory struct {
	Entry models.ProjectMemory
	Score int
}

// ScoreProjectMemory ranks memories: tag match +2, content +1 per token hit (simple).
func ScoreProjectMemory(query string, list []models.ProjectMemory) []ScoredMemory {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		var out []ScoredMemory
		for _, e := range list {
			out = append(out, ScoredMemory{Entry: e, Score: 0})
		}
		return out
	}
	tokens := tokenize(q)
	var out []ScoredMemory
	for _, e := range list {
		score := scoreEntry(tokens, strings.ToLower(e.Content), e.Tags, strings.ToLower(e.Type), strings.ToLower(e.Scope))
		if score > 0 || len(tokens) == 0 {
			out = append(out, ScoredMemory{Entry: e, Score: score})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Entry.CreatedAt.After(out[j].Entry.CreatedAt)
	})
	return out
}

func tokenize(s string) []string {
	var toks []string
	cur := strings.Builder{}
	flush := func() {
		if cur.Len() >= 2 {
			toks = append(toks, cur.String())
		}
		cur.Reset()
	}
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return toks
}

func scoreEntry(tokens []string, content string, tags []string, typ, scope string) int {
	score := 0
	for _, t := range tokens {
		if strings.Contains(content, t) {
			score++
		}
		if strings.Contains(typ, t) || strings.Contains(scope, t) {
			score++
		}
		for _, tag := range tags {
			if strings.Contains(strings.ToLower(tag), t) {
				score += 2
			}
		}
	}
	return score
}

// ScoredWorkLog is a work-log line with score.
type ScoredWorkLog struct {
	Entry models.WorkLogEntry
	Score int
}

// ScoreWorkLog ranks work log entries against query (summary, notes, tags, components).
func ScoreWorkLog(query string, list []models.WorkLogEntry) []ScoredWorkLog {
	q := strings.ToLower(strings.TrimSpace(query))
	tokens := tokenize(q)
	var out []ScoredWorkLog
	for _, e := range list {
		var parts []string
		parts = append(parts, strings.ToLower(e.Summary), strings.ToLower(e.Notes), strings.ToLower(e.Tool), strings.ToLower(e.Date))
		for _, c := range e.Components {
			parts = append(parts, strings.ToLower(c))
		}
		for _, t := range e.Tags {
			parts = append(parts, strings.ToLower(t))
		}
		blob := strings.Join(parts, " ")
		score := 0
		for _, t := range tokens {
			if strings.Contains(blob, t) {
				score++
			}
		}
		if score > 0 || q == "" {
			out = append(out, ScoredWorkLog{Entry: e, Score: score})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Entry.CreatedAt.After(out[j].Entry.CreatedAt)
	})
	return out
}
