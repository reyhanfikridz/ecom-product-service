/*
Package utils containing utilities function

This package cannot have import from another package except for config package
*/
package utils

import (
	"testing"
)

// TestGetRandomSKU test GetRandomSKU
func TestGetRandomSKU(t *testing.T) {
	for i := 0; i < 1000; i++ {
		SKU := GetRandomSKU()
		if len(SKU) != 10 {
			t.Errorf("Expected SKU length 10, but got %d", len(SKU))
		}
	}
}
