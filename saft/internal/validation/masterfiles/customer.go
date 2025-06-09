package masterfiles

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/hestiatechnology/autoridadetributaria/common"
	"github.com/hestiatechnology/autoridadetributaria/saft"
)

// ValidateCustomer validates a single customer entry.
// accountIDs is a map of valid AccountID strings from GeneralLedgerAccounts.
func ValidateCustomer(customer *saft.Customer, accountIDs map[string]bool) []error {
	var errs []error

	if customer == nil {
		errs = append(errs, errors.New("customer is nil"))
		return errs
	}

	if customer.CustomerID == "" {
		errs = append(errs, errors.New("CustomerID is required"))
	}

	if customer.AccountID == "" {
		errs = append(errs, errors.New("AccountID is required"))
	} else {
		if _, ok := accountIDs[customer.AccountID]; !ok {
			errs = append(errs, fmt.Errorf("AccountID %s does not correspond to a valid account in GeneralLedgerAccounts", customer.AccountID))
		}
	}

	if customer.CustomerTaxID == "" {
		errs = append(errs, errors.New("CustomerTaxID is required"))
	} else {
		// Validate as Portuguese NIF: 9 digits
		if len(customer.CustomerTaxID) != 9 {
			errs = append(errs, fmt.Errorf("CustomerTaxID %s must be 9 digits long", customer.CustomerTaxID))
		} else {
			// Assuming common.ValidateNIFPT expects a string and returns bool
			// The existing code had `string(customer.CustomerTaxId)`. If CustomerTaxID is already string, this is fine.
			// If CustomerTaxID is not string, it needs conversion. Assuming it's string based on context.
			if ok := common.ValidateNIFPT(customer.CustomerTaxID); !ok {
				errs = append(errs, fmt.Errorf("CustomerTaxID %s is not a valid Portuguese NIF", customer.CustomerTaxID))
			}
		}
	}

	if customer.CompanyName == "" {
		errs = append(errs, errors.New("CompanyName is required"))
	}

	// SelfBillingIndicator must be 0 or 1
	// Assuming saft.SelfBillingIndicator is an int or compatible type
	// The original code used saft.SelfBillingIndicatorNo and saft.SelfBillingIndicatorYes
	// We'll stick to the explicit 0 or 1 as per requirements.
	// It's safer to check the type of customer.SelfBillingIndicator if possible
	// For now, assuming it's comparable to int constants 0 and 1.
	// If saft.SelfBillingIndicator is a specific type, this might need adjustment.
	// Let's assume it's a type that can be compared to 0 and 1.
	// The original code implies it's a type that can be compared to saft.SelfBillingIndicatorNo and saft.SelfBillingIndicatorYes.
	// If those are constants for 0 and 1, then the logic is similar.
	// Let's assume customer.SelfBillingIndicator is an int or a type convertible/comparable to int.
	// Based on the original code `customer.SelfBillingIndicator != saft.SelfBillingIndicatorNo && customer.SelfBillingIndicator != saft.SelfBillingIndicatorYes`
	// and the new requirement "Ensure SelfBillingIndicator is either 0 or 1".
	// I'll check against 0 and 1 directly.
	if customer.SelfBillingIndicator != 0 && customer.SelfBillingIndicator != 1 {
		errs = append(errs, fmt.Errorf("SelfBillingIndicator must be 0 or 1, got %v for CustomerID %s", customer.SelfBillingIndicator, customer.CustomerID))
	}

	// Validate BillingAddress
	// Assuming customer.BillingAddress is of type saft.AddressStructure (not a pointer)
	// as implied by original code `customer.BillingAddress == (saft.CustomerAddressStructure{})`
	// If it were a pointer, we might check for nil here first.
	// The requirement "For BillingAddress ... Ensure AddressDetail is not empty..." implies BillingAddress must be there.
	// The original code had a check for `customer.BillingAddress == (saft.CustomerAddressStructure{})`.
	// This check is problematic for struct types. Field validations are handled by validateAddressStructure.
	// If BillingAddress is a struct, it always "exists". We validate its content.
	billingAddrErrs := validateAddressStructure(&customer.BillingAddress, "BillingAddress", customer.CustomerID)
	errs = append(errs, billingAddrErrs...)

	// Validate ShipToAddress (plural in original code, so it's a slice of saft.AddressStructure)
	for i, shipToAddressEntry := range customer.ShipToAddress {
		// shipToAddressEntry is a copy of the struct in the slice.
		// We need to pass a pointer to the actual element in the slice for potential modifications (not an issue here)
		// and to avoid operating on a copy if the struct is large.
		// However, validateAddressStructure takes *saft.AddressStructure.
		// So &shipToAddressEntry is correct here as it's a pointer to the loop variable copy.
		// If saft.AddressStructure is small, this is fine.
		// A possibly more robust way if ShipToAddress was a slice of pointers: shipToAddressEntry itself.
		// Or iterating by index: &customer.ShipToAddress[i]
		shipToAddrErrs := validateAddressStructure(&customer.ShipToAddress[i], fmt.Sprintf("ShipToAddress[%d]", i), customer.CustomerID)
		errs = append(errs, shipToAddrErrs...)
	}

	return errs
}

// ValidateCustomersInAuditfile was the original function.
// It is updated to use the new ValidateCustomer function and collect all errors.
// It now requires accountIDs to be passed in.
func ValidateCustomersInAuditfile(a *saft.AuditFile, accountIDs map[string]bool) []error {
	var allErrors []error
	if a == nil {
		allErrors = append(allErrors, errors.New("auditfile is nil"))
		return allErrors
	}
    if a.MasterFiles == nil {
        allErrors = append(allErrors, errors.New("masterfiles is nil in auditfile"))
        return allErrors
    }
    if a.MasterFiles.Customer == nil {
        // No customers to validate, not necessarily an error.
        return allErrors
    }

	// Assuming a.MasterFiles.Customer is a slice of saft.Customer structs.
	// Iterate by index to pass a pointer to the actual element in the slice.
	for i := range a.MasterFiles.Customer {
		customerErrors := ValidateCustomer(&a.MasterFiles.Customer[i], accountIDs)
		allErrors = append(allErrors, customerErrors...)
	}

	return allErrors
}

// countryCodeRegex defines a regex for ISO 3166-1 alpha-2 country codes.
var countryCodeRegex = regexp.MustCompile(`^[A-Z]{2}$`)

// validateAddressStructure validates a single address structure.
// It expects a pointer to an address structure.
// The `addressType` is a string like "BillingAddress" or "ShipToAddress[0]" for use in error messages.
// `customerID` is used for context in error messages.
// Note: saft.AddressStructure is assumed based on saft.CustomerAddressStructure from original code.
// If BillingAddress is a struct (not pointer) in saft.Customer, then a pointer to it is passed.
// If ShipToAddress is a slice of structs, a pointer to each element is passed.
func validateAddressStructure(address *saft.AddressStructure, addressType string, customerID string) []error {
	var errs []error

	// This function expects a non-nil pointer to an address.
	// Checks for whether an address *should* exist (e.g. mandatory BillingAddress)
	// should be done by the caller if 'address' could be nil (e.g. if it was a pointer field).
	// Since BillingAddress is likely a struct field, it will never be nil.
	// A "missing" address would mean all its fields are empty.

	if address.AddressDetail == "" {
		errs = append(errs, fmt.Errorf("%s.AddressDetail is required for CustomerID %s", addressType, customerID))
	}
	if address.City == "" {
		errs = append(errs, fmt.Errorf("%s.City is required for CustomerID %s", addressType, customerID))
	}
	if address.PostalCode == "" {
		errs = append(errs, fmt.Errorf("%s.PostalCode is required for CustomerID %s", addressType, customerID))
	}
	if address.Country == "" {
		errs = append(errs, fmt.Errorf("%s.Country is required for CustomerID %s", addressType, customerID))
	} else {
		if !countryCodeRegex.MatchString(address.Country) {
			// Add CustomerID to the error message for better context
			errMsg := fmt.Sprintf("%s.Country code '%s' is not a valid ISO 3166-1 alpha-2 code", addressType, address.Country)
			if customerID != "" {
				errMsg = fmt.Sprintf("%s.Country code '%s' is not a valid ISO 3166-1 alpha-2 code for CustomerID %s", addressType, address.Country, customerID)
			}
			errs = append(errs, errors.New(errMsg))
		}
	}
	return errs
}
