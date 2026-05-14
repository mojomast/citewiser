package citewise

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		printHelp(stdout)
		return 0
	}
	cmd := args[0]
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(stderr)
	file := fs.String("file", "", "JSON or CSV backlog file")
	goalPrompt := fs.String("goal", "", "reading goal prompt or topic words")
	budget := fs.Int("budget", 180, "available reading minutes for queue planning")
	limit := fs.Int("limit", 5, "maximum queue items")
	id := fs.String("id", "", "item id for explain")
	format := fs.String("format", "text", "output format: text, json, markdown")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if *file == "" {
		fmt.Fprintln(stderr, "citewise: --file is required")
		return 2
	}
	backlog, err := LoadBacklog(*file)
	if err != nil {
		fmt.Fprintln(stderr, "citewise:", err)
		return 1
	}
	goal := backlog.Goal
	if *goalPrompt != "" {
		goal.Prompt = *goalPrompt
		goal.Topics = splitList(*goalPrompt)
	}
	if goal.TimeMinutes == 0 {
		goal.TimeMinutes = *budget
	}
	a := Analyze(backlog, goal)
	switch cmd {
	case "roles":
		writeRoles(stdout, a, *format)
	case "score", "scores":
		writeScores(stdout, a, *format)
	case "queue", "plan":
		writeQueue(stdout, PlanQueue(a, *budget, *limit), *format)
	case "explain":
		if *id == "" {
			fmt.Fprintln(stderr, "citewise: explain requires --id")
			return 2
		}
		it, ok := a.Items[*id]
		if !ok {
			fmt.Fprintf(stderr, "citewise: unknown item id %q\n", *id)
			return 1
		}
		writeExplain(stdout, it, a.Scores[*id], *format)
	case "hygiene", "duplicates":
		writeHygiene(stdout, Hygiene(a), *format)
	case "export":
		writeMarkdownReport(stdout, a, PlanQueue(a, *budget, *limit), Hygiene(a))
	default:
		fmt.Fprintf(stderr, "citewise: unknown command %q\n\n", cmd)
		printHelp(stderr)
		return 2
	}
	return 0
}

func printHelp(w io.Writer) {
	fmt.Fprint(w, `Citewise — explainable reading backlog triage

Usage:
  citewise <command> --file backlog.json [options]

Commands:
  roles       classify every item as foundation, overview, bridge, counterpoint, stale-hype, duplicate, or curiosity-leaf
  score       score unread items using goal fit, centrality, readiness, freshness, redundancy, and energy/time fit
  queue       produce a short read-next queue with rationales and a time budget
  explain     explain one item with --id ITEM_ID
  hygiene     find duplicate clusters, orphan items, stale items, and missing prerequisite bridges
  export      write a friendly Markdown report

Options:
  --file PATH       JSON or CSV backlog file
  --goal TEXT       current reading goal/topic words
  --budget MIN      available minutes for queue/export (default 180)
  --limit N         max queue items (default 5)
  --id ITEM_ID      item id for explain
  --format FORMAT   text, json, or markdown where supported

Backlog JSON shape: {"items":[...], "edges":[...], "goal":{...}}
CSV supports item columns such as id,title,year,type,topics,length_minutes,difficulty,goal_fit and edge rows with record_type=edge,source_id,target_id,edge_type.
`)
}

func writeRoles(w io.Writer, a Analysis, format string) {
	rows := unreadItems(sortedItems(a))
	if format == "json" {
		roles := map[string]string{}
		for _, it := range rows {
			roles[it.ID] = a.Roles[it.ID]
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(roles)
		return
	}
	fmt.Fprintln(w, "Role map")
	for _, it := range rows {
		fmt.Fprintf(w, "- %s: %s — %s\n", it.ID, a.Roles[it.ID], it.Title)
	}
}

func unreadItems(items []Item) []Item {
	var out []Item
	for _, it := range items {
		if strings.ToLower(it.Status) != "read" {
			out = append(out, it)
		}
	}
	return out
}

func writeScores(w io.Writer, a Analysis, format string) {
	items := sortedItems(a)
	if format == "json" {
		var scores []Score
		for _, it := range items {
			scores = append(scores, a.Scores[it.ID])
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(scores)
		return
	}
	fmt.Fprintln(w, "Reading priority scores")
	for _, it := range sortByScore(items, a) {
		s := a.Scores[it.ID]
		fmt.Fprintf(w, "%5.1f  %-14s %-14s %s\n", s.Total, s.Role, it.ID, it.Title)
	}
}

func writeQueue(w io.Writer, p QueuePlan, format string) {
	if format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(p)
		return
	}
	fmt.Fprintf(w, "Read-next queue (%d minute budget)\n", p.BudgetMinutes)
	for i, e := range p.Entries {
		fmt.Fprintf(w, "%d. %s (%s, %d min, score %.1f)\n   Why: %s\n", i+1, e.Item.Title, e.Score.Role, e.Item.LengthMinutes, e.Score.Total, e.Rationale)
	}
	if len(p.Skipped) > 0 {
		fmt.Fprintln(w, "\nPark for later:")
		for _, e := range p.Skipped {
			fmt.Fprintf(w, "- %s: %s\n", e.Item.Title, skipReason(e))
		}
	}
}

func writeExplain(w io.Writer, it Item, s Score, format string) {
	if format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(s)
		return
	}
	fmt.Fprintf(w, "%s\nRole: %s\nScore: %.1f\n", it.Title, s.Role, s.Total)
	for _, r := range s.Rationale {
		fmt.Fprintf(w, "- %s\n", r)
	}
	fmt.Fprintf(w, "Plain English: %s\n", ExplainScore(it, s))
}

func writeHygiene(w io.Writer, h HygieneReport, format string) {
	if format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(h)
		return
	}
	fmt.Fprintln(w, "Backlog hygiene")
	fmt.Fprintf(w, "Duplicate clusters: %s\n", clustersText(h.DuplicateClusters))
	fmt.Fprintf(w, "Orphan items: %s\n", listText(h.OrphanItems))
	fmt.Fprintf(w, "Stale/hype cautions: %s\n", listText(h.StaleItems))
	fmt.Fprintf(w, "Missing prerequisite bridges: %s\n", listText(h.MissingBridges))
}

func writeMarkdownReport(w io.Writer, a Analysis, p QueuePlan, h HygieneReport) {
	fmt.Fprintln(w, "# Citewise Reading Triage Report")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Read Next")
	for i, e := range p.Entries {
		fmt.Fprintf(w, "%d. **%s** — %s, score %.1f (%d min)  \n   %s\n", i+1, e.Item.Title, e.Score.Role, e.Score.Total, e.Item.LengthMinutes, e.Rationale)
	}
	fmt.Fprintln(w, "\n## Role Map")
	for _, it := range sortedItems(a) {
		fmt.Fprintf(w, "- **%s** (`%s`): %s\n", it.Title, it.ID, a.Roles[it.ID])
	}
	fmt.Fprintln(w, "\n## Hygiene")
	fmt.Fprintf(w, "- Duplicate clusters: %s\n- Orphan items: %s\n- Stale/hype cautions: %s\n- Missing prerequisite bridges: %s\n", clustersText(h.DuplicateClusters), listText(h.OrphanItems), listText(h.StaleItems), listText(h.MissingBridges))
}

func ExplainScore(it Item, s Score) string {
	parts := []string{fmt.Sprintf("%s is a %s", it.Title, s.Role)}
	if s.GoalFit >= .7 {
		parts = append(parts, "strongly matches the goal")
	}
	if s.Centrality >= .7 {
		parts = append(parts, "unlocks several connected items")
	}
	if s.Readiness < .55 {
		parts = append(parts, "may need prerequisites first")
	}
	if s.RedundancyPenalty >= .5 {
		parts = append(parts, "looks redundant with another backlog item")
	}
	if s.EnergyTimeFit >= .8 {
		parts = append(parts, "fits the current time/energy budget")
	}
	return strings.Join(parts, "; ") + "."
}

func sortedItems(a Analysis) []Item {
	out := append([]Item(nil), a.Backlog.Items...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func sortByScore(items []Item, a Analysis) []Item {
	out := append([]Item(nil), items...)
	sort.SliceStable(out, func(i, j int) bool { return a.Scores[out[i].ID].Total > a.Scores[out[j].ID].Total })
	return out
}
func listText(xs []string) string {
	if len(xs) == 0 {
		return "none"
	}
	return strings.Join(xs, ", ")
}
func clustersText(cs [][]string) string {
	if len(cs) == 0 {
		return "none"
	}
	var p []string
	for _, c := range cs {
		p = append(p, "["+strings.Join(c, ", ")+"]")
	}
	return strings.Join(p, "; ")
}
func skipReason(e QueueEntry) string {
	if e.Score.Role == RoleDuplicate {
		return "duplicate or redundant"
	}
	if e.Item.LengthMinutes > 0 {
		return "outside this queue after stronger fits"
	}
	return "lower priority for this goal"
}
