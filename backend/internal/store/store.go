// Package store is the PostgreSQL data layer (pgx v5 connection pool).
// The schema is created automatically on startup (see schema.sql).
package store

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"warehouse_tracker/internal/models"
)

//go:embed schema.sql
var schema string

var (
	ErrNotFound          = errors.New("not found")
	ErrDuplicateSKU      = errors.New("sku already exists")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalid           = errors.New("invalid input")
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// Store wraps a pgx connection pool.
type Store struct {
	pool *pgxpool.Pool
}

// New connects to PostgreSQL, checks the connection and creates the tables.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Render's database can take a moment to accept connections after a cold start.
	var pingErr error
	for attempt := 1; attempt <= 5; attempt++ {
		pingErr = pool.Ping(ctx)
		if pingErr == nil {
			break
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt) * time.Second):
		}
	}
	if pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to database: %w", pingErr)
	}

	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close()                         { s.pool.Close() }
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

func normSKU(sku string) string { return strings.ToUpper(strings.TrimSpace(sku)) }

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

const itemCols = `id, sku, name, description, location, quantity, reorder_level, created_at, updated_at`

func scanItem(row pgx.Row) (models.Item, error) {
	var it models.Item
	err := row.Scan(&it.ID, &it.SKU, &it.Name, &it.Description, &it.Location,
		&it.Quantity, &it.ReorderLevel, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return models.Item{}, err
	}
	it.LowStock = it.ReorderLevel > 0 && it.Quantity <= it.ReorderLevel
	return it, nil
}

// ---------- Items ----------

// ListItems returns items ordered by ID. q filters on SKU/name/location
// (case-insensitive substring); lowOnly keeps only low-stock items.
func (s *Store) ListItems(ctx context.Context, q string, lowOnly bool) ([]models.Item, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+itemCols+` FROM items
		WHERE ($1::text = ''
		       OR strpos(lower(sku), lower($1::text)) > 0
		       OR strpos(lower(name), lower($1::text)) > 0
		       OR strpos(lower(location), lower($1::text)) > 0)
		  AND (NOT $2::boolean OR (reorder_level > 0 AND quantity <= reorder_level))
		ORDER BY id`, strings.TrimSpace(q), lowOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Item, 0)
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) GetItem(ctx context.Context, id int) (models.Item, error) {
	it, err := scanItem(s.pool.QueryRow(ctx, `SELECT `+itemCols+` FROM items WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Item{}, ErrNotFound
	}
	return it, err
}

func (s *Store) CreateItem(ctx context.Context, in models.ItemInput) (models.Item, error) {
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

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Item{}, err
	}
	defer tx.Rollback(ctx)

	it, err := scanItem(tx.QueryRow(ctx, `
		INSERT INTO items (sku, name, description, location, quantity, reorder_level)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+itemCols,
		sku, name, strings.TrimSpace(in.Description), strings.TrimSpace(in.Location),
		in.Quantity, in.ReorderLevel))
	if err != nil {
		if isUniqueViolation(err) {
			return models.Item{}, ErrDuplicateSKU
		}
		return models.Item{}, err
	}

	if it.Quantity > 0 {
		if _, err := tx.Exec(ctx, `
			INSERT INTO movements (item_id, type, delta, balance, note)
			VALUES ($1, 'in', $2, $2, 'initial stock')`, it.ID, it.Quantity); err != nil {
			return models.Item{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return models.Item{}, err
	}
	return it, nil
}

func (s *Store) UpdateItem(ctx context.Context, id int, in models.ItemUpdate) (models.Item, error) {
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

	it, err := scanItem(s.pool.QueryRow(ctx, `
		UPDATE items
		SET sku = $2, name = $3, description = $4, location = $5,
		    reorder_level = $6, updated_at = now()
		WHERE id = $1
		RETURNING `+itemCols,
		id, sku, name, strings.TrimSpace(in.Description), strings.TrimSpace(in.Location), in.ReorderLevel))
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Item{}, ErrNotFound
	case isUniqueViolation(err):
		return models.Item{}, ErrDuplicateSKU
	case err != nil:
		return models.Item{}, err
	}
	return it, nil
}

// DeleteItem removes an item; its movements are removed by ON DELETE CASCADE.
func (s *Store) DeleteItem(ctx context.Context, id int) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM items WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Movements ----------

// RecordMovement applies a stock movement inside a transaction. The item row is
// locked (FOR UPDATE) so two simultaneous shipments cannot oversell the stock.
func (s *Store) RecordMovement(ctx context.Context, itemID int, in models.MovementInput) (models.Movement, models.Item, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Movement{}, models.Item{}, err
	}
	defer tx.Rollback(ctx)

	var current int
	err = tx.QueryRow(ctx, `SELECT quantity FROM items WHERE id = $1 FOR UPDATE`, itemID).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Movement{}, models.Item{}, ErrNotFound
	}
	if err != nil {
		return models.Movement{}, models.Item{}, err
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
		if in.Quantity > current {
			return models.Movement{}, models.Item{}, ErrInsufficientStock
		}
		delta = -in.Quantity
	case models.MovementAdjust:
		if in.Quantity < 0 {
			return models.Movement{}, models.Item{}, invalid("quantity cannot be negative")
		}
		delta = in.Quantity - current
	default:
		return models.Movement{}, models.Item{}, invalid("type must be one of: in, out, adjust")
	}
	balance := current + delta

	it, err := scanItem(tx.QueryRow(ctx, `
		UPDATE items SET quantity = $2, updated_at = now()
		WHERE id = $1 RETURNING `+itemCols, itemID, balance))
	if err != nil {
		return models.Movement{}, models.Item{}, err
	}

	m := models.Movement{
		ItemID:  itemID,
		Type:    in.Type,
		Delta:   delta,
		Balance: balance,
		Note:    strings.TrimSpace(in.Note),
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO movements (item_id, type, delta, balance, note)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		m.ItemID, m.Type, m.Delta, m.Balance, m.Note).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return models.Movement{}, models.Item{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Movement{}, models.Item{}, err
	}
	return m, it, nil
}

// ListMovements returns the ledger, newest first (max 1000 rows). itemID 0 means all items.
func (s *Store) ListMovements(ctx context.Context, itemID int) ([]models.Movement, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, item_id, type, delta, balance, note, created_at
		FROM movements
		WHERE ($1::bigint = 0 OR item_id = $1::bigint)
		ORDER BY id DESC
		LIMIT 1000`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Movement, 0)
	for rows.Next() {
		var m models.Movement
		if err := rows.Scan(&m.ID, &m.ItemID, &m.Type, &m.Delta, &m.Balance, &m.Note, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ---------- Summary ----------

func (s *Store) Summary(ctx context.Context) (models.Summary, error) {
	var sum models.Summary
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(quantity), 0),
		       COUNT(*) FILTER (WHERE reorder_level > 0 AND quantity <= reorder_level)
		FROM items`).Scan(&sum.TotalItems, &sum.TotalUnits, &sum.LowStockItems)
	if err != nil {
		return sum, err
	}
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM movements`).Scan(&sum.TotalMovements)
	return sum, err
}
