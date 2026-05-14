package citewise

import (
	"bytes"
	"strings"
	"testing"
)

func fixtureBacklog() Backlog {
	return Backlog{
		Goal: Goal{Prompt: "urban trees cooling", Topics: []string{"urban", "trees", "cooling"}, Familiarity: 2, Energy: "medium", TimeMinutes: 150},
		Items: []Item{
			{ID: "classic", Title: "Classic Foundations of Urban Heat Islands", Year: 1998, Type: "book", Topics: []string{"urban", "climate"}, LengthMinutes: 80, Difficulty: 2, GoalFit: .80},
			{ID: "review", Title: "A Review and Primer on Urban Tree Cooling", Year: 2023, Type: "review", Topics: []string{"urban", "trees", "cooling"}, LengthMinutes: 35, Difficulty: 2, GoalFit: .95, Energy: "medium"},
			{ID: "bridge", Title: "From Street Design to Canopy Equity", Year: 2022, Type: "article", Topics: []string{"urban", "trees", "equity"}, LengthMinutes: 45, Difficulty: 2, GoalFit: .75},
			{ID: "counter", Title: "Counterpoint: Shade Claims Overstate Benefits", Year: 2021, Type: "article", Topics: []string{"cooling"}, LengthMinutes: 30, Difficulty: 3, GoalFit: .55},
			{ID: "hype", Title: "Hot Trend Breakthrough App for City Trees", Year: 2017, Type: "newsletter", Topics: []string{"trees"}, LengthMinutes: 20, Difficulty: 1, GoalFit: .35, RecommendedCount: 4},
			{ID: "leaf", Title: "A Curious History of Park Benches", Year: 2020, Type: "essay", Topics: []string{"benches"}, LengthMinutes: 20, Difficulty: 1, GoalFit: .10},
			{ID: "dupe1", Title: "Urban Tree Cooling Primer", Year: 2023, Type: "article", Topics: []string{"trees", "cooling"}, LengthMinutes: 25, Difficulty: 1, GoalFit: .8},
			{ID: "dupe2", Title: "The Urban Tree Cooling Primer", Year: 2023, Type: "article", Topics: []string{"trees", "cooling"}, LengthMinutes: 25, Difficulty: 1, GoalFit: .8},
			{ID: "advanced", Title: "Advanced Boundary Layer Model", Year: 2024, Type: "paper", Topics: []string{"urban", "physics"}, LengthMinutes: 120, Difficulty: 5, GoalFit: .7},
		},
		Edges: []Edge{
			{SourceID: "classic", TargetID: "advanced", Type: "prerequisite"},
			{SourceID: "review", TargetID: "classic", Type: "cites"},
			{SourceID: "bridge", TargetID: "classic", Type: "cites"},
			{SourceID: "bridge", TargetID: "counter", Type: "contradicts"},
			{SourceID: "review", TargetID: "advanced", Type: "overview"},
			{SourceID: "dupe1", TargetID: "dupe2", Type: "duplicate"},
		},
	}
}

func TestParseBacklogJSON(t *testing.T) {
	json := `{"items":[{"id":"a","title":"A Review","topics":["x"]}],"edges":[{"source_id":"a","target_id":"a","type":"cites"}]}`
	b, err := ParseBacklog(strings.NewReader(json), ".json")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Items) != 1 || b.Items[0].ID != "a" || b.Items[0].Status != "unread" {
		t.Fatalf("unexpected parse: %+v", b.Items)
	}
}

func TestParseBacklogJSONArrayReturnsArrayError(t *testing.T) {
	_, err := ParseBacklog(strings.NewReader(`[{"title":1}]`), ".json")
	if err == nil {
		t.Fatal("expected malformed array error")
	}
	if !strings.Contains(err.Error(), "Item.title") || strings.Contains(err.Error(), "Backlog") {
		t.Fatalf("got error %q, want array unmarshal error", err.Error())
	}
}

func TestSlugIDLongTitleStaysShort(t *testing.T) {
	b, err := ParseBacklog(strings.NewReader(`[{"title":"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefgh"}]`), ".json")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Items) != 1 {
		t.Fatalf("unexpected item count %d", len(b.Items))
	}
	if len(b.Items[0].ID) >= 35 {
		t.Fatalf("slug ID too long: %q (%d)", b.Items[0].ID, len(b.Items[0].ID))
	}
}

func TestParseBacklogCSVItemsAndEdges(t *testing.T) {
	csv := "record_type,id,title,year,type,topics,source_id,target_id,edge_type\nitem,a,Primer,2024,review,trees,,,\nitem,b,Deep Paper,2024,paper,cooling,,,\nedge,,,,,,a,b,cites\n"
	b, err := ParseBacklog(strings.NewReader(csv), ".csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Items) != 2 || len(b.Edges) != 1 || b.Edges[0].Type != "cites" {
		t.Fatalf("bad csv parse: %+v", b)
	}
}

func TestNormalizeBacklogClampsDifficultyAndTrust(t *testing.T) {
	b, err := ParseBacklog(strings.NewReader(`{"items":[{"id":"low","title":"Low","difficulty":-4,"trust":-0.5},{"id":"high","title":"High","difficulty":8,"trust":1.7}]}`), ".json")
	if err != nil {
		t.Fatal(err)
	}
	if b.Items[0].Difficulty != 1 || b.Items[0].Trust != 0 {
		t.Fatalf("low item not clamped: %+v", b.Items[0])
	}
	if b.Items[1].Difficulty != 5 || b.Items[1].Trust != 1 {
		t.Fatalf("high item not clamped: %+v", b.Items[1])
	}
}

func TestRoleClassificationTable(t *testing.T) {
	a := Analyze(fixtureBacklog(), fixtureBacklog().Goal)
	cases := map[string]string{"classic": RoleFoundation, "review": RoleOverview, "counter": RoleCounterpoint, "hype": RoleStaleHype, "leaf": RoleCuriosityLeaf, "dupe1": RoleDuplicate, "dupe2": RoleDuplicate}
	for id, want := range cases {
		if got := a.Roles[id]; got != want {
			t.Fatalf("%s role got %s want %s", id, got, want)
		}
	}
}

func TestBridgeRoleDetected(t *testing.T) {
	a := Analyze(fixtureBacklog(), fixtureBacklog().Goal)
	if got := a.Roles["bridge"]; got != RoleBridge {
		t.Fatalf("bridge role got %s", got)
	}
}

func TestScoringPenalizesDuplicates(t *testing.T) {
	a := Analyze(fixtureBacklog(), fixtureBacklog().Goal)
	if a.Scores["dupe1"].RedundancyPenalty < .5 {
		t.Fatalf("duplicate penalty too low: %+v", a.Scores["dupe1"])
	}
	if a.Scores["dupe1"].Total >= a.Scores["review"].Total {
		t.Fatalf("duplicate outranked review: dupe %.1f review %.1f", a.Scores["dupe1"].Total, a.Scores["review"].Total)
	}
}

func TestExplainScorePhrases(t *testing.T) {
	cases := []struct {
		name  string
		score Score
		want  string
	}{
		{name: "duplicate", score: Score{Role: RoleDuplicate, RedundancyPenalty: .75}, want: "looks redundant"},
		{name: "low readiness", score: Score{Role: RoleFoundation, Readiness: .40}, want: "may need prerequisites first"},
		{name: "goal fit", score: Score{Role: RoleFoundation, GoalFit: .70}, want: "strongly matches the goal"},
		{name: "centrality", score: Score{Role: RoleBridge, Centrality: .70}, want: "unlocks several connected items"},
		{name: "energy time", score: Score{Role: RoleOverview, EnergyTimeFit: .80}, want: "fits the current time/energy budget"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExplainScore(Item{Title: "Test Item"}, tc.score)
			if !strings.Contains(got, tc.want) {
				t.Fatalf("ExplainScore() = %q, want phrase %q", got, tc.want)
			}
		})
	}
}

func TestReadinessAccountsForUnreadPrerequisites(t *testing.T) {
	a := Analyze(fixtureBacklog(), fixtureBacklog().Goal)
	if a.Scores["advanced"].Readiness >= .75 {
		t.Fatalf("advanced should not be fully ready: %+v", a.Scores["advanced"])
	}
}

func TestQueuePlanningHonorsBudgetAndLimit(t *testing.T) {
	a := Analyze(fixtureBacklog(), fixtureBacklog().Goal)
	p := PlanQueue(a, 90, 3)
	if len(p.Entries) == 0 || len(p.Entries) > 3 {
		t.Fatalf("bad queue length %d", len(p.Entries))
	}
	total := 0
	for _, e := range p.Entries {
		total += e.Item.LengthMinutes
	}
	if total > 90 && len(p.Entries) > 1 {
		t.Fatalf("queue exceeded budget: %d", total)
	}
	if p.Entries[0].Item.ID != "review" {
		t.Fatalf("expected review first, got %s", p.Entries[0].Item.ID)
	}
}

func TestPlanQueueSignalsBudgetExceededForFirstItemOverride(t *testing.T) {
	a := Analyze(Backlog{Items: []Item{{ID: "long", Title: "Long", LengthMinutes: 120, GoalFit: 1}}}, Goal{})
	p := PlanQueue(a, 30, 1)
	if len(p.Entries) != 1 || p.Entries[0].Item.ID != "long" {
		t.Fatalf("expected forced first item, got %+v", p.Entries)
	}
	if !p.BudgetExceeded {
		t.Fatalf("expected budget exceeded signal: %+v", p)
	}
}

func TestWriteRolesSkipsReadItems(t *testing.T) {
	a := Analyze(Backlog{Items: []Item{{ID: "read", Title: "Read", Status: "read"}, {ID: "unread", Title: "Unread", Status: "unread"}}}, Goal{})
	var out bytes.Buffer
	writeRoles(&out, a, "text")
	got := out.String()
	if strings.Contains(got, "- read:") || !strings.Contains(got, "- unread:") {
		t.Fatalf("roles output did not filter read items: %q", got)
	}
}

func TestDuplicateDetectionByTitleSimilarity(t *testing.T) {
	clusters := DetectDuplicates(fixtureBacklog())
	found := false
	for _, c := range clusters {
		if strings.Join(c, ",") == "dupe1,dupe2" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected dupe cluster, got %+v", clusters)
	}
}

func TestHygieneFindsStaleAndOrphans(t *testing.T) {
	a := Analyze(fixtureBacklog(), fixtureBacklog().Goal)
	h := Hygiene(a)
	if !contains(h.StaleItems, "hype") {
		t.Fatalf("stale missing: %+v", h.StaleItems)
	}
	if !contains(h.OrphanItems, "leaf") {
		t.Fatalf("orphan missing: %+v", h.OrphanItems)
	}
}

func TestExportMarkdownIncludesQueueAndHygiene(t *testing.T) {
	a := Analyze(fixtureBacklog(), fixtureBacklog().Goal)
	var out bytes.Buffer
	writeMarkdownReport(&out, a, PlanQueue(a, 120, 4), Hygiene(a))
	s := out.String()
	if !strings.Contains(s, "# Citewise Reading Triage Report") || !strings.Contains(s, "## Hygiene") {
		t.Fatalf("bad markdown: %s", s)
	}
}

func TestCLIHelpAndScore(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run([]string{"--help"}, &out, &errb); code != 0 {
		t.Fatalf("help code %d err %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "Citewise") {
		t.Fatalf("help missing Citewise: %s", out.String())
	}
}

func TestCLIRequiresFile(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run([]string{"score"}, &out, &errb); code != 2 {
		t.Fatalf("want usage error, got %d", code)
	}
	if !strings.Contains(errb.String(), "--file is required") {
		t.Fatalf("unexpected stderr: %s", errb.String())
	}
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
