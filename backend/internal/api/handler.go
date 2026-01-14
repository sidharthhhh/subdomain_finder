package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"subdomain-finder/internal/orchestrator"
)

type ScanRequest struct {
	Domain string `json:"domain"`
}

type ScanResponse struct {
	Subdomains []string `json:"subdomains"`
	Count      int      `json:"count"`
}

type Handler struct {
	orchestrator *orchestrator.Orchestrator
}

func NewHandler(orch *orchestrator.Orchestrator) *Handler {
	return &Handler{
		orchestrator: orch,
	}
}

func (h *Handler) HandleScan(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Domain == "" {
		http.Error(w, "Domain is required", http.StatusBadRequest)
		return
	}

	// Create a context with timeout
	// For production, maybe 5-10 minutes depending on scanning depth.
	// For MVP demo, 60 seconds is reasonable.
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	results, err := h.orchestrator.Run(ctx, req.Domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transform to simple strings for the response as per spec request
	// Request: { "domain": ... } -> Response: { "subdomains": [...] }
	subs := make([]string, len(results))
	for i, r := range results {
		subs[i] = r.FullDomain
	}

	resp := ScanResponse{
		Subdomains: subs,
		Count:      len(subs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Needed imports workaround because I typed context.WithTimeout but missed importing "context"
// Fixing imports below in the same file write.
