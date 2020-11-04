package foodbot

import (
	"sort"
)

// AddProductKcal adds a new energy value for the given product.
func (b *Bot) AddProductKcal(food string, kcal uint32) {
	if _, ok := b.products[food]; !ok {
		b.products[food] = make(map[uint32]bool)
	}
	b.products[food][kcal] = true
	b.db.insertProduct(food, kcal)
}

// GetProductKcals returns a list of possible energy values for the given product.
// Returns false if such product has not been added before.
func (b *Bot) GetProductKcals(food string) ([]uint32, bool) {
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
