/*
Package validator containing any validator function
*/
package validator

import (
	"fmt"
	"strings"

	"github.com/reyhanfikridz/ecom-product-service/internal/model"
)

// IsProductInfoValid check if product info data is valid
//
// return error nil if it's valid
func IsProductInfoValid(pi model.ProductInfo) error {
	if strings.TrimSpace(pi.Name) == "" {
		return fmt.Errorf("name empty/not found")
	}

	if pi.Price == 0 {
		return fmt.Errorf("price empty/not found")
	}

	if pi.Weight == 0 {
		return fmt.Errorf("weight empty/not found")
	}

	return nil
}
