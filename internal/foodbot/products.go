package foodbot

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/lelika1/foodbot/internal/sqlite"
)

// Product is alias for sqlite.Product.
type Product sqlite.Product

func (p *Product) String() string {
	return fmt.Sprintf("%v %v kcal", p.Name, p.Kcal)
}

// products stores information as (product's name: {kcal1: true, kcal2: true}).
// Also stores cache of last added products.
type products struct {
	all       map[string]map[uint32]bool
	cache     []string
	cacheSize int
}

func newProducts(list []sqlite.Product) products {
	all := make(map[string]map[uint32]bool)
	for _, p := range list {
		if _, ok := all[p.Name]; !ok {
			all[p.Name] = make(map[uint32]bool)
		}
		all[p.Name][p.Kcal] = true
	}
	return products{all: all, cacheSize: 5}
}

func (p *products) lastAdded() []Product {
	var ret []Product = make([]Product, len(p.cache))
	for i, cached := range p.cache {
		json.Unmarshal([]byte(cached), &ret[i])
	}
	return ret
}

func (p *products) similar(name string) []Product {
	pattern := normalize(name)
	if _, ok := p.all[pattern]; !ok {
		return nil
	}

	var ret []Product
	for k := range p.all[pattern] {
		ret = append(ret, Product{pattern, k})
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Kcal < ret[j].Kcal })
	return ret
}

// returns true if (name, kcal) was new, and false otherwise.
func (p *products) addProduct(name string, kcal uint32) bool {
	food := normalize(name)
	if p.updateCache(Product{food, kcal}) {
		return false
	}

	isNew := false
	if _, ok := p.all[food]; !ok {
		p.all[food] = make(map[uint32]bool)
		isNew = true
	}
	if _, ok := p.all[food][kcal]; !ok {
		p.all[food][kcal] = true
		isNew = true
	}
	return isNew
}

// true - already in cache, false - otherwise.
func (p *products) updateCache(product Product) bool {
	bytes, _ := json.Marshal(product)
	added := string(bytes)

	for _, cached := range p.cache {
		if cached == added {
			return true
		}
	}

	p.cache = append(p.cache, added)
	if len(p.cache) > p.cacheSize {
		p.cache = p.cache[1:]
	}
	return false
}

func normalize(pattern string) string {
	return strings.Trim(strings.ToLower(pattern), "\n\t\r, .\"'")
}
