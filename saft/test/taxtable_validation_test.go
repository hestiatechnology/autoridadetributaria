package test

import (
	"testing"

	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/masterfiles"
)

// Helper functions (can be shared)
func containsErrorMsgTaxTable(errs []error, msg string) bool {
	for _, err := range errs {
		if err.Error() == msg {
			return true
		}
	}
	return false
}

func containsErrorMsgPrefixTaxTable(errs []error, prefix string) bool {
	for _, err := range errs {
		// Using strings.HasPrefix might be cleaner if allowed/imported
		if len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func TestValidateTaxTable_Valid_FullEntry(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{
				TaxType:          "IVA",
				TaxCountryRegion: "PT",
				TaxCode:          "NOR",
				Description:      "IVA Normal Rate",
				TaxPercentage:    "23.00",
				TaxAmount:        "", // Optional if percentage is given
			},
			{
				TaxType:          "IS",
				TaxCountryRegion: "PT",
				TaxCode:          "ISENTO_ARTIGO_X",
				Description:      "Stamp Duty Exemption",
				TaxPercentage:    "0.00", // Or could be empty
				TaxAmount:        "",
			},
			{
				TaxType:          "NS", // Not Subject
				TaxCountryRegion: "PT",
				TaxCode:          "NA",
				Description:      "Not Subject to Tax",
				TaxPercentage:    "",
				TaxAmount:        "",
			},
			{ // Example with TaxAmount
				TaxType:          "IVA",
				TaxCountryRegion: "PT",
				TaxCode:          "FIX",
				Description:      "Fixed IVA Amount",
				TaxPercentage:    "",
				TaxAmount:        "10.50",
			},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid TaxTable, got %d: %v", len(errs), errs)
	}
}

func TestValidateTaxTable_Valid_EmptyEntries(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{}, // No entries, should be valid by this validator
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for TaxTable with empty entries, got %d: %v", len(errs), errs)
	}
}

func TestValidateTaxTable_Valid_NilTable(t *testing.T) {
	errs := masterfiles.ValidateTaxTable(nil) // Nil table, should be valid by this validator
	if len(errs) > 0 {
		t.Errorf("Expected no errors for nil TaxTable, got %d: %v", len(errs), errs)
	}
}


func TestValidateTaxTable_Invalid_TaxType(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "XXX", TaxCountryRegion: "PT", TaxCode: "NOR", Description: "Desc", TaxPercentage: "23"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	expectedPrefix := "TaxTableEntry[0] (TaxCode: NOR, TaxType: XXX): invalid TaxType 'XXX'"
	if !containsErrorMsgPrefixTaxTable(errs, expectedPrefix) {
		t.Errorf("Expected error starting with '%s', got: %v", expectedPrefix, errs)
	}
}

func TestValidateTaxTable_Invalid_EmptyTaxCountryRegion(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "", TaxCode: "NOR", Description: "Desc", TaxPercentage: "23"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	expectedPrefix := "TaxTableEntry[0] (TaxCode: NOR, TaxType: IVA): TaxCountryRegion is required"
	if !containsErrorMsgPrefixTaxTable(errs, expectedPrefix) {
		t.Errorf("Expected error starting with '%s', got: %v", expectedPrefix, errs)
	}
}

func TestValidateTaxTable_Invalid_TaxCountryRegionFormat(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "PTX", TaxCode: "NOR", Description: "Desc", TaxPercentage: "23"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	expectedPrefix := "TaxTableEntry[0] (TaxCode: NOR, TaxType: IVA): TaxCountryRegion 'PTX' is not a valid ISO 3166-1 alpha-2 code"
	if !containsErrorMsgPrefixTaxTable(errs, expectedPrefix) {
		t.Errorf("Expected error starting with '%s', got: %v", expectedPrefix, errs)
	}
}

func TestValidateTaxTable_Invalid_EmptyTaxCode(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "PT", TaxCode: "", Description: "Desc", TaxPercentage: "23"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	expectedPrefix := "TaxTableEntry[0] (TaxCode: , TaxType: IVA): TaxCode is required" // TaxCode is empty in identifier
	if !containsErrorMsgPrefixTaxTable(errs, expectedPrefix) {
		t.Errorf("Expected error starting with '%s', got: %v", expectedPrefix, errs)
	}
}

func TestValidateTaxTable_Invalid_EmptyDescription(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "PT", TaxCode: "NOR", Description: "", TaxPercentage: "23"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	expectedPrefix := "TaxTableEntry[0] (TaxCode: NOR, TaxType: IVA): Description is required"
	if !containsErrorMsgPrefixTaxTable(errs, expectedPrefix) {
		t.Errorf("Expected error starting with '%s', got: %v", expectedPrefix, errs)
	}
}

func TestValidateTaxTable_Invalid_TaxPercentageFormat(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "PT", TaxCode: "NOR", Description: "Desc", TaxPercentage: "ABC"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	expectedPrefix := "TaxTableEntry[0] (TaxCode: NOR, TaxType: IVA): TaxPercentage 'ABC' is not a valid decimal"
	if !containsErrorMsgPrefixTaxTable(errs, expectedPrefix) {
		t.Errorf("Expected error starting with '%s', got: %v", expectedPrefix, errs)
	}
}

func TestValidateTaxTable_Invalid_TaxAmountFormat(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "PT", TaxCode: "NOR", Description: "Desc", TaxAmount: "XYZ"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	expectedPrefix := "TaxTableEntry[0] (TaxCode: NOR, TaxType: IVA): TaxAmount 'XYZ' is not a valid decimal"
	if !containsErrorMsgPrefixTaxTable(errs, expectedPrefix) {
		t.Errorf("Expected error starting with '%s', got: %v", expectedPrefix, errs)
	}
}

func TestValidateTaxTable_Valid_TaxPercentageOnly(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "PT", TaxCode: "NOR", Description: "Desc", TaxPercentage: "23.00", TaxAmount: ""},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	if len(errs) > 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}
}

func TestValidateTaxTable_Valid_TaxAmountOnly(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "PT", TaxCode: "NOR", Description: "Desc", TaxPercentage: "", TaxAmount: "10.00"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	if len(errs) > 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}
}

func TestValidateTaxTable_Valid_BothTaxPercentageAndAmountZero(t *testing.T) {
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "PT", TaxCode: "ISE", Description: "IVA Isento", TaxPercentage: "0.00", TaxAmount: "0.00"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for zero percentage and amount, got: %v", errs)
	}
}

func TestValidateTaxTable_Valid_BothTaxPercentageAndAmountNonZero(t *testing.T) {
    // The validator does not forbid both being present, only that if present, they are valid decimals.
    // Specific SAFT rules might forbid this combination for certain TaxTypes/TaxCodes, but that's a deeper check.
	taxTable := &saft.TaxTable{
		TaxTableEntry: []saft.TaxTableEntry{
			{TaxType: "IVA", TaxCountryRegion: "PT", TaxCode: "NOR", Description: "Desc", TaxPercentage: "23.00", TaxAmount: "2.30"},
		},
	}
	errs := masterfiles.ValidateTaxTable(taxTable)
	if len(errs) > 0 {
		t.Errorf("Expected no errors when both TaxPercentage and TaxAmount are valid, got: %v", errs)
	}
}
