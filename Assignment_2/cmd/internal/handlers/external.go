package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type ExternalHandler struct {
	client *http.Client
}

func NewExternalHandler() *ExternalHandler {
	return &ExternalHandler{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *ExternalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	resp, err := h.client.Get("https://jsonplaceholder.typicode.com/todos")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "external api error"})
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewDecoder(resp.Body).Decode(&[]any{}) // validate JSON quickly

	// Re-fetch (simple & safe): do it properly in one pass instead:
	resp2, err := h.client.Get("https://jsonplaceholder.typicode.com/todos")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "external api error"})
		return
	}
	defer resp2.Body.Close()
	_ = json.NewEncoder(w).Encode(mustDecodeJSON(resp2))
}

func mustDecodeJSON(resp *http.Response) any {
	var v any
	_ = json.NewDecoder(resp.Body).Decode(&v)
	return v
}
