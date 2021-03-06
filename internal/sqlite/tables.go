package sqlite

import (
	"fmt"
	"time"
)

// User represents a row in the USER table.
type User struct {
	ID       int
	Name     string
	Limit    uint32
	Location *time.Location
}

// Product represents a row in the Product table.
type Product struct {
	Name string `json:"Name"`
	Kcal uint32 `json:"Kcal"`
}

func (p Product) String() string {
	return fmt.Sprintf("%v %v kcal", p.Name, p.Kcal)
}

// Report represents a row in the Reports table.
type Report struct {
	Product
	Grams uint32
	When  time.Time
}

// TotalKcal calculates energy of all eaten food.
func TotalKcal(reports []Report) uint32 {
	var ret uint32
	for _, r := range reports {
		ret += r.Kcal * r.Grams
	}
	return ret / 100
}
