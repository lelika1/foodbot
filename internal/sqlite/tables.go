package sqlite

import "time"

// User represents a row in the USER table.
type User struct {
	ID    int
	Name  string
	Limit uint32
}

// Product represents a row in the Product table.
type Product struct {
	Name string
	Kcal uint32
}

// Report represents a row in the Reports table.
type Report struct {
	When    time.Time
	Product string
	Kcal    uint32
	Grams   uint32
}
