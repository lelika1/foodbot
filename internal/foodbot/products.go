package foodbot

import (
	"sort"
	"strings"

	"github.com/lelika1/foodbot/internal/sqlite"
)

// Products stores information about products as (name: {kcal1: true, kcal2: true}).
type Products map[string]map[uint32]bool

func createProducts(products []sqlite.Product) Products {
	ret := make(Products)
	for _, p := range products {
		if _, ok := ret[p.Name]; !ok {
			ret[p.Name] = make(map[uint32]bool)
		}
		ret[p.Name][p.Kcal] = true
	}
	return ret
}

// AddProductKcal adds a new energy value for the given product.
func (b *Bot) AddProductKcal(name string, kcal uint32) {
	food := strings.ToLower(name)
	if _, ok := b.products[food]; !ok {
		b.products[food] = make(map[uint32]bool)
	}
	if _, ok := b.products[food][kcal]; !ok {
		b.SaveProduct(food, kcal)
		b.products[food][kcal] = true
	}
}

// GetProductKcals returns a list of possible energy values for the given product.
// Returns false if such product has not been added before.
func (b *Bot) GetProductKcals(name string) ([]uint32, bool) {
	food := strings.ToLower(name)
	if _, ok := b.products[food]; !ok {
		return nil, false
	}

	var ret []uint32
	for k := range b.products[food] {
		ret = append(ret, k)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return ret, true
}
