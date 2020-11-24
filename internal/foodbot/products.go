package foodbot

import (
	"sort"
	"strings"

	"github.com/lelika1/foodbot/internal/sqlite"
)

// products stores information as (product's name: {kcal1: true, kcal2: true}).
type products map[string]map[uint32]bool

func newProducts(list []sqlite.Product) products {
	ret := make(map[string]map[uint32]bool)
	for _, p := range list {
		if _, ok := ret[p.Name]; !ok {
			ret[p.Name] = make(map[uint32]bool)
		}
		ret[p.Name][p.Kcal] = true
	}
	return ret
}

func (p products) similar(name string) []sqlite.Product {
	pattern := normalize(name)
	if _, ok := p[pattern]; !ok {
		return nil
	}

	var ret []sqlite.Product
	for k := range p[pattern] {
		ret = append(ret, sqlite.Product{Name: pattern, Kcal: k})
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Kcal < ret[j].Kcal })
	return ret
}

// returns true if (name, kcal) was new, and false otherwise.
func (p products) addProduct(name string, kcal uint32) bool {
	food := normalize(name)

	isNew := false
	if _, ok := p[food]; !ok {
		p[food] = make(map[uint32]bool)
		isNew = true
	}
	if _, ok := p[food][kcal]; !ok {
		p[food][kcal] = true
		isNew = true
	}
	return isNew
}

func normalize(pattern string) string {
	return strings.Trim(strings.ToLower(pattern), "\n\t\r, .\"'")
}
