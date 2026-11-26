// Package models holds the data types shared by the store and the HTTP layer.
package models

import "time"

// Movement types.
const (
	MovementIn     = "in"     // stock received
	MovementOut    = "out"    // stock shipped / consumed
	MovementAdjust = "adjust" // set quantity to an absolute value (stock count correction)
)

// Item is a product stored in the warehouse.
type Item struct {
	ID           int       `json:"id"`
	SKU          string    `json:"sku"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Location     string    `json:"location"` // e.g. "A-03-2" (aisle-rack-shelf)
	Quantity     int       `json:"quantity"`
	ReorderLevel int       `json:"reorder_level"` // low-stock when Quantity <= ReorderLevel (0 disables)
	LowStock     bool      `json:"low_stock"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ItemInput is the request body for creating an item.
type ItemInput struct {
	SKU          string `json:"sku"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	Quantity     int    `json:"quantity"`
	ReorderLevel int    `json:"reorder_level"`
}

// ItemUpdate is the request body for editing an item's details.
// Quantity is deliberately absent: it only changes through movements,
// so every change is traceable.
type ItemUpdate struct {
	SKU          string `json:"sku"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	ReorderLevel int    `json:"reorder_level"`
}

// Movement is one entry in the stock ledger.
type Movement struct {
	ID        int       `json:"id"`
	ItemID    int       `json:"item_id"`
	Type      string    `json:"type"`
	Delta     int       `json:"delta"`   // signed change applied to quantity
	Balance   int       `json:"balance"` // quantity after this movement
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// MovementInput is the request body for recording a movement.
// For "in"/"out" Quantity is the amount moved (> 0).
// For "adjust" Quantity is the new absolute stock level (>= 0).
type MovementInput struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
	Note     string `json:"note"`
}

// Summary is a dashboard-style overview.
type Summary struct {
	TotalItems     int `json:"total_items"`
	TotalUnits     int `json:"total_units"`
	LowStockItems  int `json:"low_stock_items"`
	TotalMovements int `json:"total_movements"`
}