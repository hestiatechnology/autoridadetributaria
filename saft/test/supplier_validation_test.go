package test

import (
	"testing"

	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/masterfiles"
)

// Reusing testValidAccountIDs from customer_validation_test.go scope (conceptually)
// In a real scenario, this might be a shared var or initialized per test suite/file.
// var testValidAccountIDs = map[string]bool{"ACC_SUP1": true, "ACC_GEN": true}

// Helper functions (can be shared)
func containsErrorMsgSupplier(errs []error, msg string) bool {
	for _, err := range errs {
		if err.Error() == msg {
			return true
		}
	}
	return false
}

func containsErrorMsgPrefixSupplier(errs []error, prefix string) bool {
	for _, err := range errs {
		if len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}


func TestValidateSupplier_Valid(t *testing.T) {
	supplier := &saft.Supplier{
		SupplierID:      "SUP001",
		AccountID:       "ACC_SUP1", // Assuming this is a valid account ID passed in the map
		SupplierTaxID:   "509876543", // Valid NIF structure
		CompanyName:     "Valid Supplier Co.",
		SelfBillingIndicator: "0", // String "0" or "1"
		BillingAddress: saft.AddressStructure{
			AddressDetail: "Supplier Street 1",
			City:          "Supplier City",
			PostalCode:    "2000-001",
			Country:       "PT",
		},
		// ShipFromAddress is optional in some contexts, or can be empty slice if not applicable
		ShipFromAddress: []saft.AddressStructure{},
	}
	// Mock account IDs including the one used by the supplier
	mockAccountIDs := map[string]bool{"ACC_SUP1": true, "OTHER_ACC": true}

	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid supplier, got %d: %v", len(errs), errs)
	}
}

func TestValidateSupplier_Invalid_EmptySupplierID(t *testing.T) {
	supplier := &saft.Supplier{SupplierID: "", AccountID: "ACC_SUP1", SupplierTaxID: "509876543", CompanyName: "Test Sup", BillingAddress: saft.AddressStructure{AddressDetail: "1", City: "C", PostalCode: "P", Country: "PT"}, SelfBillingIndicator: "0"}
	mockAccountIDs := map[string]bool{"ACC_SUP1": true}
	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	if !containsErrorMsgSupplier(errs, "SupplierID is required") {
		t.Errorf("Expected 'SupplierID is required', got: %v", errs)
	}
}

func TestValidateSupplier_Invalid_AccountIDNotFound(t *testing.T) {
	supplier := &saft.Supplier{SupplierID: "SUP001", AccountID: "INVALID_ACC_SUP", SupplierTaxID: "509876543", CompanyName: "Test Sup", BillingAddress: saft.AddressStructure{AddressDetail: "1", City: "C", PostalCode: "P", Country: "PT"}, SelfBillingIndicator: "0"}
	mockAccountIDs := map[string]bool{"ACC_SUP1": true} // INVALID_ACC_SUP is not here
	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	expectedMsg := "AccountID INVALID_ACC_SUP for SupplierID SUP001 does not correspond to a valid account in GeneralLedgerAccounts"
	if !containsErrorMsgSupplier(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSupplier_Invalid_SupplierTaxID_Format(t *testing.T) {
	supplier := &saft.Supplier{SupplierID: "SUP001", AccountID: "ACC_SUP1", SupplierTaxID: "123", CompanyName: "Test Sup", BillingAddress: saft.AddressStructure{AddressDetail: "1", City: "C", PostalCode: "P", Country: "PT"}, SelfBillingIndicator: "0"}
	mockAccountIDs := map[string]bool{"ACC_SUP1": true}
	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	expectedMsg := "SupplierTaxID 123 for SupplierID SUP001 must be 9 digits long"
	if !containsErrorMsgSupplier(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSupplier_Invalid_SupplierTaxID_InvalidNIF(t *testing.T) {
	// 500000000 is often an invalid NIF by checksum
	supplier := &saft.Supplier{SupplierID: "SUP001", AccountID: "ACC_SUP1", SupplierTaxID: "500000000", CompanyName: "Test Sup", BillingAddress: saft.AddressStructure{AddressDetail: "1", City: "C", PostalCode: "P", Country: "PT"}, SelfBillingIndicator: "0"}
	mockAccountIDs := map[string]bool{"ACC_SUP1": true}
	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	expectedMsg := "SupplierTaxID 500000000 for SupplierID SUP001 is not a valid Portuguese NIF"
	if !containsErrorMsgSupplier(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSupplier_Invalid_EmptyCompanyName(t *testing.T) {
	supplier := &saft.Supplier{SupplierID: "SUP001", AccountID: "ACC_SUP1", SupplierTaxID: "509876543", CompanyName: "", BillingAddress: saft.AddressStructure{AddressDetail: "1", City: "C", PostalCode: "P", Country: "PT"}, SelfBillingIndicator: "0"}
	mockAccountIDs := map[string]bool{"ACC_SUP1": true}
	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	expectedMsg := "CompanyName is required for SupplierID SUP001"
	if !containsErrorMsgSupplier(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSupplier_Invalid_BillingAddress_EmptyCity(t *testing.T) {
	supplier := &saft.Supplier{
		SupplierID: "SUP001", AccountID: "ACC_SUP1", SupplierTaxID: "509876543", CompanyName: "Test Sup",
		BillingAddress: saft.AddressStructure{AddressDetail: "1", City: "", PostalCode: "P", Country: "PT"}, // Empty City
		SelfBillingIndicator: "0",
	}
	mockAccountIDs := map[string]bool{"ACC_SUP1": true}
	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	expectedMsg := "BillingAddress.City is required for SupplierID SUP001"
	if !containsErrorMsgSupplier(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSupplier_Invalid_BillingAddress_InvalidCountry(t *testing.T) {
	supplier := &saft.Supplier{
		SupplierID: "SUP001", AccountID: "ACC_SUP1", SupplierTaxID: "509876543", CompanyName: "Test Sup",
		BillingAddress: saft.AddressStructure{AddressDetail: "1", City: "Supplier City", PostalCode: "P", Country: "PTX"}, // Invalid Country
		SelfBillingIndicator: "0",
	}
	mockAccountIDs := map[string]bool{"ACC_SUP1": true}
	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	expectedMsg := "BillingAddress.Country code 'PTX' is not a valid ISO 3166-1 alpha-2 code for SupplierID SUP001"
	if !containsErrorMsgSupplier(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSupplier_Invalid_ShipFromAddress_EmptyDetail(t *testing.T) {
	supplier := &saft.Supplier{
		SupplierID: "SUP001", AccountID: "ACC_SUP1", SupplierTaxID: "509876543", CompanyName: "Test Sup",
		BillingAddress: saft.AddressStructure{AddressDetail: "1", City: "Supplier City", PostalCode: "P", Country: "PT"},
		ShipFromAddress: []saft.AddressStructure{ // Note: ShipFromAddress for suppliers
			{AddressDetail: "", City: "Warehouse City", PostalCode: "3000", Country: "PT"}, // Empty AddressDetail
		},
		SelfBillingIndicator: "0",
	}
	mockAccountIDs := map[string]bool{"ACC_SUP1": true}
	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	expectedMsg := "ShipFromAddress[0].AddressDetail is required for SupplierID SUP001"
	if !containsErrorMsgSupplier(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSupplier_Invalid_SelfBillingIndicator(t *testing.T) {
	supplier := &saft.Supplier{
		SupplierID: "SUP001", AccountID: "ACC_SUP1", SupplierTaxID: "509876543", CompanyName: "Test Sup",
		BillingAddress: saft.AddressStructure{AddressDetail: "1", City: "C", PostalCode: "P", Country: "PT"},
		SelfBillingIndicator: "2", // Invalid, must be string "0" or "1"
	}
	mockAccountIDs := map[string]bool{"ACC_SUP1": true}
	errs := masterfiles.ValidateSupplier(supplier, mockAccountIDs)
	expectedMsg := "SelfBillingIndicator must be string \"0\" or \"1\", got \"2\" for SupplierID SUP001"
	if !containsErrorMsgSupplier(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSupplier_NilInput(t *testing.T) {
	mockAccountIDs := map[string]bool{}
	errs := masterfiles.ValidateSupplier(nil, mockAccountIDs)
	if !containsErrorMsgSupplier(errs, "supplier is nil") {
		t.Errorf("Expected 'supplier is nil', got %v", errs)
	}
}
