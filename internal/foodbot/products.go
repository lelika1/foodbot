package foodbot

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Products and their energy values that have ever been added to the bot.
var Products map[string]map[uint32]bool

// ProductsFilePath is a path to the file there products are stored in between of restarts.
const ProductsFilePath = "products.db"

// UpdateProduct adds a new energy value for the given product.
func UpdateProduct(food string, kcal uint32) {
	if _, ok := Products[food]; !ok {
		Products[food] = make(map[uint32]bool)
	}
	Products[food][kcal] = true
	SaveProducts(ProductsFilePath)
}

// GetProductKcals returns a list of possible energy values for the given product.
// Returns false if such product has not been added before.
func GetProductKcals(food string) ([]uint32, bool) {
	if _, ok := Products[food]; !ok {
		return nil, false
	}

	var ret []uint32
	for k := range Products[food] {
		ret = append(ret, k)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return ret, true
}

// LoadProducts and their energy values from the given file.
func LoadProducts(path string) {
	Products = make(map[string]map[uint32]bool)

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	defer file.Close()
	if err != nil {
		log.Printf("os.OpenFile(%v) failed with %v\n", path, err)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		prodDescr := strings.Split(line, " ")
		if len(prodDescr) != 2 {
			continue
		}

		Products[prodDescr[0]] = make(map[uint32]bool)
		for _, kcalStr := range strings.Split(prodDescr[1], ",") {
			kcal, err := strconv.ParseUint(kcalStr, 10, 32)
			if err == nil {
				Products[prodDescr[0]][uint32(kcal)] = true
			}
		}
	}

}

// SaveProducts and their energy values to the given file.
func SaveProducts(path string) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer file.Close()
	if err != nil {
		log.Panic(fmt.Errorf("os.OpenFile(%v) failed with %v", path, err))
	}

	for name, kcals := range Products {
		var sb strings.Builder
		sb.WriteString(name)
		sb.WriteByte(' ')

		var kcalStr []string
		for kcal := range kcals {
			kcalStr = append(kcalStr, fmt.Sprint(kcal))
		}

		sb.WriteString(strings.Join(kcalStr, ","))
		sb.WriteByte('\n')
		file.WriteString(sb.String())
	}
}
