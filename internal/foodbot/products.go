package foodbot

import (
	"sort"
	"strings"
)

// AddProductKcal adds a new energy value for the given product.
func (b *Bot) AddProductKcal(name string, kcal uint32) {
	food := strings.ToLower(name)
	if _, ok := b.products[food]; !ok {
		b.products[food] = make(map[uint32]bool)
	}
	if _, ok := b.products[food][kcal]; !ok {
		b.db.insertProduct(food, kcal)
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
