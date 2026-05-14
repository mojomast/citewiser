package citewise

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	RoleFoundation    = "foundation"
	RoleOverview      = "overview"
	RoleBridge        = "bridge"
	RoleCounterpoint  = "counterpoint"
	RoleStaleHype     = "stale-hype"
	RoleDuplicate     = "duplicate"
	RoleCuriosityLeaf = "curiosity-leaf"
)

const (
	staleHypeMinAge    = 7
	staleCurrentMinAge = 12
)

type Analysis struct {
	Backlog    Backlog
	Items      map[string]Item
	EdgesOut   map[string][]Edge
	EdgesIn    map[string][]Edge
	Centrality map[string]float64
	Roles      map[string]string
	Scores     map[string]Score
	Duplicates [][]string
	NowYear    int
}

type Score struct {
	ItemID            string
	Total             float64
	GoalFit           float64
	Centrality        float64
	Readiness         float64
	Freshness         float64
	RedundancyPenalty float64
	EnergyTimeFit     float64
	Role              string
	Rationale         []string
}

type QueueEntry struct {
	Item      Item
	Score     Score
	Rationale string
}

type QueuePlan struct {
	Entries       []QueueEntry
	Skipped       []QueueEntry
	BudgetMinutes int
}

type HygieneReport struct {
	DuplicateClusters [][]string
	OrphanItems       []string
	StaleItems        []string
	MissingBridges    []string
}

func Analyze(b Backlog, goal Goal) Analysis {
	if goal.Prompt == "" && len(goal.Topics) == 0 {
		goal = b.Goal
	}
	if goal.Familiarity == 0 {
		goal.Familiarity = b.Goal.Familiarity
	}
	if goal.Energy == "" {
		goal.Energy = b.Goal.Energy
	}
	items := map[string]Item{}
	out := map[string][]Edge{}
	in := map[string][]Edge{}
	for _, it := range b.Items {
		items[it.ID] = it
	}
	for _, e := range b.Edges {
		if _, ok := items[e.SourceID]; !ok {
			continue
		}
		if _, ok := items[e.TargetID]; !ok {
			continue
		}
		out[e.SourceID] = append(out[e.SourceID], e)
		in[e.TargetID] = append(in[e.TargetID], e)
	}
	a := Analysis{Backlog: b, Items: items, EdgesOut: out, EdgesIn: in, Centrality: map[string]float64{}, Roles: map[string]string{}, Scores: map[string]Score{}, NowYear: time.Now().Year()}
	a.Centrality = centrality(b.Items, out, in)
	a.Duplicates = DetectDuplicates(b)
	for _, it := range b.Items {
		a.Roles[it.ID] = classify(it, a, goal)
	}
	for _, it := range b.Items {
		a.Scores[it.ID] = scoreItem(it, a, goal)
	}
	return a
}

func centrality(items []Item, out, in map[string][]Edge) map[string]float64 {
	raw := map[string]float64{}
	max := 0.0
	for _, it := range items {
		v := 0.0
		for _, e := range in[it.ID] {
			v += edgeWeight(e.Type) * e.Confidence
		}
		for _, e := range out[it.ID] {
			v += 0.65 * edgeWeight(e.Type) * e.Confidence
		}
		raw[it.ID] = v
		if v > max {
			max = v
		}
	}
	for k, v := range raw {
		if max == 0 {
			raw[k] = 0
		} else {
			raw[k] = clamp01(v / max)
		}
	}
	return raw
}

func edgeWeight(t string) float64 {
	switch normEdgeType(t) {
	case "prerequisite":
		return 1.5
	case "cites":
		return 1.1
	case "overview":
		return 1.25
	case "contradicts":
		return 1.0
	case "same-question":
		return .8
	case "duplicate":
		return .2
	default:
		return .6
	}
}

func classify(it Item, a Analysis, goal Goal) string {
	// Edges are pre-normalized by normalizeBacklog; normEdgeType here is
	// a safety net for any manually constructed Analysis values in tests.
	if inDupCluster(it.ID, a.Duplicates) || hasEdgeType(it.ID, a, "duplicate") {
		return RoleDuplicate
	}
	if hasIncomingEdgeType(it.ID, a, "contradicts") || strings.Contains(strings.ToLower(it.Title+" "+it.Notes), "counterpoint") {
		return RoleCounterpoint
	}
	if isOverview(it) || hasEdgeType(it.ID, a, "overview") {
		return RoleOverview
	}
	if isStaleHype(it, a.NowYear) {
		return RoleStaleHype
	}
	if foundationScore(it, a) >= 0.6 {
		return RoleFoundation
	}
	if bridgeScore(it, a) >= 0.55 {
		return RoleBridge
	}
	if degree(it.ID, a) <= 1 && goalFit(it, goal) < 0.45 {
		return RoleCuriosityLeaf
	}
	if a.Centrality[it.ID] >= 0.7 {
		return RoleFoundation
	}
	return RoleCuriosityLeaf
}

func hasEdgeType(id string, a Analysis, typ string) bool {
	for _, e := range append(a.EdgesIn[id], a.EdgesOut[id]...) {
		if normEdgeType(e.Type) == typ {
			return true
		}
	}
	return false
}

func hasIncomingEdgeType(id string, a Analysis, typ string) bool {
	for _, e := range a.EdgesIn[id] {
		if normEdgeType(e.Type) == typ {
			return true
		}
	}
	return false
}

func isOverview(it Item) bool {
	s := strings.ToLower(it.Type + " " + it.Title + " " + it.Notes)
	for _, cue := range []string{"review", "overview", "survey", "primer", "introduction", "guide", "handbook"} {
		if strings.Contains(s, cue) {
			return true
		}
	}
	return false
}

func isStaleHype(it Item, now int) bool {
	text := strings.ToLower(it.Title + " " + it.Notes)
	hypeCue := strings.Contains(text, "hype") || strings.Contains(text, "trend") || strings.Contains(text, "hot take") || strings.Contains(text, "breakthrough")
	if it.Year > 0 && now-it.Year >= staleHypeMinAge && (hypeCue || it.RecommendedCount >= 3) {
		return true
	}
	if it.Year > 0 && now-it.Year >= staleCurrentMinAge && strings.Contains(text, "current") {
		return true
	}
	return false
}

func foundationScore(it Item, a Analysis) float64 {
	incomingPrereq := 0
	incomingCites := 0
	for _, e := range a.EdgesIn[it.ID] {
		if e.Type == "prerequisite" {
			incomingPrereq++
		}
	}
	for _, e := range a.EdgesIn[it.ID] {
		if e.Type == "cites" || e.Type == "prerequisite" {
			incomingCites++
		}
	}
	age := 0.0
	if it.Year > 0 && a.NowYear-it.Year >= 5 {
		age = .2
	}
	cue := 0.0
	if strings.Contains(strings.ToLower(it.Type+" "+it.Title+" "+it.Notes), "foundation") || strings.Contains(strings.ToLower(it.Type+" "+it.Title+" "+it.Notes), "classic") {
		cue = .25
	}
	return clamp01(a.Centrality[it.ID]*.55 + math.Min(1, float64(incomingPrereq+incomingCites)/3)*.35 + age + cue)
}

func bridgeScore(it Item, a Analysis) float64 {
	topicSet := map[string]bool{}
	for _, t := range it.Topics {
		topicSet[strings.ToLower(t)] = true
	}
	neighborTopics := map[string]bool{}
	for _, e := range append(a.EdgesIn[it.ID], a.EdgesOut[it.ID]...) {
		other := e.SourceID
		if other == it.ID {
			other = e.TargetID
		}
		for _, t := range a.Items[other].Topics {
			neighborTopics[strings.ToLower(t)] = true
		}
	}
	return clamp01(float64(len(topicSet)+len(neighborTopics)) / 6 * clamp01(float64(degree(it.ID, a))/3))
}

func scoreItem(it Item, a Analysis, goal Goal) Score {
	gf := goalFit(it, goal)
	cent := a.Centrality[it.ID]
	ready := readiness(it, a, goal)
	fresh := freshness(it, a.NowYear)
	red := redundancyPenalty(it.ID, a)
	fit := energyTimeFit(it, goal)
	total := 100 * clamp01(gf*.30+cent*.20+ready*.18+fresh*.12+fit*.20-red*.22)
	r := []string{
		percentPhrase("goal fit", gf), percentPhrase("network centrality", cent), percentPhrase("readiness", ready), percentPhrase("energy/time fit", fit),
	}
	if red > 0 {
		r = append(r, percentPhrase("redundancy penalty", red))
	}
	return Score{ItemID: it.ID, Total: round1(total), GoalFit: round2(gf), Centrality: round2(cent), Readiness: round2(ready), Freshness: round2(fresh), RedundancyPenalty: round2(red), EnergyTimeFit: round2(fit), Role: a.Roles[it.ID], Rationale: r}
}

func goalFit(it Item, goal Goal) float64 {
	if it.GoalFit > 0 {
		return clamp01(it.GoalFit)
	}
	text := strings.ToLower(it.Title + " " + it.Notes + " " + strings.Join(it.Topics, " "))
	matches := 0
	for _, t := range append(goal.Topics, strings.Fields(goal.Prompt)...) {
		t = strings.ToLower(strings.Trim(t, " ,.;:!?()[]{}\"'"))
		if len(t) > 3 && strings.Contains(text, t) {
			matches++
		}
	}
	if matches == 0 {
		return .10
	}
	return clamp01(.35 + float64(matches)*.18)
}

func readiness(it Item, a Analysis, goal Goal) float64 {
	fam := goal.Familiarity
	if fam == 0 {
		fam = 3
	}
	diffFit := 1 - math.Abs(float64(it.Difficulty-fam))/5
	missing := 0
	total := 0
	// Edge convention: source prerequisite target. If a prerequisite source is unread, target is less ready.
	for _, e := range a.EdgesIn[it.ID] {
		if e.Type == "prerequisite" {
			total++
			if strings.ToLower(a.Items[e.SourceID].Status) != "read" {
				missing++
			}
		}
	}
	pre := 1.0
	if total > 0 {
		pre = 1 - float64(missing)/float64(total)
	}
	return clamp01(diffFit*.6 + pre*.4)
}

func freshness(it Item, now int) float64 {
	if it.Year == 0 {
		return .65
	}
	age := now - it.Year
	if age <= 2 {
		return .9
	}
	if age <= 7 {
		return .8
	}
	if strings.Contains(strings.ToLower(it.Type+" "+it.Notes), "classic") {
		return .82
	}
	return math.Max(.25, 1-float64(age)/24)
}

func redundancyPenalty(id string, a Analysis) float64 {
	if a.Roles[id] == RoleDuplicate {
		return .75
	}
	for _, c := range a.Duplicates {
		for _, x := range c {
			if x == id {
				return .55
			}
		}
	}
	return 0
}

func energyTimeFit(it Item, goal Goal) float64 {
	budget := goal.TimeMinutes
	if budget <= 0 {
		budget = 90
	}
	timeFit := 1.0
	if it.LengthMinutes > budget {
		timeFit = math.Max(.2, float64(budget)/float64(it.LengthMinutes))
	}
	energyFit := .85
	ge := strings.ToLower(goal.Energy)
	ie := strings.ToLower(it.Energy)
	if ge == "" || ie == "" {
		energyFit = .8
	} else if ge == ie {
		energyFit = 1
	} else if ge == "low" && (ie == "high" || it.Difficulty >= 4) {
		energyFit = .35
	} else {
		energyFit = .7
	}
	return clamp01(timeFit*.55 + energyFit*.45)
}

func PlanQueue(a Analysis, budget, limit int) QueuePlan {
	if budget <= 0 {
		budget = 180
	}
	if limit <= 0 {
		limit = 5
	}
	var candidates []QueueEntry
	for _, it := range a.Backlog.Items {
		if strings.ToLower(it.Status) == "read" {
			continue
		}
		s := a.Scores[it.ID]
		candidates = append(candidates, QueueEntry{Item: it, Score: s, Rationale: ExplainScore(it, s)})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score.Total == candidates[j].Score.Total {
			return candidates[i].Item.ID < candidates[j].Item.ID
		}
		return candidates[i].Score.Total > candidates[j].Score.Total
	})
	var plan QueuePlan
	plan.BudgetMinutes = budget
	used := 0
	for _, c := range candidates {
		if len(plan.Entries) >= limit {
			plan.Skipped = append(plan.Skipped, c)
			continue
		}
		// Always include at least one item regardless of budget, so the caller
		// always gets an actionable recommendation.
		if used+c.Item.LengthMinutes <= budget || len(plan.Entries) == 0 {
			plan.Entries = append(plan.Entries, c)
			used += c.Item.LengthMinutes
		} else {
			plan.Skipped = append(plan.Skipped, c)
		}
	}
	return plan
}

func Hygiene(a Analysis) HygieneReport {
	h := HygieneReport{DuplicateClusters: a.Duplicates}
	for _, it := range a.Backlog.Items {
		if degree(it.ID, a) == 0 {
			h.OrphanItems = append(h.OrphanItems, it.ID)
		}
		if isStaleHype(it, a.NowYear) || a.Roles[it.ID] == RoleStaleHype {
			h.StaleItems = append(h.StaleItems, it.ID)
		}
	}
	// Missing prerequisite bridges: unread item has prerequisites, but no overview/bridge neighbor to ease transition.
	for _, it := range a.Backlog.Items {
		hasPrereq := false
		hasBridge := false
		for _, e := range append(a.EdgesIn[it.ID], a.EdgesOut[it.ID]...) {
			if e.Type == "prerequisite" {
				hasPrereq = true
			}
			other := e.SourceID
			if other == it.ID {
				other = e.TargetID
			}
			if a.Roles[other] == RoleOverview || a.Roles[other] == RoleBridge {
				hasBridge = true
			}
		}
		if hasPrereq && !hasBridge {
			h.MissingBridges = append(h.MissingBridges, it.ID)
		}
	}
	sort.Strings(h.OrphanItems)
	sort.Strings(h.StaleItems)
	sort.Strings(h.MissingBridges)
	return h
}

func DetectDuplicates(b Backlog) [][]string {
	parent := map[string]string{}
	for _, it := range b.Items {
		parent[it.ID] = it.ID
	}
	var find func(string) string
	find = func(x string) string {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b string) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[rb] = ra
		}
	}
	for _, e := range b.Edges {
		if normEdgeType(e.Type) == "duplicate" {
			if _, ok := parent[e.SourceID]; ok {
				if _, ok2 := parent[e.TargetID]; ok2 {
					union(e.SourceID, e.TargetID)
				}
			}
		}
	}
	for i := range b.Items {
		for j := i + 1; j < len(b.Items); j++ {
			if titleSimilarity(b.Items[i].Title, b.Items[j].Title) >= .88 {
				union(b.Items[i].ID, b.Items[j].ID)
			}
		}
	}
	groups := map[string][]string{}
	for id := range parent {
		groups[find(id)] = append(groups[find(id)], id)
	}
	var out [][]string
	for _, g := range groups {
		if len(g) > 1 {
			sort.Strings(g)
			out = append(out, g)
		}
	}
	sort.Slice(out, func(i, j int) bool { return strings.Join(out[i], ",") < strings.Join(out[j], ",") })
	return out
}

func normalizeTitle(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	words := strings.Fields(b.String())
	stop := map[string]bool{"the": true, "a": true, "an": true, "and": true, "of": true, "to": true, "for": true, "in": true, "on": true}
	var keep []string
	for _, w := range words {
		if !stop[w] {
			keep = append(keep, w)
		}
	}
	return strings.Join(keep, " ")
}

func titleSimilarity(a, b string) float64 {
	wa, wb := strings.Fields(normalizeTitle(a)), strings.Fields(normalizeTitle(b))
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	set := map[string]int{}
	for _, w := range wa {
		set[w] |= 1
	}
	for _, w := range wb {
		set[w] |= 2
	}
	inter, union := 0, 0
	for _, mask := range set {
		union++
		if mask == 3 {
			inter++
		}
	}
	return float64(inter) / float64(union)
}

func inDupCluster(id string, clusters [][]string) bool {
	for _, c := range clusters {
		for _, x := range c {
			if x == id {
				return true
			}
		}
	}
	return false
}
func degree(id string, a Analysis) int { return len(a.EdgesIn[id]) + len(a.EdgesOut[id]) }
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
func round1(v float64) float64                    { return math.Round(v*10) / 10 }
func round2(v float64) float64                    { return math.Round(v*100) / 100 }
func percentPhrase(name string, v float64) string { return name + " " + strconvPercent(v) }
func strconvPercent(v float64) string             { return strconvI(int(math.Round(clamp01(v)*100))) + "%" }
func strconvI(i int) string                       { return strconv.FormatInt(int64(i), 10) }
