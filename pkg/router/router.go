package router

type DefaultRouter struct{}

func NewRouter() DefaultRouter { return DefaultRouter{} }

func (DefaultRouter) Route(query string, metadata GraphMetadata) Recommendation {
	tokens := tokenize(query)
	text := normalizedText(query)
	if (metadata.HasPermissionNode || metadata.HasDecisionBasis) && containsAny(tokens, actionVerbs) {
		return recAgentic("permission or decision-basis metadata with action verb")
	}
	if containsAny(tokens, temporalCues) || hasYearToken(tokens) || containsPhrase(text, "as of") {
		return Recommendation{QueryType: "Temporal", RetrievalMode: ModeHybridBM25Dense, ContextBudgetHint: 4500, Reasons: []string{"temporal cue"}}
	}
	if containsPhraseAny(text, proceduralPhrases) || containsAny(tokens, proceduralCues) {
		return Recommendation{QueryType: "Procedural", RetrievalMode: ModeLocalNeighborhood, ContextBudgetHint: 5000, Reasons: []string{"procedural cue"}}
	}
	if containsAny(tokens, comparisonCues) || containsPhrase(text, "which is better") {
		return Recommendation{QueryType: "Comparative", RetrievalMode: ModeDRIFTChain, ContextBudgetHint: 5500, Reasons: []string{"comparison cue"}}
	}
	if containsPhraseAny(text, adversarialPhrases) || containsAny(tokens, adversarialCues) {
		return Recommendation{QueryType: "Adversarial", RetrievalMode: ModeDRIFTChain, ContextBudgetHint: 6000, Reasons: []string{"adversarial or safety cue"}, CounterpointRequired: true}
	}
	if entity := exactEntity(tokens, metadata.EntityIDs); entity != "" {
		return Recommendation{QueryType: "Factual", RetrievalMode: ModeLocalNeighborhood, ContextBudgetHint: 3500, Reasons: []string{"exact entity match: " + entity}}
	}
	if metadata.MaxTopicSpan > 3 || containsPhraseAny(text, broadPhrases) {
		return Recommendation{QueryType: "Exploratory", RetrievalMode: ModeGlobalGraph, ContextBudgetHint: 6500, Reasons: []string{"broad sensemaking or topic span"}}
	}
	if len(tokens) <= 8 {
		return Recommendation{QueryType: "Factual", RetrievalMode: ModeHybridBM25Dense, ContextBudgetHint: 3000, Reasons: []string{"short query without entity match"}}
	}
	return Recommendation{QueryType: "Factual", RetrievalMode: ModeHybridBM25Dense, ContextBudgetHint: 4000, Reasons: []string{"fallback factual hybrid retrieval"}}
}

func recAgentic(reason string) Recommendation {
	return Recommendation{QueryType: "Agentic", RetrievalMode: ModeDRIFTChain, ContextBudgetHint: 6000, Reasons: []string{reason}}
}
