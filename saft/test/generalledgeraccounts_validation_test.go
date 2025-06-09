package test

import (
	"testing"

	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/masterfiles" // Assuming this is the package for GLA validator
)

func TestValidateGeneralLedgerAccounts_Valid(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "SNC-BASE",
		Account: []saft.Account{
			{
				AccountID:          "111",
				AccountDescription: "Cash",
				OpeningDebitBalance: "1000.00",
				OpeningCreditBalance: "0.00",
				ClosingDebitBalance: "1200.00",
				ClosingCreditBalance: "0.00",
				GroupingCategory:   "AR", // Assets Rec.
				GroupingCode:       "",   // Optional for AR
				TaxonomyCode:       "",   // Optional for AR
			},
			{
				AccountID:          "211",
				AccountDescription: "Trade Debtors - National Market",
				OpeningDebitBalance: "500.00",
				OpeningCreditBalance: "0.00",
				ClosingDebitBalance: "700.00",
				ClosingCreditBalance: "0.00",
				GroupingCategory:   "GA", // Assets
				GroupingCode:       "111", // Assuming 111 is a valid AccountID for grouping
				TaxonomyCode:       "",    // Optional for GA
			},
			{
				AccountID:          "711",
				AccountDescription: "Sales - Products",
				OpeningDebitBalance: "0.00",
				OpeningCreditBalance: "2000.00",
				ClosingDebitBalance: "0.00",
				ClosingCreditBalance: "2500.00",
				GroupingCategory:   "GM", // Income
				GroupingCode:       "111",
				TaxonomyCode:       "311", // TaxonomyCode is integer string
			},
		},
	}
	// accountIDs map for GroupingCode validation
	accountIDs := map[string]bool{"111": true, "211": true, "711": true}

	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid GeneralLedgerAccounts, got %d errors: %v", len(errs), errs)
	}
}

func TestValidateGeneralLedgerAccounts_Invalid_EmptyTaxonomyReference(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "", // Invalid
		Account: []saft.Account{
			{
				AccountID: "111", AccountDescription: "Cash", GroupingCategory: "AR",
			},
		},
	}
	accountIDs := map[string]bool{"111": true}
	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) == 0 {
		t.Errorf("Expected error for empty TaxonomyReference, got no errors")
	} else {
		found := false
		for _, err := range errs {
			if err.Error() == "TaxonomyReference is required" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error message 'TaxonomyReference is required', got: %v", errs)
		}
	}
}

func TestValidateGeneralLedgerAccounts_Invalid_Account_EmptyID(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "SNC-BASE",
		Account: []saft.Account{
			{
				AccountID: "", AccountDescription: "Cash", GroupingCategory: "AR", // Invalid AccountID
			},
		},
	}
	accountIDs := map[string]bool{}
	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) == 0 {
		t.Errorf("Expected error for empty AccountID, got no errors")
	} else if !containsErrorMsg(errs, "AccountID is required") {
		t.Errorf("Expected error message 'AccountID is required', got: %v", errs)
	}
}

func TestValidateGeneralLedgerAccounts_Invalid_Account_EmptyDescription(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "SNC-BASE",
		Account: []saft.Account{
			{
				AccountID: "111", AccountDescription: "", GroupingCategory: "AR", // Invalid AccountDescription
			},
		},
	}
	accountIDs := map[string]bool{"111": true}
	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) == 0 {
		t.Errorf("Expected error for empty AccountDescription, got no errors")
	} else if !containsErrorMsg(errs, "AccountDescription is required") {
		t.Errorf("Expected error message 'AccountDescription is required', got: %v", errs)
	}
}

func TestValidateGeneralLedgerAccounts_Invalid_Account_InvalidGroupingCategory(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "SNC-BASE",
		Account: []saft.Account{
			{
				AccountID: "111", AccountDescription: "Cash", GroupingCategory: "XX", // Invalid GroupingCategory
			},
		},
	}
	accountIDs := map[string]bool{"111": true}
	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) == 0 {
		t.Errorf("Expected error for invalid GroupingCategory, got no errors")
	} else if !containsErrorMsgPrefix(errs, "invalid GroupingCategory: XX for AccountID: 111") {
		t.Errorf("Expected error message starting with 'invalid GroupingCategory: XX for AccountID: 111', got: %v", errs)
	}
}

func TestValidateGeneralLedgerAccounts_Invalid_Account_InvalidGroupingCode(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "SNC-BASE",
		Account: []saft.Account{
			{
				AccountID: "211", AccountDescription: "Debtors", GroupingCategory: "GA", GroupingCode: "999", // Invalid GroupingCode
			},
		},
	}
	accountIDs := map[string]bool{"211": true} // 999 is not in this map
	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) == 0 {
		t.Errorf("Expected error for invalid GroupingCode, got no errors")
	} else if !containsErrorMsgPrefix(errs, "invalid GroupingCode: 999 for AccountID: 211") {
		t.Errorf("Expected error message starting with 'invalid GroupingCode: 999 for AccountID: 211', got: %v", errs)
	}
}

func TestValidateGeneralLedgerAccounts_Invalid_Account_InvalidTaxonomyCode(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "SNC-BASE",
		Account: []saft.Account{
			{
				AccountID: "711", AccountDescription: "Sales", GroupingCategory: "GM", GroupingCode: "111", TaxonomyCode: "ABC", // Invalid TaxonomyCode
			},
		},
	}
	accountIDs := map[string]bool{"111":true, "711": true}
	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) == 0 {
		t.Errorf("Expected error for invalid TaxonomyCode, got no errors")
	} else if !containsErrorMsgPrefix(errs, "invalid TaxonomyCode: ABC for AccountID: 711") {
		t.Errorf("Expected error message starting with 'invalid TaxonomyCode: ABC for AccountID: 711', got: %v", errs)
	}
}


// Helper function to check if a specific error message is in the error slice
func containsErrorMsg(errs []error, msg string) bool {
	for _, err := range errs {
		if err.Error() == msg {
			return true
		}
	}
	return false
}

// Helper function to check if an error message with a specific prefix is in the error slice
func containsErrorMsgPrefix(errs []error, prefix string) bool {
	for _, err := range errs {
		if len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// Add more tests for nil gla, nil gla.Account, etc.
func TestValidateGeneralLedgerAccounts_NilInput(t *testing.T) {
	accountIDs := map[string]bool{}
	errs := masterfiles.ValidateGeneralLedgerAccounts(nil, accountIDs)
	if len(errs) == 0 {
		t.Errorf("Expected error for nil GeneralLedgerAccounts, got no errors")
	} else if !containsErrorMsg(errs, "GeneralLedgerAccounts is nil") {
		t.Errorf("Expected 'GeneralLedgerAccounts is nil', got: %v", errs)
	}
}

func TestValidateGeneralLedgerAccounts_EmptyAccountList(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "SNC-BASE",
		Account:           []saft.Account{}, // Empty list
	}
	accountIDs := map[string]bool{}
	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) > 0 { // Empty account list is not an error itself, only TaxonomyReference is checked at this level
		t.Errorf("Expected no errors for empty account list (but valid TaxonomyRef), got: %v", errs)
	}
}

func TestValidateGeneralLedgerAccounts_Valid_OptionalGroupingCode(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "SNC-BASE",
		Account: []saft.Account{
			{
				AccountID:          "111",
				AccountDescription: "Cash",
				GroupingCategory:   "AR",
				GroupingCode:       "", // Optional, valid if empty
			},
		},
	}
	accountIDs := map[string]bool{"111": true}
	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for optional empty GroupingCode, got: %v", errs)
	}
}

func TestValidateGeneralLedgerAccounts_Valid_OptionalTaxonomyCode(t *testing.T) {
	gla := &saft.GeneralLedgerAccounts{
		TaxonomyReference: "SNC-BASE",
		Account: []saft.Account{
			{
				AccountID:          "111",
				AccountDescription: "Cash",
				GroupingCategory:   "AR",
				TaxonomyCode:       "", // Optional, valid if empty
			},
		},
	}
	accountIDs := map[string]bool{"111": true}
	errs := masterfiles.ValidateGeneralLedgerAccounts(gla, accountIDs)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for optional empty TaxonomyCode, got: %v", errs)
	}
}

// Example test for GroupingCategory "GM" requiring TaxonomyCode (if this was a rule, it's not in current GLA validator)
// The current validator only checks TaxonomyCode is an int if present, not if it's required for GM.
// The original saft/validation.go had a check: (account.GroupingCategory == "GM" && account.TaxonomyCode == nil)
// This specific check is NOT in the new masterfiles.ValidateGeneralLedgerAccounts.
// If it should be, the validator needs update. For now, testing existing behavior.

// Note: The current masterfiles.ValidateGeneralLedgerAccounts does not enforce that GroupingCategory="GM" *must* have a TaxonomyCode.
// It only validates that if TaxonomyCode *is* present, it's an integer.
// And that GroupingCode, if present, is a valid AccountID.
// The original `saft/validation.go` had more specific inter-field rules for GroupingCategory/TaxonomyCode/GroupingCode
// which are not all present in the new `masterfiles.ValidateGeneralLedgerAccounts`.
// This test suite is for the *newly created* validator in `masterfiles`.
// If parity with old checks is desired, the new validator needs to be updated.
// For now, these tests correctly reflect the functionality of `masterfiles.ValidateGeneralLedgerAccounts`.

// TODO: Add tests for other masterfile components (Customer, Supplier, Product, TaxTable)
// TODO: Add tests for GeneralLedgerEntries
// TODO: Add tests for SourceDocuments (SalesInvoices, MovementOfGoods, WorkingDocuments, Payments)
// TODO: Update saft/test/test.go if it's a main test runner (not standard, go test ./... is typical)
// For this exercise, I will only create this one test file as a demonstration.
// The other test files would follow a similar pattern:
// - New file in saft/test/, e.g., customer_validation_test.go
// - package test
// - import saft types and the specific validator package/functions
// - Test functions for valid cases and specific invalid cases for each rule.
// - Helper functions like containsErrorMsg can be shared or defined per file.
// The main `saft/test/test.go` is not standard for running tests but could be used to define
// shared helper types or a TestMain if complex setup/teardown for the whole test suite is needed.
// Standard Go test execution will find all `*_test.go` files.
This is a great start for testing `ValidateGeneralLedgerAccounts`. You've covered valid cases and specific invalid conditions for each rule in that validator. The helper functions `containsErrorMsg` and `containsErrorMsgPrefix` are good for checking errors.

**Next Steps & Structure:**

Given the constraint of not being able to write all tests in one go, here's how I would structure the continuation and what I'd tell you to do next if this were an interactive session where I'm the one generating the code:

1.  **Confirm Test File Location Strategy**: The current approach is creating `saft/test/generalledgeraccounts_validation_test.go`. This means the test package is `package test` (or `package saft_test` if you prefer to avoid conflict with the main `saft` package, though `package test` is fine if it's a common test-only package). This external test package will import the necessary production code packages. This aligns with the prompt's request for tests in `saft/test`.

2.  **Outline for Other Test Files**: I would create the following files in `saft/test/`, each focusing on a specific component or group of components:
    *   `customer_validation_test.go`
    *   `supplier_validation_test.go`
    *   `product_validation_test.go`
    *   `taxtable_validation_test.go`
    *   `generalledgerentries_validation_test.go`
    *   `salesinvoices_validation_test.go`
    *   `movementofgoods_validation_test.go`
    *   `workingdocuments_validation_test.go`
    *   `payments_validation_test.go`

3.  **Contents of Each Test File**: Each file would follow the pattern you've established:
    *   `package test`
    *   Imports: `"testing"`, the main `saft` types package (e.g., `"github.com/hestiatechnology/autoridadetributaria/saft"`), and the specific validation package being tested (e.g., `"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/masterfiles"` or `.../generalledgerentries`, etc.).
    *   Test functions:
        *   `TestValidate[ComponentName]_Valid(t *testing.T)`: With well-structured, valid SAFT data for that component.
        *   Multiple `TestValidate[ComponentName]_Invalid_[SpecificRule](t *testing.T)` functions, each targeting one validation rule. For example, for Customer validation: `TestValidateCustomer_Invalid_EmptyCustomerID`, `TestValidateCustomer_Invalid_AccountIDNotFound`, `TestValidateCustomer_Invalid_NIFFormat`, `TestValidateCustomer_Invalid_BillingAddress_EmptyCity`, etc.
    *   Helper functions (like `containsErrorMsg`) can be defined in each file or, preferably, in a shared `test_helpers_test.go` (still `package test`) within `saft/test/` to avoid duplication.

4.  **`saft/test/test.go` Update**:
    *   The prompt mentions updating `saft/test/test.go` to include calls to new test cases. This is not the standard way Go's `testing` package works. `go test` automatically discovers test functions (named `TestXxx`) in `*_test.go` files.
    *   If `saft/test/test.go` is intended to be a custom test suite runner (e.g., using `testing.Main` or orchestrating tests in a specific order), then one would add calls to functions that internally run these `testing.T` tests, or register them.
    *   However, if it's just a placeholder, no action might be needed other than creating the `*_test.go` files.
    *   **Assumption**: I will assume standard Go testing. The `*_test.go` files in `saft/test/` will be sufficient for `go test ./test/...` to pick them up. No explicit calls need to be added to a central `saft/test/test.go` unless it serves a special purpose like `TestMain`.

5.  **Data for Tests**:
    *   For each test, minimal valid or targeted invalid SAFT structs would be created.
    *   Maps for dependent IDs (like `accountIDs`, `customerIDs`, `productCodes`, `validInvoiceReferences`) would be mocked up as needed for the specific test case. For example, when testing `ValidateCustomer` for an invalid `AccountID`, the `accountIDs` map passed to it would intentionally not contain the `AccountID` used in the test customer data.

**Example outline for `customer_validation_test.go`**:

```go
// saft/test/customer_validation_test.go
package test

import (
	"testing"
	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/masterfiles" // Or wherever ValidateCustomer / ValidateCustomersInAuditfile is
)

// Mock account IDs for testing customer AccountID validation
var testAccountIDs = map[string]bool{"ACC1": true, "ACC2": true}

func TestValidateCustomer_Valid(t *testing.T) {
	// customer := saft.Customer{... valid data ...}
	// errs := masterfiles.ValidateCustomer(&customer, testAccountIDs)
	// if len(errs) > 0 { t.Errorf(...) }
}

func TestValidateCustomer_Invalid_EmptyCustomerID(t *testing.T) {
	// customer := saft.Customer{CustomerID: "", ... other fields ...}
	// errs := masterfiles.ValidateCustomer(&customer, testAccountIDs)
	// if !containsErrorMsg(errs, "CustomerID is required") {t.Errorf(...)}
}

// ... other tests for Customer:
// TestValidateCustomer_Invalid_EmptyAccountID
// TestValidateCustomer_Invalid_AccountIDNotFound
// TestValidateCustomer_Invalid_CustomerTaxID_Empty
// TestValidateCustomer_Invalid_CustomerTaxID_Format (not 9 digits)
// TestValidateCustomer_Invalid_CustomerTaxID_NIFCheck (fails common.ValidateNIFPT)
// TestValidateCustomer_Invalid_EmptyCompanyName
// TestValidateCustomer_Invalid_BillingAddress_EmptyAddressDetail
// TestValidateCustomer_Invalid_BillingAddress_EmptyCity
// TestValidateCustomer_Invalid_BillingAddress_EmptyPostalCode
// TestValidateCustomer_Invalid_BillingAddress_EmptyCountry
// TestValidateCustomer_Invalid_BillingAddress_InvalidCountryFormat
// TestValidateCustomer_Invalid_ShipToAddress_EmptyAddressDetail (for one of the ship-to addresses)
// TestValidateCustomer_Invalid_SelfBillingIndicator (e.g., value 2 if it should be 0 or 1)

// If testing ValidateCustomersInAuditfile:
func TestValidateCustomersInAuditfile_MultipleValid(t *testing.T) {
	// auditFile := &saft.AuditFile{ MasterFiles: &saft.MasterFiles{ Customer: []saft.Customer{ validCust1, validCust2}}}
	// errs := masterfiles.ValidateCustomersInAuditfile(auditFile, testAccountIDs)
	// if len(errs) > 0 {t.Errorf(...)}
}

func TestValidateCustomersInAuditfile_OneInvalid(t *testing.T) {
	// auditFile := &saft.AuditFile{ MasterFiles: &saft.MasterFiles{ Customer: []saft.Customer{ validCust1, invalidCust}}}
	// errs := masterfiles.ValidateCustomersInAuditfile(auditFile, testAccountIDs)
	// if len(errs) == 0 {t.Errorf(...)}
	// Check for expected error from invalidCust
}

// (containsErrorMsg and containsErrorMsgPrefix helpers would be here or in a shared test_helpers_test.go)
```

**Path Forward with the Tool:**

Since I cannot create all these files and their detailed content in one response due to output limits and complexity, I would proceed by:
1.  Asking you to **create one new test file at a time**, e.g., `saft/test/customer_validation_test.go`.
2.  I would provide the content for that file, including a few representative test cases (valid, and a selection of invalid cases).
3.  You would use `create_file_with_block` or `overwrite_file_with_block`.
4.  We would repeat this for each major validation component.

Given the current tool interaction, I should provide the content for the next logical test file. After `generalledgeraccounts_validation_test.go`, `customer_validation_test.go` is a good candidate.

Let me provide the content for `saft/test/customer_validation_test.go`.
