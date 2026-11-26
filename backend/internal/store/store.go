// Package store is a small thread-safe data layer. Everything lives in memory
// and is persisted to a JSON file after each write, so there are no external
// dependencies. Swap this package for a SQL implementation later without
// touching the handlers (keep the method signatures).
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"warehouse_tracker/internal/models"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrDuplicateSKU      = errors.New("sku already exists")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalid           = errors.New("invalid input")
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

type fileData struct {
	Items          map[int]*models.Item `json:"items"`
	Movements      []models.Movement    `json:"movements"`
	NextItemID     int                  `json:"next_item_id"`
	NextMovementID int                  `json:"next_movement_id"`
}

// Store holds all warehouse data.
type Store struct {
	mu   sync.RWMutex
	path string // empty = no persistence (useful for tests)
	data fileData
}

// New opens (or creates) a store backed by the JSON file at path.
func New(path string) (*Store, error) {
	s := &Store{
		path: path,
		data: fileData{Items: map[int]*models.Item{}},
	}
	if path == "" {
		return s, nil
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &s.data); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
	}
	if s.data.Items == nil {
		s.data.Items = map[int]*models.Item{}
	}
	return s, nil
}

// saveLocked writes the data atomically. Caller must hold the write lock.
func (s *Store) saveLocked() error {
	if s.path == "" {
		return nil
	}
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func normSKU(sku string) string { return strings.ToUpper(strings.TrimSpace(sku)) }

func withFlags(it *models.Item) models.Item {
	c := *it
	c.LowStock = c.ReorderLevel > 0 && c.Quantity <= c.ReorderLevel
	return c
}

func (s *Store) skuTakenLocked(sku string, exceptID int) bool {
	for _, it := range s.data.Items {
		if it.ID != exceptID && it.SKU == sku {
			return true
		}
	}
	return false
}

// ---------- Items ----------

// ListItems returns items sorted by ID. q filters on SKU/name/location
// (case-insensitive substring); lowOnly keeps only low-stock items.
func (s *Store) ListItems(q string, lowOnly bool) []models.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q = strings.ToLower(strings.TrimSpace(q))
	out := make([]models.Item, 0, len(s.data.Items))
	for _, it := range s.data.Items {
		v := withFlags(it)
		if lowOnly && !v.LowStock {
			continue
		}
		if q != "" &&
			!strings.Contains(strings.ToLower(v.SKU), q) &&
			!strings.Contains(strings.ToLower(v.Name), q) &&
			!strings.Contains(strings.ToLower(v.Location), q) {
			continue
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) GetItem(id int) (models.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, ok := s.data.Items[id]
	if !ok {
		return models.Item{}, ErrNotFound
	}
	return withFlags(it), nil
}

func (s *Store) CreateItem(in models.ItemInput) (models.Item, error) {
	sku := normSKU(in.SKU)
	name := strings.TrimSpace(in.Name)
	switch {
	case sku == "":
		return models.Item{}, invalid("sku is required")
	case name == "":
		return models.Item{}, invalid("name is required")
	case in.Quantity < 0:
		return models.Item{}, invalid("quantity cannot be negative")
	case in.ReorderLevel < 0:
		return models.Item{}, invalid("reorder_level cannot be negative")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.skuTakenLocked(sku, 0) {
		return models.Item{}, ErrDuplicateSKU
	}
	now := time.Now().UTC()
	s.data.NextItemID++
	it := &models.Item{
		ID:           s.data.NextItemID,
		SKU:          sku,
		Name:         name,
		Description:  strings.TrimSpace(in.Description),
		Location:     strings.TrimSpace(in.Location),
		Quantity:     in.Quantity,
		ReorderLevel: in.ReorderLevel,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.data.Items[it.ID] = it
	if it.Quantity > 0 {
		s.addMovementLocked(it, models.MovementIn, it.Quantity, "initial stock", now)
	}
	if err := s.saveLocked(); err != nil {
		return models.Item{}, err
	}
	return withFlags(it), nil
}

func (s *Store) UpdateItem(id int, in models.ItemUpdate) (models.Item, error) {
	sku := normSKU(in.SKU)
	name := strings.TrimSpace(in.Name)
	switch {
	case sku == "":
		return models.Item{}, invalid("sku is required")
	case name == "":
		return models.Item{}, invalid("name is required")
	case in.ReorderLevel < 0:
		return models.Item{}, invalid("reorder_level cannot be negative")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	it, ok := s.data.Items[id]
	if !ok {
		return models.Item{}, ErrNotFound
	}
	if s.skuTakenLocked(sku, id) {
		return models.Item{}, ErrDuplicateSKU
	}
	it.SKU = sku
	it.Name = name
	it.Description = strings.TrimSpace(in.Description)
	it.Location = strings.TrimSpace(in.Location)
	it.ReorderLevel = in.ReorderLevel
	it.UpdatedAt = time.Now().UTC()
	if err := s.saveLocked(); err != nil {
		return models.Item{}, err
	}
	return withFlags(it), nil
}

// DeleteItem removes an item and its movement history.
func (s *Store) DeleteItem(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data.Items[id]; !ok {
		return ErrNotFound
	}
	delete(s.data.Items, id)
	kept := make([]models.Movement, 0, len(s.data.Movements))
	for _, m := range s.data.Movements {
		if m.ItemID != id {
			kept = append(kept, m)
		}
	}
	s.data.Movements = kept
	return s.saveLocked()
}

// ---------- Movements ----------

func (s *Store) addMovementLocked(it *models.Item, typ string, delta int, note string, at time.Time) models.Movement {
	s.data.NextMovementID++
	m := models.Movement{
		ID:        s.data.NextMovementID,
		ItemID:    it.ID,
		Type:      typ,
		Delta:     delta,
		Balance:   it.Quantity,
		Note:      strings.TrimSpace(note),
		CreatedAt: at,
	}
	s.data.Movements = append(s.data.Movements, m)
	return m
}

// RecordMovement applies a stock movement and returns it with the updated item.
func (s *Store) RecordMovement(itemID int, in models.MovementInput) (models.Movement, models.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	it, ok := s.data.Items[itemID]
	if !ok {
		return models.Movement{}, models.Item{}, ErrNotFound
	}

	var delta int
	switch in.Type {
	case models.MovementIn:
		if in.Quantity <= 0 {
			return models.Movement{}, models.Item{}, invalid("quantity must be greater than 0")
		}
		delta = in.Quantity
	case models.MovementOut:
		if in.Quantity <= 0 {
			return models.Movement{}, models.Item{}, invalid("quantity must be greater than 0")
		}
		if in.Quantity > it.Quantity {
			return models.Movement{}, models.Item{}, ErrInsufficientStock
		}
		delta = -in.Quantity
	case models.MovementAdjust:
		if in.Quantity < 0 {
			return models.Movement{}, models.Item{}, invalid("quantity cannot be negative")
		}
		delta = in.Quantity - it.Quantity
	default:
		return models.Movement{}, models.Item{}, invalid("type must be one of: in, out, adjust")
	}

	now := time.Now().UTC()
	it.Quantity += delta
	it.UpdatedAt = now
	m := s.addMovementLocked(it, in.Type, delta, in.Note, now)
	if err := s.saveLocked(); err != nil {
		return models.Movement{}, models.Item{}, err
	}
	return m, withFlags(it), nil
}

// ListMovements returns the ledger, newest first. itemID 0 means all items.
func (s *Store) ListMovements(itemID int) []models.Movement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]models.Movement, 0)
	for i := len(s.data.Movements) - 1; i >= 0; i-- {
		m := s.data.Movements[i]
		if itemID == 0 || m.ItemID == itemID {
			out = append(out, m)
		}
	}
	return out
}

// ---------- Summary ----------

func (s *Store) Summary() models.Summary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sum := models.Summary{
		TotalItems:     len(s.data.Items),
		TotalMovements: len(s.data.Movements),
	}
	for _, it := range s.data.Items {
		sum.TotalUnits += it.Quantity
		if withFlags(it).LowStock {
			sum.LowStockItems++
		}
	}
	return sum
}
