package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/mojomast/citewiseussy/pkg/rag"
)

type stdioRequest struct {
	Operation string          `json:"operation"`
	Request   json.RawMessage `json:"request"`
}

type stdioResponse struct {
	OK       bool            `json:"ok"`
	Response json.RawMessage `json:"response,omitempty"`
	Error    string          `json:"error,omitempty"`
}

func runStdio(in io.Reader, out io.Writer, p rag.Pipeline) error {
	var req stdioRequest
	dec := json.NewDecoder(in)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return err
	}
	value, err := dispatchStdio(req, p)
	resp := stdioResponse{OK: err == nil}
	if err != nil {
		resp.Error = err.Error()
	} else {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		resp.Response = data
	}
	return json.NewEncoder(out).Encode(resp)
}

func dispatchStdio(req stdioRequest, p rag.Pipeline) (any, error) {
	switch req.Operation {
	case "router":
		var r routeRequest
		if err := json.Unmarshal(req.Request, &r); err != nil {
			return nil, err
		}
		return rag.DefaultRouter().Route(r.Query, r.Metadata), nil
	case "rank":
		var r candidateRequest
		if err := json.Unmarshal(req.Request, &r); err != nil {
			return nil, err
		}
		analysis, err := rag.Analyze(r.candidateSet())
		if err != nil {
			return nil, err
		}
		ranked, err := rag.DefaultRanker().Rank(r.Access, analysis, r.TokenBudget)
		if err != nil {
			return nil, err
		}
		return rankResponse{Ranked: ranked}, nil
	case "pack":
		var r candidateRequest
		if err := json.Unmarshal(req.Request, &r); err != nil {
			return nil, err
		}
		resp, err := p.Run(r.ragRequest())
		if err != nil && !errors.Is(err, rag.ErrRedCorrectiveSignal) {
			return nil, err
		}
		return packResponse{Plan: resp.Plan}, nil
	case "hygiene":
		var r candidateRequest
		if err := json.Unmarshal(req.Request, &r); err != nil {
			return nil, err
		}
		analysis, err := rag.Analyze(r.candidateSet())
		if err != nil {
			return nil, err
		}
		return hygieneResponse{Hygiene: rag.DefaultHygieneAnalyzer().Analyze(analysis, r.AllowDegradedPlan)}, nil
	default:
		return nil, fmt.Errorf("unknown operation %q", req.Operation)
	}
}
