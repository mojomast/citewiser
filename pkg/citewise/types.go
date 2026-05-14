package citewise

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Item is a single unread/read backlog entry. The struct intentionally accepts
// lightweight metadata because Citewise is a local triage tool, not a citation manager.
type Item struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Authors          []string `json:"authors,omitempty"`
	Source           string   `json:"source,omitempty"`
	Year             int      `json:"year,omitempty"`
	Type             string   `json:"type,omitempty"`
	URL              string   `json:"url,omitempty"`
	LengthMinutes    int      `json:"length_minutes,omitempty"`
	Difficulty       int      `json:"difficulty,omitempty"` // 1 easy .. 5 dense
	Status           string   `json:"status,omitempty"`     // unread, read, skipped
	Topics           []string `json:"topics,omitempty"`
	Notes            string   `json:"notes,omitempty"`
	GoalFit          float64  `json:"goal_fit,omitempty"` // optional 0..1 hint
	Energy           string   `json:"energy,omitempty"`   // low, medium, high
	Density          string   `json:"density,omitempty"`  // light, medium, dense
	Trust            float64  `json:"trust,omitempty"`    // 0..1 source confidence
	RecommendedCount int      `json:"recommended_count,omitempty"`
}

type Edge struct {
	SourceID   string  `json:"source_id"`
	TargetID   string  `json:"target_id"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence,omitempty"`
	Note       string  `json:"note,omitempty"`
}

type Goal struct {
	Prompt       string   `json:"prompt,omitempty"`
	Topics       []string `json:"topics,omitempty"`
	Familiarity  int      `json:"familiarity,omitempty"` // 1 novice .. 5 expert
	Energy       string   `json:"energy,omitempty"`
	TimeMinutes  int      `json:"time_minutes,omitempty"`
	DesiredDepth string   `json:"desired_depth,omitempty"`
}

type Backlog struct {
	Items []Item `json:"items"`
	Edges []Edge `json:"edges,omitempty"`
	Goal  Goal   `json:"goal,omitempty"`
}

func LoadBacklog(path string) (Backlog, error) {
	f, err := os.Open(path)
	if err != nil {
		return Backlog{}, err
	}
	defer f.Close()
	return ParseBacklog(f, filepath.Ext(path))
}

func ParseBacklog(r io.Reader, ext string) (Backlog, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return Backlog{}, err
	}
	trim := strings.TrimSpace(string(b))
	if trim == "" {
		return Backlog{}, errors.New("empty backlog")
	}
	if strings.EqualFold(ext, ".json") || strings.HasPrefix(trim, "{") {
		var backlog Backlog
		if err := json.Unmarshal([]byte(trim), &backlog); err != nil {
			var items []Item
			if err2 := json.Unmarshal([]byte(trim), &items); err2 == nil {
				backlog.Items = items
			} else {
				return Backlog{}, err
			}
		}
		return normalizeBacklog(backlog)
	}
	return parseCSV(strings.NewReader(trim))
}

func parseCSV(r io.Reader) (Backlog, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	records, err := cr.ReadAll()
	if err != nil {
		return Backlog{}, err
	}
	if len(records) < 2 {
		return Backlog{}, errors.New("csv needs a header and at least one row")
	}
	head := map[string]int{}
	for i, h := range records[0] {
		head[strings.ToLower(strings.TrimSpace(h))] = i
	}
	get := func(row []string, names ...string) string {
		for _, n := range names {
			if idx, ok := head[n]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
		}
		return ""
	}
	var backlog Backlog
	for _, row := range records[1:] {
		recType := strings.ToLower(get(row, "record_type", "kind"))
		if recType == "edge" || get(row, "edge_type", "type") != "" && get(row, "source_id") != "" && get(row, "target_id") != "" {
			backlog.Edges = append(backlog.Edges, Edge{SourceID: get(row, "source_id"), TargetID: get(row, "target_id"), Type: get(row, "edge_type", "type"), Confidence: parseFloat(get(row, "confidence"))})
			continue
		}
		it := Item{ID: get(row, "id"), Title: get(row, "title"), Source: get(row, "source"), Type: get(row, "type"), URL: get(row, "url"), Status: get(row, "status"), Notes: get(row, "notes"), Energy: get(row, "energy"), Density: get(row, "density")}
		it.Year = parseInt(get(row, "year"))
		it.LengthMinutes = parseInt(get(row, "length_minutes", "minutes", "length"))
		it.Difficulty = parseInt(get(row, "difficulty"))
		it.GoalFit = parseFloat(get(row, "goal_fit"))
		it.Trust = parseFloat(get(row, "trust"))
		it.RecommendedCount = parseInt(get(row, "recommended_count", "recommendations"))
		it.Authors = splitList(get(row, "authors", "creators"))
		it.Topics = splitList(get(row, "topics", "tags"))
		backlog.Items = append(backlog.Items, it)
	}
	return normalizeBacklog(backlog)
}

func normalizeBacklog(b Backlog) (Backlog, error) {
	seen := map[string]bool{}
	for i := range b.Items {
		b.Items[i].ID = strings.TrimSpace(b.Items[i].ID)
		if b.Items[i].ID == "" {
			b.Items[i].ID = slugID(b.Items[i].Title, i+1)
		}
		if b.Items[i].Title == "" {
			return Backlog{}, fmt.Errorf("item %q is missing title", b.Items[i].ID)
		}
		if seen[b.Items[i].ID] {
			return Backlog{}, fmt.Errorf("duplicate item id %q", b.Items[i].ID)
		}
		seen[b.Items[i].ID] = true
		if b.Items[i].Status == "" {
			b.Items[i].Status = "unread"
		}
		if b.Items[i].Difficulty == 0 {
			b.Items[i].Difficulty = 3
		}
		if b.Items[i].LengthMinutes == 0 {
			b.Items[i].LengthMinutes = 45
		}
	}
	for i := range b.Edges {
		b.Edges[i].Type = normEdgeType(b.Edges[i].Type)
		if b.Edges[i].Confidence == 0 {
			b.Edges[i].Confidence = 1
		}
	}
	return b, nil
}

func parseInt(s string) int       { v, _ := strconv.Atoi(strings.TrimSpace(s)); return v }
func parseFloat(s string) float64 { v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64); return v }

func splitList(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ';' || r == '|' || r == ',' })
	var out []string
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func slugID(title string, n int) string {
	s := normalizeTitle(title)
	s = strings.ReplaceAll(s, " ", "-")
	if s == "" {
		s = "item"
	}
	if len(s) > 32 {
		s = s[:32]
	}
	return fmt.Sprintf("%s-%d", s, n)
}

func normEdgeType(s string) string {
	s = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(s, "_", "-")))
	s = strings.ReplaceAll(s, " ", "-")
	switch s {
	case "same-question", "same-topic", "similar":
		return "same-question"
	case "duplicates", "duplicate-of", "dupe":
		return "duplicate"
	case "contradicts", "counterpoint", "responds-to":
		return "contradicts"
	case "cites", "mentions", "mention":
		return "cites"
	case "prereq", "requires":
		return "prerequisite"
	case "review", "surveys":
		return "overview"
	}
	return s
}
