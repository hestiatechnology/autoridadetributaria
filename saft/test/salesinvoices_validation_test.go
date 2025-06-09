package test

import (
	"testing"
	"time" // For time.Parse in test setup

	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/sourcedocuments" // Assuming this is package path
)

// --- Helper functions (can be shared, defined per file for this exercise) ---
func containsErrorMsgSI(errs []error, msg string) bool {
	for _, err := range errs {
		if err.Error() == msg {
			return true
		}
	}
	return false
}

func containsErrorMsgPrefixSI(errs []error, prefix string) bool {
	for _, err := range errs {
		if len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// --- Mock data for dependencies ---
var (
	mockProductCodesSI = map[string]bool{"PROD001": true, "PROD002": true}
	mockCustomerIDsSI  = map[string]bool{"CUST001": true, "CUST002": true}
	// mockTaxTableEntriesSI (if deeper tax validation was implemented for TaxCode validity)
)

func sampleValidSalesInvoice(invoiceNo string, custID string, prodCode string) saft.Invoice {
	return saft.Invoice{
		InvoiceNo:   invoiceNo,
		ATCUD:       "ATCUDVALID123",
		DocumentStatus: &saft.DocumentStatus{ // Assuming DocumentStatus is a pointer
			InvoiceStatus:     "N", // Normal
			InvoiceStatusDate: "2023-10-30T10:00:00",
			SourceID:          "UserX",
			SourceBilling:     "P", // Production
		},
		Hash:          "HASH1234567890ABCDEF=",
		HashControl:   "1",
		Period:        "10",
		InvoiceDate:   "2023-10-30",
		InvoiceType:   "FT", // Invoice
		SpecialRegimes: &saft.SpecialRegimes{ // Assuming SpecialRegimes is a pointer
			SelfBillingIndicator:         "0",
			CashVATSchemeIndicator:       0, // Integer
			ThirdPartiesBillingIndicator: "0",
		},
		SourceID:        "UserX", // Invoice level SourceID
		SystemEntryDate: "2023-10-30T09:00:00",
		CustomerID:      custID,
		Line: []saft.Line{
			{
				LineNumber:         "1",
				ProductCode:        prodCode,
				ProductDescription: "Test Product",
				Quantity:           "2.00",
				UnitOfMeasure:      "UN",
				UnitPrice:          "50.00",
				TaxPointDate:       "2023-10-30",
				Description:        "Line item description",
				Tax: &saft.Tax{ // Assuming Tax is a pointer
					TaxType:          "IVA",
					TaxCountryRegion: "PT",
					TaxCode:          "NOR",
					TaxPercentage:    "23.00",
				},
				// DebitAmount/CreditAmount not typical for SalesInvoice Line, NetTotal derived.
			},
		},
		DocumentTotals: &saft.DocumentTotals{ // Assuming DocumentTotals is a pointer
			TaxPayable: "23.00",  // 2 * 50 * 0.23
			NetTotal:   "100.00", // 2 * 50
			GrossTotal: "123.00", // 100 + 23
			Currency:   nil,      // Assuming base currency EUR, so Currency block not needed
		},
	}
}

func TestValidateSalesInvoices_Valid(t *testing.T) {
	si := &saft.SalesInvoices{
		NumberOfEntries: "1",
		TotalDebit:      "123.00",
		TotalCredit:     "123.00",
		Invoice:         []saft.Invoice{sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")},
	}

	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid SalesInvoices, got %d: %v", len(errs), errs)
	}
}

func TestValidateSalesInvoices_Invalid_NumberOfEntriesMismatch(t *testing.T) {
	si := &saft.SalesInvoices{
		NumberOfEntries: "2", // Mismatch
		TotalDebit:      "123.00",
		TotalCredit:     "123.00",
		Invoice:         []saft.Invoice{sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	if !containsErrorMsgPrefixSI(errs, "mismatch in NumberOfEntries: declared 2, actual 1") {
		t.Errorf("Expected NumberOfEntries mismatch error, got: %v", errs)
	}
}

func TestValidateSalesInvoices_Invalid_TotalCreditMismatch(t *testing.T) {
	si := &saft.SalesInvoices{
		NumberOfEntries: "1",
		TotalDebit:      "123.00",
		TotalCredit:     "100.00", // Mismatch, actual is 123.00
		Invoice:         []saft.Invoice{sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	// Note: float comparison formatting might vary slightly. Using prefix.
	if !containsErrorMsgPrefixSI(errs, "mismatch in TotalCredit: declared 100.000000, calculated 123.000000") {
		t.Errorf("Expected TotalCredit mismatch error, got: %v", errs)
	}
}

func TestValidateSalesInvoices_Invoice_Invalid_EmptyInvoiceNo(t *testing.T) {
	invoice := sampleValidSalesInvoice("", "CUST001", "PROD001") // Empty InvoiceNo
	si := &saft.SalesInvoices{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Invoice: []saft.Invoice{invoice},
	}
	// Adjust totals because GrossTotal might be parsed as 0 if DocumentTotals is not fully formed due to missing InvoiceNo for identifier
	// For this test, assume DocumentTotals are still calculated based on sample data
	si.TotalDebit = "123.00"
	si.TotalCredit = "123.00"

	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	if !containsErrorMsgPrefixSI(errs, "Invoice[0 No:]: InvoiceNo is required") {
		t.Errorf("Expected InvoiceNo required error, got: %v", errs)
	}
}

func TestValidateSalesInvoices_Invoice_Invalid_DocumentStatus_InvalidStatus(t *testing.T) {
	invoice := sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")
	invoice.DocumentStatus.InvoiceStatus = "X" // Invalid status
	si := &saft.SalesInvoices{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Invoice: []saft.Invoice{invoice},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	expectedMsg := "Invoice[0 No:FT/001]: DocumentStatus.InvoiceStatus 'X' is invalid"
	if !containsErrorMsgSI(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSalesInvoices_Invoice_Invalid_InvoiceType(t *testing.T) {
	invoice := sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")
	invoice.InvoiceType = "XX" // Invalid type
	si := &saft.SalesInvoices{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Invoice: []saft.Invoice{invoice},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	expectedMsg := "Invoice[0 No:FT/001]: InvoiceType 'XX' is invalid"
	if !containsErrorMsgSI(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSalesInvoices_Invoice_Invalid_SpecialRegimes_SelfBilling(t *testing.T) {
	invoice := sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")
	invoice.SpecialRegimes.SelfBillingIndicator = "2" // Invalid
	si := &saft.SalesInvoices{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Invoice: []saft.Invoice{invoice},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	expectedMsg := "Invoice[0 No:FT/001]: SpecialRegimes.SelfBillingIndicator '2' is invalid (must be \"0\" or \"1\")"
	if !containsErrorMsgSI(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSalesInvoices_Invoice_Invalid_CustomerID_NotFound(t *testing.T) {
	invoice := sampleValidSalesInvoice("FT/001", "CUST_INVALID", "PROD001") // Invalid CustomerID
	si := &saft.SalesInvoices{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Invoice: []saft.Invoice{invoice},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	expectedMsg := "Invoice[0 No:FT/001]: CustomerID 'CUST_INVALID' is invalid or not found"
	if !containsErrorMsgSI(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSalesInvoices_Line_Invalid_ProductCode_NotFound(t *testing.T) {
	invoice := sampleValidSalesInvoice("FT/001", "CUST001", "PROD_INVALID") // Invalid ProductCode
	si := &saft.SalesInvoices{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Invoice: []saft.Invoice{invoice},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	expectedMsg := "Invoice[0 No:FT/001] Line[1]: ProductCode 'PROD_INVALID' is invalid or not found"
	if !containsErrorMsgSI(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}


func TestValidateSalesInvoices_Line_Invalid_Tax_EmptyTaxCode(t *testing.T) {
	invoice := sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")
	invoice.Line[0].Tax.TaxCode = "" // Empty TaxCode
	si := &saft.SalesInvoices{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Invoice: []saft.Invoice{invoice},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	expectedMsg := "Invoice[0 No:FT/001] Line[1]: Tax.TaxCode is required"
	if !containsErrorMsgSI(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateSalesInvoices_Line_Invalid_Tax_IVAMissingPercentageAndAmount(t *testing.T) {
	invoice := sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")
	invoice.Line[0].Tax.TaxType = "IVA"
	invoice.Line[0].Tax.TaxPercentage = "" // Missing
	invoice.Line[0].Tax.TaxAmount = ""     // Missing
	si := &saft.SalesInvoices{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Invoice: []saft.Invoice{invoice},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	expectedMsg := "Invoice[0 No:FT/001] Line[1]: For TaxType IVA, either TaxPercentage or TaxAmount must be provided"
	if !containsErrorMsgSI(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}


func TestValidateSalesInvoices_DocumentTotals_Invalid_NetTotalFormat(t *testing.T) {
	invoice := sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")
	invoice.DocumentTotals.NetTotal = "INVALID_FORMAT"
	si := &saft.SalesInvoices{
		NumberOfEntries: "1",
		// Totals need to be recalculated or tests will fail on header totals before line totals.
		// For simplicity, setting them to what would be expected if this invalid format was 0.
		TotalDebit:      "23.00", // GrossTotal would be TaxPayable if NetTotal is 0
		TotalCredit:     "23.00",
		Invoice:         []saft.Invoice{invoice},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	expectedMsgPrefix := "Invoice[0 No:FT/001]: DocumentTotals.NetTotal 'INVALID_FORMAT' is not a valid decimal"
	if !containsErrorMsgPrefixSI(errs, expectedMsgPrefix) {
		t.Errorf("Expected error starting with '%s', got: %v", expectedMsgPrefix, errs)
	}
}

func TestValidateSalesInvoices_DocumentTotals_Invalid_CurrencyCode(t *testing.T) {
	invoice := sampleValidSalesInvoice("FT/001", "CUST001", "PROD001")
	invoice.DocumentTotals.Currency = &saft.Currency{ // Assuming Currency is a pointer
		CurrencyCode:   "USDD", // Invalid, > 3 chars
		CurrencyAmount: "150.00",
	}
	si := &saft.SalesInvoices{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Invoice: []saft.Invoice{invoice},
	}
	errs := sourcedocuments.ValidateSalesInvoices(si, mockProductCodesSI, mockCustomerIDsSI)
	expectedMsg := "Invoice[0 No:FT/001]: DocumentTotals.Currency.CurrencyCode 'USDD' is invalid"
	if !containsErrorMsgSI(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

// Further tests could include:
// - Nil SalesInvoices input.
// - More permutations of DocumentStatus, SpecialRegimes.
// - More Line validation: empty fields (ProductDescription, Quantity, UnitOfMeasure, UnitPrice, TaxPointDate, Description).
// - Invalid formats for Quantity, UnitPrice.
// - More Tax validation: invalid TaxType, TaxCountryRegion format.
// - DocumentTotals: invalid TaxPayable, GrossTotal formats. Invalid CurrencyAmount.
// - Multiple invoices, some valid, some invalid.
// - Multiple lines, some valid, some invalid.
// - Cases where DocumentStatus, SpecialRegimes, Tax, DocumentTotals, Currency are nil pointers.
//   The sampleValidSalesInvoice already assumes these are pointers and provides valid structs.
//   A test where, e.g. invoice.DocumentStatus = nil, should trigger "DocumentStatus is required".
