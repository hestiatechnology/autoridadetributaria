package masterfiles

import (
	"errors"
	"fmt"

	"github.com/hestiatechnology/autoridadetributaria/saft"
)

// ValidateProduct validates a single product entry.
// Assumes saft.Product is defined in the saft package.
func ValidateProduct(product *saft.Product) []error {
	var errs []error

	if product == nil {
		errs = append(errs, errors.New("product is nil"))
		return errs // Early exit if product is nil
	}

	// Validate ProductType
	// Ensure ProductType is valid (P, S, O, I, E).
	validProductTypes := map[string]bool{
		"P": true, // Products
		"S": true, // Services
		"O": true, // Other
		"I": true, // Intermediary products / Work in progress
		"E": true, // Fixed Assets
	}
	if _, ok := validProductTypes[product.ProductType]; !ok {
		errs = append(errs, fmt.Errorf("ProductType '%s' for ProductCode '%s' is invalid. Must be one of P, S, O, I, E", product.ProductType, product.ProductCode))
	}

	// Validate ProductCode
	// Ensure ProductCode is not empty.
	if product.ProductCode == "" {
		errs = append(errs, errors.New("ProductCode is required"))
	}

	// Validate ProductDescription
	// Ensure ProductDescription is not empty.
	if product.ProductDescription == "" {
		errs = append(errs, fmt.Errorf("ProductDescription is required for ProductCode '%s'", product.ProductCode))
	}

	// Validate ProductNumberCode
	// Ensure ProductNumberCode is not empty.
	// This typically is the EAN or other barcode.
	if product.ProductNumberCode == "" {
		errs = append(errs, fmt.Errorf("ProductNumberCode is required for ProductCode '%s'", product.ProductCode))
	}

	return errs
}
