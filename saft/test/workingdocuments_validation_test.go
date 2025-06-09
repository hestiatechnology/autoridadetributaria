package test

import (
	"testing"
	// "time" // Not strictly needed for these test cases unless parsing dates for setup

	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/sourcedocuments"
)

// --- Helper functions (can be shared) ---
func containsErrorMsgWD(errs []error, msg string) bool {
	for _, err := range errs {
		if err.Error() == msg {
			return true
		}
	}
	return false
}

func containsErrorMsgPrefixWD(errs []error, prefix string) bool {
	for _, err := range errs {
		if len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// --- Mock data for dependencies (reused conceptually) ---
var (
	mockProductCodesWD = map[string]bool{"PRODWD001": true, "PRODSALE001": true}
	mockCustomerIDsWD  = map[string]bool{"CUSTWD001": true, "CUSTSALE001": true}
)

func sampleValidWorkDocument(docNum string, custID string, prodCode string, workType string) saft.WorkDocument {
	return saft.WorkDocument{
		DocumentNumber: docNum,
		ATCUD:          "ATCUDWORK123",
		DocumentStatus: &saft.WorkDocumentStatus{ // Note: WorkDocumentStatus struct
			WorkStatus:     "N", // Normal
			WorkStatusDate: "2023-11-05T14:00:00",
			SourceID:       "UserWork",
			SourceBilling:  "P",
		},
		Hash:          "WORKHASH1234567890=",
		HashControl:   "1",
		Period:        "11",
		WorkDate:      "2023-11-05",
		WorkType:      workType, // e.g., "OR" for Orcamento
		SourceID:      "UserWork_DocLevel",
		SystemEntryDate: "2023-11-05T13:00:00",
		CustomerID:    custID,
		Line: []saft.WorkLine{ // Note: WorkLine struct
			{
				LineNumber:         "1",
				ProductCode:        prodCode,
				ProductDescription: "Product for Work Document",
				Quantity:           "1.00",
				UnitOfMeasure:      "EA",
				UnitPrice:          "200.00",
				TaxPointDate:       "2023-11-05",
				Description:        "Work document line item",
				Tax: &saft.Tax{ // Assuming Tax struct is reused
					TaxType:          "IVA",
					TaxCountryRegion: "PT",
					TaxCode:          "NOR",
					TaxPercentage:    "23.00",
				},
			},
		},
		DocumentTotals: &saft.WorkDocumentTotals{ // Note: WorkDocumentTotals struct
			TaxPayable: "46.00",  // 1 * 200 * 0.23
			NetTotal:   "200.00", // 1 * 200
			GrossTotal: "246.00", // 200 + 46
		},
	}
}

func TestValidateWorkingDocuments_Valid(t *testing.T) {
	wd := &saft.WorkingDocuments{
		NumberOfEntries: "1",
		TotalDebit:      "246.00", // Sum of GrossTotals
		TotalCredit:     "246.00", // Sum of GrossTotals
		WorkDocument:    []saft.WorkDocument{sampleValidWorkDocument("OR/001", "CUSTWD001", "PRODWD001", "OR")},
	}
	errs := sourcedocuments.ValidateWorkingDocuments(wd, mockProductCodesWD, mockCustomerIDsWD)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid WorkingDocuments, got %d: %v", len(errs), errs)
	}
}

func TestValidateWorkingDocuments_Invalid_NumberOfEntriesMismatch(t *testing.T) {
	wd := &saft.WorkingDocuments{
		NumberOfEntries: "2", // Mismatch
		TotalDebit:      "246.00", TotalCredit: "246.00",
		WorkDocument: []saft.WorkDocument{sampleValidWorkDocument("OR/001", "CUSTWD001", "PRODWD001", "OR")},
	}
	errs := sourcedocuments.ValidateWorkingDocuments(wd, mockProductCodesWD, mockCustomerIDsWD)
	if !containsErrorMsgPrefixWD(errs, "mismatch in NumberOfEntries: declared 2, actual 1") {
		t.Errorf("Expected NumberOfEntries mismatch error, got: %v", errs)
	}
}

func TestValidateWorkingDocuments_Invalid_TotalDebitMismatch(t *testing.T) {
	wd := &saft.WorkingDocuments{
		NumberOfEntries: "1",
		TotalDebit:      "200.00", // Mismatch, actual is 246.00
		TotalCredit:     "246.00",
		WorkDocument:    []saft.WorkDocument{sampleValidWorkDocument("OR/001", "CUSTWD001", "PRODWD001", "OR")},
	}
	errs := sourcedocuments.ValidateWorkingDocuments(wd, mockProductCodesWD, mockCustomerIDsWD)
	if !containsErrorMsgPrefixWD(errs, "mismatch in TotalDebit: declared 200.000000, calculated 246.000000") {
		t.Errorf("Expected TotalDebit mismatch error, got: %v", errs)
	}
}

func TestValidateWorkingDocuments_WorkDoc_Invalid_EmptyDocNumber(t *testing.T) {
	workDoc := sampleValidWorkDocument("", "CUSTWD001", "PRODWD001", "OR") // Empty DocumentNumber
	wd := &saft.WorkingDocuments{
		NumberOfEntries: "1", TotalDebit: "246.00", TotalCredit: "246.00",
		WorkDocument: []saft.WorkDocument{workDoc},
	}
	errs := sourcedocuments.ValidateWorkingDocuments(wd, mockProductCodesWD, mockCustomerIDsWD)
	if !containsErrorMsgPrefixWD(errs, "WorkDocument[0 No:]: DocumentNumber is required") {
		t.Errorf("Expected DocumentNumber required error, got: %v", errs)
	}
}

func TestValidateWorkingDocuments_WorkDoc_Invalid_WorkStatus(t *testing.T) {
	workDoc := sampleValidWorkDocument("OR/001", "CUSTWD001", "PRODWD001", "OR")
	workDoc.DocumentStatus.WorkStatus = "X" // Invalid status
	wd := &saft.WorkingDocuments{
		NumberOfEntries: "1", TotalDebit: "246.00", TotalCredit: "246.00",
		WorkDocument: []saft.WorkDocument{workDoc},
	}
	errs := sourcedocuments.ValidateWorkingDocuments(wd, mockProductCodesWD, mockCustomerIDsWD)
	expectedMsg := "WorkDocument[0 No:OR/001]: DocumentStatus.WorkStatus 'X' is invalid"
	if !containsErrorMsgWD(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateWorkingDocuments_WorkDoc_Invalid_WorkType(t *testing.T) {
	workDoc := sampleValidWorkDocument("OR/001", "CUSTWD001", "PRODWD001", "XX") // Invalid WorkType
	wd := &saft.WorkingDocuments{
		NumberOfEntries: "1", TotalDebit: "246.00", TotalCredit: "246.00",
		WorkDocument: []saft.WorkDocument{workDoc},
	}
	errs := sourcedocuments.ValidateWorkingDocuments(wd, mockProductCodesWD, mockCustomerIDsWD)
	expectedMsg := "WorkDocument[0 No:OR/001]: WorkType 'XX' is invalid"
	if !containsErrorMsgWD(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateWorkingDocuments_WorkDoc_Invalid_CustomerID_NotFound(t *testing.T) {
	workDoc := sampleValidWorkDocument("OR/001", "CUST_INVALID", "PRODWD001", "OR") // Invalid CustomerID
	wd := &saft.WorkingDocuments{
		NumberOfEntries: "1", TotalDebit: "246.00", TotalCredit: "246.00",
		WorkDocument: []saft.WorkDocument{workDoc},
	}
	errs := sourcedocuments.ValidateWorkingDocuments(wd, mockProductCodesWD, mockCustomerIDsWD)
	expectedMsg := "WorkDocument[0 No:OR/001]: CustomerID 'CUST_INVALID' is invalid or not found"
	if !containsErrorMsgWD(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateWorkingDocuments_Line_Invalid_ProductCode_NotFound(t *testing.T) {
	workDoc := sampleValidWorkDocument("OR/001", "CUSTWD001", "PROD_INVALID", "OR") // Invalid ProductCode
	wd := &saft.WorkingDocuments{
		NumberOfEntries: "1", TotalDebit: "246.00", TotalCredit: "246.00",
		WorkDocument: []saft.WorkDocument{workDoc},
	}
	errs := sourcedocuments.ValidateWorkingDocuments(wd, mockProductCodesWD, mockCustomerIDsWD)
	expectedMsg := "WorkDocument[0 No:OR/001] Line[1]: ProductCode 'PROD_INVALID' is invalid or not found"
	if !containsErrorMsgWD(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateWorkingDocuments_Line_Invalid_Tax_IVAMissingPercentageAndAmount(t *testing.T) {
	workDoc := sampleValidWorkDocument("OR/001", "CUSTWD001", "PRODWD001", "OR")
	if workDoc.Line[0].Tax != nil {
		workDoc.Line[0].Tax.TaxType = "IVA"
		workDoc.Line[0].Tax.TaxPercentage = ""
		workDoc.Line[0].Tax.TaxAmount = ""
	}
	wd := &saft.WorkingDocuments{
		NumberOfEntries: "1", TotalDebit: "246.00", TotalCredit: "246.00",
		WorkDocument: []saft.WorkDocument{workDoc},
	}
	// Adjust DocumentTotals as tax might change
	workDoc.DocumentTotals.TaxPayable = "0.00"
	workDoc.DocumentTotals.GrossTotal = workDoc.DocumentTotals.NetTotal

	errs := sourcedocuments.ValidateWorkingDocuments(wd, mockProductCodesWD, mockCustomerIDsWD)
	expectedMsg := "WorkDocument[0 No:OR/001] Line[1]: For TaxType IVA, either TaxPercentage or TaxAmount must be provided"
	if !containsErrorMsgWD(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

// Further tests could include:
// - Nil WorkingDocuments input.
// - More DocumentStatus checks (SourceID, WorkStatusDate, SourceBilling).
// - Empty/invalid Hash, HashControl, Period, WorkDate, SourceID (doc level), SystemEntryDate.
// - Line validation: empty/invalid ProductDescription, Quantity, UnitOfMeasure, UnitPrice, TaxPointDate, Description.
// - Tax validation: invalid TaxType, TaxCountryRegion, TaxCode.
// - DocumentTotals validation: invalid TaxPayable, NetTotal, GrossTotal, Currency fields.
// - Multiple WorkDocument entries and multiple lines with mixed validity.
// - Nil pointers for DocumentStatus, Tax, DocumentTotals, Currency.
