package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"Assignment_2/cmd/internal/store"
)

const maxTitleLen = 200 // optional “limit title length” extra :contentReference[oaicite:5]{index=5}

type TaskHandler struct {
	st *store.Store
}

func NewTaskHandler(st *store.Store) *TaskHandler {
	return &TaskHandler{st: st}
}

func (h *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// JSON only + correct status codes :contentReference[oaicite:6]{index=6}
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodPatch:
		h.handlePatch(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (h *TaskHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// If id is provided => return single task :contentReference[oaicite:8]{index=8}
	if idStr := q.Get("id"); idStr != "" {
		id, ok := parsePositiveInt(idStr)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		task, exists := h.st.Get(id)
		if !exists {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		writeJSON(w, http.StatusOK, task)
		return
	}

	// Optional filtering: GET /tasks?done=true :contentReference[oaicite:9]{index=9}
	var doneFilter *bool
	if doneStr := q.Get("done"); doneStr != "" {
		v, ok := parseBoolStrict(doneStr)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid done"})
			return
		}
		doneFilter = &v
	}

	tasks := h.st.List(doneFilter)
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	// Create task
	var req struct {
		Title string `json:"title"`
	}
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" || len(title) > maxTitleLen {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid title"})
		return
	}

	task := h.st.Create(title)
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	// PATCH /tasks?id=1, id required, done must be boolean
	idStr := r.URL.Query().Get("id")
	id, ok := parsePositiveInt(idStr)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req struct {
		Done *bool `json:"done"`
	}
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Done == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid done"})
		return
	}

	updated := h.st.UpdateDone(id, *req.Done)
	if !updated {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"updated": true})
}

func (h *TaskHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, ok := parsePositiveInt(idStr)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if !h.st.Delete(id) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

func parsePositiveInt(s string) (int, bool) {
	if strings.TrimSpace(s) == "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func parseBoolStrict(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}
