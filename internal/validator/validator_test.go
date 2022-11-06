/*
Package validator containing any validator function
*/
package validator

import (
	"fmt"
	"testing"

	"github.com/reyhanfikridz/ecom-product-service/internal/model"
)

// TestIsProductInfoValid test IsProductInfoValid
func TestIsProductInfoValid(t *testing.T) {
	// initialize testing table
	testTable := []struct {
		TestName       string
		Product        model.ProductInfo
		ExpectedResult error
	}{
		{
			TestName: "Test Form Complete",
			Product: model.ProductInfo{
				Name:   "test product",
				Price:  1000000.50,
				Weight: 1.52,
				Stock:  100,
			},
			ExpectedResult: nil,
		},
		{
			TestName: "Test Form Incomplete 1",
			Product: model.ProductInfo{
				Name:   "",
				Price:  1000000.50,
				Weight: 1.52,
				Stock:  100,
			},
			ExpectedResult: fmt.Errorf("name empty/not found"),
		},
		{
			TestName: "Test Form Incomplete 2",
			Product: model.ProductInfo{
				Name:   "test product",
				Price:  0,
				Weight: 1.52,
				Stock:  100,
			},
			ExpectedResult: fmt.Errorf("price empty/not found"),
		},
		{
			TestName: "Test Form Incomplete 3",
			Product: model.ProductInfo{
				Name:   "test product",
				Price:  1000000.50,
				Weight: 0,
				Stock:  100,
			},
			ExpectedResult: fmt.Errorf("weight empty/not found"),
		},
	}

	// Do the test
	for _, test := range testTable {
		err := IsProductInfoValid(test.Product)
		if test.ExpectedResult == nil && err != nil {
			t.Errorf("Expected product info valid, but got invalid => %s", err.Error())
		} else if test.ExpectedResult != nil {
			if err == nil {
				t.Errorf("Expected product info invalid, but got valid")
			} else if test.ExpectedResult.Error() != err.Error() {
				t.Errorf("Expected error '" +
					test.ExpectedResult.Error() + "' got '" + err.Error() + "'")
			}
		}
	}
}
