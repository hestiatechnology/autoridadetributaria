package masterfiles

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/hestiatechnology/autoridadetributaria/common"
	"github.com/hestiatechnology/autoridadetributaria/saft"
)

// countryCodeRegex defines a regex for ISO 3166-1 alpha-2 country codes.
// Redefined here assuming it's not yet in a shared utility file.
var countryCodeRegex = regexp.MustCompile(`^[A-Z]{2}$`)

// validateAddressStructure validates a single address structure for suppliers.
// Redefined here assuming it's not yet in a shared utility file or that Supplier addresses might differ.
// It expects a pointer to an address structure (e.g. saft.AddressStructure).
// `addressType` is a string like "BillingAddress" or "ShipFromAddress[0]" for use in error messages.
// `supplierID` is used for context in error messages.
func validateSupplierAddressStructure(address *saft.AddressStructure, addressType string, supplierID string) []error {
	var errs []error

	if address.AddressDetail == "" {
		errs = append(errs, fmt.Errorf("%s.AddressDetail is required for SupplierID %s", addressType, supplierID))
	}
	if address.City == "" {
		errs = append(errs, fmt.Errorf("%s.City is required for SupplierID %s", addressType, supplierID))
	}
	if address.PostalCode == "" {
		errs = append(errs, fmt.Errorf("%s.PostalCode is required for SupplierID %s", addressType, supplierID))
	}
	if address.Country == "" {
		errs = append(errs, fmt.Errorf("%s.Country is required for SupplierID %s", addressType, supplierID))
	} else {
		if !countryCodeRegex.MatchString(address.Country) {
			errMsg := fmt.Sprintf("%s.Country code '%s' is not a valid ISO 3166-1 alpha-2 code for SupplierID %s", addressType, address.Country, supplierID)
			errs = append(errs, errors.New(errMsg))
		}
	}
	return errs
}

// ValidateSupplier validates a single supplier entry.
// accountIDs is a map of valid AccountID strings from GeneralLedgerAccounts.
// Assumes saft.Supplier and saft.AddressStructure are defined in the saft package.
func ValidateSupplier(supplier *saft.Supplier, accountIDs map[string]bool) []error {
	var errs []error

	if supplier == nil {
		errs = append(errs, errors.New("supplier is nil"))
		return errs // Early exit if supplier is nil
	}

	// Validate SupplierID
	if supplier.SupplierID == "" {
		errs = append(errs, errors.New("SupplierID is required"))
	}

	// Validate AccountID
	if supplier.AccountID == "" {
		errs = append(errs, fmt.Errorf("AccountID is required for SupplierID %s", supplier.SupplierID))
	} else {
		if _, ok := accountIDs[supplier.AccountID]; !ok {
			errs = append(errs, fmt.Errorf("AccountID %s for SupplierID %s does not correspond to a valid account in GeneralLedgerAccounts", supplier.AccountID, supplier.SupplierID))
		}
	}

	// Validate SupplierTaxID
	if supplier.SupplierTaxID == "" {
		errs = append(errs, fmt.Errorf("SupplierTaxID is required for SupplierID %s", supplier.SupplierID))
	} else {
		// Validate as Portuguese NIF: 9 digits
		// Assuming SupplierTaxID is a string.
		if len(supplier.SupplierTaxID) != 9 {
			errs = append(errs, fmt.Errorf("SupplierTaxID %s for SupplierID %s must be 9 digits long", supplier.SupplierTaxID, supplier.SupplierID))
		} else if !common.ValidateNIFPT(supplier.SupplierTaxID) {
			errs = append(errs, fmt.Errorf("SupplierTaxID %s for SupplierID %s is not a valid Portuguese NIF", supplier.SupplierTaxID, supplier.SupplierID))
		}
	}

	// Validate CompanyName
	if supplier.CompanyName == "" {
		errs = append(errs, fmt.Errorf("CompanyName is required for SupplierID %s", supplier.SupplierID))
	}

	// Validate BillingAddress
	// Assuming supplier.BillingAddress is of type saft.AddressStructure (not a pointer).
	billingAddrErrs := validateSupplierAddressStructure(&supplier.BillingAddress, "BillingAddress", supplier.SupplierID)
	errs = append(errs, billingAddrErrs...)

	// Validate ShipFromAddress (plural, so it's a slice of saft.AddressStructure)
	// Assuming saft.Supplier uses ShipFromAddress []saft.AddressStructure
	for i := range supplier.ShipFromAddress {
		// Pass a pointer to the actual element in the slice.
		shipAddrErrs := validateSupplierAddressStructure(&supplier.ShipFromAddress[i], fmt.Sprintf("ShipFromAddress[%d]", i), supplier.SupplierID)
		errs = append(errs, shipAddrErrs...)
	}

	// Validate SelfBillingIndicator - must be STRING "0" or "1"
	if supplier.SelfBillingIndicator != "0" && supplier.SelfBillingIndicator != "1" {
		errMsg := fmt.Sprintf("SelfBillingIndicator must be string \"0\" or \"1\", got \"%v\" for SupplierID %s", supplier.SelfBillingIndicator, supplier.SupplierID)
		errs = append(errs, errors.New(errMsg))
	}

	return errs
}
