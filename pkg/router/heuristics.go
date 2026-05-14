package router

import (
	"strconv"
	"strings"
	"unicode"
)

var actionVerbs = map[string]bool{"create": true, "issue": true, "approve": true, "delete": true, "send": true, "change": true, "execute": true, "deploy": true, "grant": true, "revoke": true, "purchase": true, "refund": true, "file": true}
var temporalCues = map[string]bool{"current": true, "latest": true, "today": true, "changed": true, "new": true, "old": true, "superseded": true, "version": true}
var proceduralCues = map[string]bool{"steps": true, "runbook": true, "procedure": true, "workflow": true, "sop": true, "troubleshoot": true}
var comparisonCues = map[string]bool{"compare": true, "versus": true, "vs": true, "tradeoff": true, "difference": true}
var adversarialCues = map[string]bool{"risk": true, "allowed": true, "not": true, "never": true, "forbidden": true, "exception": true, "counterexample": true}

var proceduralPhrases = []string{"how do i"}
var adversarialPhrases = []string{"should i", "is it safe", "can i"}
var broadPhrases = []string{"what are the themes", "summarize", "overview", "map", "explain the landscape"}

func normalizedText(query string) string {
	return strings.Join(tokenize(query), " ")
}

func tokenize(query string) []string {
	fields := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := []string{}
	for _, field := range fields {
		if field != "" {
			out = append(out, field)
		}
	}
	return out
}

func containsAny(tokens []string, values map[string]bool) bool {
	for _, token := range tokens {
		if values[token] {
			return true
		}
	}
	return false
}

func containsPhraseAny(text string, phrases []string) bool {
	for _, phrase := range phrases {
		if containsPhrase(text, phrase) {
			return true
		}
	}
	return false
}

func containsPhrase(text, phrase string) bool {
	return strings.Contains(text, strings.ToLower(phrase))
}

func hasYearToken(tokens []string) bool {
	for _, token := range tokens {
		if len(token) != 4 {
			continue
		}
		year, err := strconv.Atoi(token)
		if err == nil && year >= 1900 && year <= 2199 {
			return true
		}
	}
	return false
}

func exactEntity(tokens []string, entities []string) string {
	query := " " + strings.Join(tokens, " ") + " "
	for _, entity := range entities {
		entityText := normalizedText(entity)
		if entityText != "" && strings.Contains(query, " "+entityText+" ") {
			return entity
		}
	}
	return ""
}
