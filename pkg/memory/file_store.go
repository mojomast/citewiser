package memory

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mojomast/citewiseussy/pkg/access"
	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/ragnode"
)

const DefaultPath = "citewiserag_memory.jsonl"

var ErrRedPlan = errors.New("memory: red context plans are not eligible for write-back")

// FileStore stores one accepted context plan per JSONL line.
type FileStore struct {
	Path         string
	Caller       access.Context
	Access       access.Controller
	CurrentNodes map[string]ragnode.RAGNode
	Now          func() time.Time

	mu sync.Mutex
}

// StoreContextPlan appends a non-red plan to JSONL after populating its
// write-back payload.
func (s *FileStore) StoreContextPlan(queryID string, plan packer.ContextPlan) error {
	if plan.HygieneSignal == packer.HygieneRed {
		return ErrRedPlan
	}
	if queryID == "" {
		queryID = plan.QueryID
	}
	if plan.QueryID == "" {
		plan.QueryID = queryID
	}
	createdAt := s.now().UTC().Format(time.RFC3339)
	topics := planTopics(plan)
	versions := NodeVersions(plan)
	payload := WriteBackPayload{
		PlanHash:      PlanHash(plan.QueryType, plan.EvidencePath, versions),
		QueryID:       queryID,
		QueryType:     plan.QueryType,
		Topics:        topics,
		EvidencePath:  append([]string(nil), plan.EvidencePath...),
		NodeVersions:  versions,
		HygieneSignal: plan.HygieneSignal,
		CreatedAt:     createdAt,
	}
	plan.WriteBackPayload = payload
	record := StoredPlan{QueryID: queryID, Topics: topics, Plan: plan, CreatedAt: createdAt, PlanHash: payload.PlanHash}
	line, err := json.Marshal(record)
	if err != nil {
		return err
	}
	line = append(line, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.OpenFile(s.path(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(line); err != nil {
		return err
	}
	return file.Sync()
}

// LoadPriorPlan returns the newest stored plan for queryID after access re-gating.
func (s *FileStore) LoadPriorPlan(queryID string) (packer.ContextPlan, bool, error) {
	records, err := s.records()
	if err != nil {
		return packer.ContextPlan{}, false, err
	}
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].QueryID == queryID {
			return s.regate(records[i].Plan), true, nil
		}
	}
	return packer.ContextPlan{}, false, nil
}

// SimilarPriorPlans returns reusable plans sorted by topic similarity then query ID.
func (s *FileStore) SimilarPriorPlans(topics []string, limit int) ([]packer.ContextPlan, error) {
	records, err := s.records()
	if err != nil {
		return nil, err
	}
	type candidate struct {
		plan  packer.ContextPlan
		score float64
	}
	candidates := []candidate{}
	for _, record := range records {
		score := TopicJaccard(topics, record.Topics)
		if score < 0.70 {
			continue
		}
		plan := s.regate(record.Plan)
		if rejection, ok := s.ReuseRejection(plan, topics, plan.QueryType); ok {
			plan.Suppressed = append(plan.Suppressed, packer.SuppressedEntry{NodeID: plan.QueryID, Reason: rejection.Reason, Detail: rejection.Detail})
			continue
		}
		candidates = append(candidates, candidate{plan: plan, score: score})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].plan.QueryID < candidates[j].plan.QueryID
	})
	if limit <= 0 || limit > len(candidates) {
		limit = len(candidates)
	}
	plans := make([]packer.ContextPlan, 0, limit)
	for i := 0; i < limit; i++ {
		plans = append(plans, candidates[i].plan)
	}
	return plans, nil
}

// ReuseRejection reports the first policy reason that prevents plan reuse.
func (s *FileStore) ReuseRejection(plan packer.ContextPlan, topics []string, queryType packer.QueryType) (ReuseRejection, bool) {
	if !compatibleQueryType(plan.QueryType, queryType) {
		return ReuseRejection{QueryID: plan.QueryID, Reason: RejectQueryType, Detail: fmt.Sprintf("prior %s requested %s", plan.QueryType, queryType)}, true
	}
	if TopicJaccard(topics, planTopics(plan)) < 0.70 {
		return ReuseRejection{QueryID: plan.QueryID, Reason: RejectTopicMismatch, Detail: "topic Jaccard below 0.70"}, true
	}
	storedVersions := NodeVersions(plan)
	for id, stored := range storedVersions {
		if current, ok := s.CurrentNodes[id]; ok {
			if current.Version != stored {
				return ReuseRejection{QueryID: plan.QueryID, Reason: RejectVersionMismatch, Detail: id}, true
			}
			if current.SupersededBy != "" {
				return ReuseRejection{QueryID: plan.QueryID, Reason: RejectSuperseded, Detail: id}, true
			}
		}
	}
	if len(s.regate(plan).Slots) != len(plan.Slots) {
		return ReuseRejection{QueryID: plan.QueryID, Reason: RejectAccessControl, Detail: "caller cannot access every prior slot"}, true
	}
	return ReuseRejection{}, false
}

func (s *FileStore) records() ([]StoredPlan, error) {
	file, err := os.Open(s.path())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	records := []StoredPlan{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record StoredPlan
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, scanner.Err()
}

func (s *FileStore) path() string {
	if s.Path == "" {
		return DefaultPath
	}
	return s.Path
}

func (s *FileStore) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// PlanHash computes the stable memory hash from query type, evidence path, and versions.
func PlanHash(queryType packer.QueryType, evidencePath []string, nodeVersions map[string]string) string {
	keys := make([]string, 0, len(nodeVersions))
	for key := range nodeVersions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := []string{string(queryType), strings.Join(evidencePath, ">")}
	for _, key := range keys {
		parts = append(parts, key+"="+nodeVersions[key])
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// NodeVersions extracts slot source versions keyed by node ID.
func NodeVersions(plan packer.ContextPlan) map[string]string {
	versions := map[string]string{}
	for _, slot := range plan.Slots {
		versions[slot.NodeID] = slot.Source.Version
	}
	return versions
}

// TopicJaccard returns deterministic case-insensitive Jaccard similarity.
func TopicJaccard(a, b []string) float64 {
	left := topicSet(a)
	right := topicSet(b)
	if len(left) == 0 && len(right) == 0 {
		return 1
	}
	intersection := 0
	for topic := range left {
		if right[topic] {
			intersection++
		}
	}
	union := len(left) + len(right) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func topicSet(topics []string) map[string]bool {
	set := map[string]bool{}
	for _, topic := range topics {
		topic = strings.ToLower(strings.TrimSpace(topic))
		if topic != "" {
			set[topic] = true
		}
	}
	return set
}

func planTopics(plan packer.ContextPlan) []string {
	set := map[string]bool{}
	for _, slot := range plan.Slots {
		if slot.Source.CommunityID != "" {
			set[strings.ToLower(slot.Source.CommunityID)] = true
		}
	}
	topics := make([]string, 0, len(set))
	for topic := range set {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	return topics
}

func compatibleQueryType(prior, requested packer.QueryType) bool {
	return prior == requested || (prior == packer.QueryFactual && requested == packer.QueryAgentic)
}
