package masterfiles

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hestiatechnology/autoridadetributaria/saft"
)

// countryCodeRegex defines a regex for ISO 3166-1 alpha-2 country codes.
// Redefined here assuming it's not yet in a shared utility file.
var countryCodeRegex = regexp.MustCompile(`^[A-Z]{2}$`)

// isValidDecimal checks if a string can be parsed as a valid float.
// SAFT typically uses dot as decimal separator.
func isValidDecimal(s string) bool {
	if s == "" { // Not provided, so not invalid by this check alone. Caller decides if required.
		return true
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// ValidateTaxTable validates the TaxTable section of the SAFT file.
// Assumes saft.TaxTable and saft.TaxTableEntry are defined.
func ValidateTaxTable(taxTable *saft.TaxTable) []error {
	var errs []error

	if taxTable == nil {
		// A nil taxTable might be acceptable if there are no taxes to report.
		// However, if the MasterFiles section exists, it usually implies content.
		// For now, let's assume a nil taxTable itself isn't an error, but an empty one might be checked elsewhere.
		// If it's present but has no entries, the loop just won't run.
		return errs // No errors if taxTable is nil.
	}

	if len(taxTable.TaxTableEntry) == 0 {
		// This might be an error depending on SAFT rules (e.g. at least one entry expected if table exists).
		// For now, just validating entries if they exist.
		// errs = append(errs, errors.New("TaxTable must contain at least one TaxTableEntry"))
	}

	for i, entry := range taxTable.TaxTableEntry {
		entryIdentifier := fmt.Sprintf("TaxTableEntry[%d] (TaxCode: %s, TaxType: %s)", i, entry.TaxCode, entry.TaxType)

		// Validate TaxType (IVA, IS, NS)
		validTaxTypes := map[string]bool{
			"IVA": true, // Value Added Tax
			"IS":  true, // Stamp Duty
			"NS":  true, // Not subject to tax / Exempt
		}
		if _, ok := validTaxTypes[entry.TaxType]; !ok {
			errs = append(errs, fmt.Errorf("%s: invalid TaxType '%s'. Must be one of IVA, IS, NS", entryIdentifier, entry.TaxType))
		}

		// Validate TaxCountryRegion (not empty, valid country code)
		if entry.TaxCountryRegion == "" {
			errs = append(errs, fmt.Errorf("%s: TaxCountryRegion is required", entryIdentifier))
		} else {
			if !countryCodeRegex.MatchString(entry.TaxCountryRegion) {
				errs = append(errs, fmt.Errorf("%s: TaxCountryRegion '%s' is not a valid ISO 3166-1 alpha-2 code", entryIdentifier, entry.TaxCountryRegion))
			}
		}

		// Validate TaxCode (not empty)
		if entry.TaxCode == "" {
			errs = append(errs, fmt.Errorf("%s: TaxCode is required", entryIdentifier))
		}

		// Validate Description (not empty)
		if entry.Description == "" {
			errs = append(errs, fmt.Errorf("%s: Description is required", entryIdentifier))
		}

		// Validate TaxPercentage (if provided, ensure it's a valid decimal)
		// Assuming entry.TaxPercentage is a string field.
		// If it's a float64, the check is different (e.g., range check, or just that it's present if it's a pointer).
		// If it's a string that can be empty.
		if entry.TaxPercentage != "" && !isValidDecimal(entry.TaxPercentage) {
			errs = append(errs, fmt.Errorf("%s: TaxPercentage '%s' is not a valid decimal", entryIdentifier, entry.TaxPercentage))
		}

		// Validate TaxAmount (if provided, ensure it's a valid decimal)
		// Assuming entry.TaxAmount is a string field.
		// If it's a float64, the check is different.
		// If it's a string that can be empty.
		if entry.TaxAmount != "" && !isValidDecimal(entry.TaxAmount) {
			errs = append(errs, fmt.Errorf("%s: TaxAmount '%s' is not a valid decimal", entryIdentifier, entry.TaxAmount))
		}
	}

	return errs
}
