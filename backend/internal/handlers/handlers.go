// Package handlers contains the HTTP API.
package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"warehouse_tracker/internal/models"
	"warehouse_tracker/internal/store"
)

// API wires HTTP routes to the store.
type API struct {
	store *store.Store
}

func New(s *store.Store) *API { return &API{store: s} }

// Routes returns the fully wrapped HTTP handler.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", a.health)
	mux.HandleFunc("GET /api/summary", a.summary)

	mux.HandleFunc("GET /api/items", a.listItems)
	mux.HandleFunc("POST /api/items", a.createItem)
	mux.HandleFunc("GET /api/items/{id}", a.getItem)
	mux.HandleFunc("PUT /api/items/{id}", a.updateItem)
	mux.HandleFunc("DELETE /api/items/{id}", a.deleteItem)

	mux.HandleFunc("GET /api/items/{id}/movements", a.itemMovements)
	mux.HandleFunc("POST /api/items/{id}/movements", a.createMovement)
	mux.HandleFunc("GET /api/movements", a.listMovements)

	return chain(mux, recoverer, logger, cors)
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}

func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return 0, false
	}
	return id, true
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrDuplicateSKU), errors.Is(err, store.ErrInsufficientStock):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, store.ErrInvalid):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// ---------- handlers ----------

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) summary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.Summary())
}

func (a *API) listItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	low := r.URL.Query().Get("low_stock") == "true"
	writeJSON(w, http.StatusOK, a.store.ListItems(q, low))
}

func (a *API) createItem(w http.ResponseWriter, r *http.Request) {
	var in models.ItemInput
	if !decode(w, r, &in) {
		return
	}
	it, err := a.store.CreateItem(in)
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Location", "/api/items/"+strconv.Itoa(it.ID))
	writeJSON(w, http.StatusCreated, it)
}

func (a *API) getItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	it, err := a.store.GetItem(id)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (a *API) updateItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in models.ItemUpdate
	if !decode(w, r, &in) {
		return
	}
	it, err := a.store.UpdateItem(id, in)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (a *API) deleteItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := a.store.DeleteItem(id); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) itemMovements(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if _, err := a.store.GetItem(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a.store.ListMovements(id))
}

func (a *API) createMovement(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in models.MovementInput
	if !decode(w, r, &in) {
		return
	}
	m, it, err := a.store.RecordMovement(id, in)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"movement": m, "item": it})
}

func (a *API) listMovements(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.ListMovements(0))
}