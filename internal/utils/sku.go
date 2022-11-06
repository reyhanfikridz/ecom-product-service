/*
Package utils containing utilities function

This package cannot have import from another package except for config package
*/
package utils

import (
	"math/rand"
	"time"
)

// GetRandomSKU get random SKU
func GetRandomSKU() string {
	// change rand seed so the result is different everytime program running
	rand.Seed(time.Now().UnixNano())

	// set SKU length
	n := 10

	// get random sku
	base := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	runeOfSKU := make([]rune, n)
	for i := range runeOfSKU {
		runeOfSKU[i] = base[rand.Intn(len(base))]
	}

	return string(runeOfSKU)
}
