package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mojomast/citewiseussy/pkg/packer"
	"github.com/mojomast/citewiseussy/pkg/rag"
)

type server struct {
	pipeline rag.Pipeline
}

func newServer(p rag.Pipeline) http.Handler {
	s := server{pipeline: p}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /router", s.route)
	mux.HandleFunc("POST /rank", s.rank)
	mux.HandleFunc("POST /pack", s.pack)
	mux.HandleFunc("POST /hygiene", s.hygiene)
	return mux
}

func (s server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok", GraphHygieneSignal: packer.HygieneGreen})
}

func (s server) route(w http.ResponseWriter, r *http.Request) {
	var req routeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	writeJSON(w, http.StatusOK, rag.DefaultRouter().Route(req.Query, req.Metadata))
}

func (s server) rank(w http.ResponseWriter, r *http.Request) {
	var req candidateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	analysis, err := rag.Analyze(req.candidateSet())
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	ranked, err := rag.DefaultRanker().Rank(req.Access, analysis, req.TokenBudget)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rankResponse{Ranked: ranked})
}

func (s server) pack(w http.ResponseWriter, r *http.Request) {
	var req candidateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.pipeline.Run(req.ragRequest())
	if err != nil && !errors.Is(err, rag.ErrRedCorrectiveSignal) {
		status := http.StatusInternalServerError
		if errors.Is(err, rag.ErrInvalidCandidates) {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, packResponse{Plan: resp.Plan})
}

func (s server) hygiene(w http.ResponseWriter, r *http.Request) {
	var req candidateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	analysis, err := rag.Analyze(req.candidateSet())
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, hygieneResponse{Hygiene: rag.DefaultHygieneAnalyzer().Analyze(analysis, req.AllowDegradedPlan)})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
