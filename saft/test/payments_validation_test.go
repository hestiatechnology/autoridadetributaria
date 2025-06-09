package test

import (
	"testing"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/sourcedocuments"
)

// --- Helper functions (can be shared) ---
func containsErrorMsgPay(errs []error, msg string) bool {
	for _, err := range errs {
		if err.Error() == msg {
			return true
		}
	}
	return false
}

func containsErrorMsgPrefixPay(errs []error, prefix string) bool {
	for _, err := range errs {
		if len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// --- Mock data for dependencies ---
var (
	mockCustomerIDsPay = map[string]bool{"CUSTPAY001": true}
	// For SourceDocumentID validation: map[OriginatingON]InvoiceDate
	mockValidInvoiceRefsPay = map[string]time.Time{
		"FT XYZ/1": time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC),
		"FT ABC/2": time.Date(2023, 10, 20, 0, 0, 0, 0, time.UTC),
	}
)

func sampleValidPayment(refNo string, custID string, origON string, invDateStr string) saft.Payment {
	return saft.Payment{
		PaymentRefNo:  refNo,
		ATCUD:         "ATCUDPAY123",
		Period:        "11", // Optional, but can be present
		TransactionID: "TRANSIDPAY001", // Optional
		TransactionDate: "2023-11-10",
		PaymentType:   "RC", // Receipt
		Description:   "Payment for services", // Optional
		SystemID:      "SYS001", // Optional
		DocumentStatus: &saft.PaymentStatus{ // Note: PaymentStatus struct
			PaymentStatus:     "N", // Normal
			PaymentStatusDate: "2023-11-10T10:00:00",
			SourceID:          "UserPay",
			SourcePayment:     "P",
		},
		PaymentMethod: []saft.PaymentMethod{
			{PaymentMechanism: "TB", PaymentAmount: "123.00", PaymentDate: "2023-11-10"},
		},
		SourceID:        "UserPay_DocLevel",
		SystemEntryDate: "2023-11-10T09:50:00",
		CustomerID:    custID,
		Line: []saft.PaymentLine{ // Note: PaymentLine struct
			{
				LineNumber: "1",
				SourceDocumentID: []saft.SourceDocumentIDType{ // Note: Slice of SourceDocumentIDType
					{OriginatingON: origON, InvoiceDate: invDateStr},
				},
				// SettlementAmount optional
				CreditAmount: "123.00", // Payment received credits customer account / AR
			},
		},
		DocumentTotals: &saft.PaymentTotals{ // Note: PaymentTotals struct
			TaxPayable: "0.00",   // Usually no new tax on simple payment
			NetTotal:   "123.00",
			GrossTotal: "123.00",
			// Settlement and Currency are optional
		},
	}
}

func TestValidatePayments_Valid(t *testing.T) {
	invoiceDate := mockValidInvoiceRefsPay["FT XYZ/1"]
	p := &saft.Payments{
		NumberOfEntries: "1",
		TotalDebit:      "123.00", // Sum of GrossTotals (Cash debit)
		TotalCredit:     "123.00", // Sum of GrossTotals (AR credit)
		Payment:         []saft.Payment{sampleValidPayment("RCPAY/001", "CUSTPAY001", "FT XYZ/1", invoiceDate.Format("2006-01-02"))},
	}
	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid Payments, got %d: %v", len(errs), errs)
	}
}

func TestValidatePayments_Invalid_NumberOfEntriesMismatch(t *testing.T) {
	invoiceDate := mockValidInvoiceRefsPay["FT XYZ/1"]
	p := &saft.Payments{
		NumberOfEntries: "2", // Mismatch
		TotalDebit:      "123.00", TotalCredit: "123.00",
		Payment: []saft.Payment{sampleValidPayment("RCPAY/001", "CUSTPAY001", "FT XYZ/1", invoiceDate.Format("2006-01-02"))},
	}
	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	if !containsErrorMsgPrefixPay(errs, "mismatch in NumberOfEntries: declared 2, actual 1") {
		t.Errorf("Expected NumberOfEntries mismatch error, got: %v", errs)
	}
}

func TestValidatePayments_Invalid_TotalDebitMismatch(t *testing.T) {
	invoiceDate := mockValidInvoiceRefsPay["FT XYZ/1"]
	p := &saft.Payments{
		NumberOfEntries: "1",
		TotalDebit:      "100.00", // Mismatch, actual is 123.00
		TotalCredit:     "123.00",
		Payment: []saft.Payment{sampleValidPayment("RCPAY/001", "CUSTPAY001", "FT XYZ/1", invoiceDate.Format("2006-01-02"))},
	}
	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	if !containsErrorMsgPrefixPay(errs, "mismatch in TotalDebit: declared 100.000000, calculated 123.000000") {
		t.Errorf("Expected TotalDebit mismatch error, got: %v", errs)
	}
}

func TestValidatePayments_Payment_Invalid_EmptyPaymentRefNo(t *testing.T) {
	invoiceDate := mockValidInvoiceRefsPay["FT XYZ/1"]
	paymentDoc := sampleValidPayment("", "CUSTPAY001", "FT XYZ/1", invoiceDate.Format("2006-01-02")) // Empty PaymentRefNo
	p := &saft.Payments{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Payment: []saft.Payment{paymentDoc},
	}
	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	if !containsErrorMsgPrefixPay(errs, "Payment[0 RefNo:]: PaymentRefNo is required") {
		t.Errorf("Expected PaymentRefNo required error, got: %v", errs)
	}
}

func TestValidatePayments_Payment_Invalid_PaymentType(t *testing.T) {
	invoiceDate := mockValidInvoiceRefsPay["FT XYZ/1"]
	paymentDoc := sampleValidPayment("RCPAY/001", "CUSTPAY001", "FT XYZ/1", invoiceDate.Format("2006-01-02"))
	paymentDoc.PaymentType = "XX" // Invalid type
	p := &saft.Payments{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Payment: []saft.Payment{paymentDoc},
	}
	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	expectedMsg := "Payment[0 RefNo:RCPAY/001]: PaymentType 'XX' is invalid"
	if !containsErrorMsgPay(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidatePayments_Payment_Invalid_PaymentStatus(t *testing.T) {
	invoiceDate := mockValidInvoiceRefsPay["FT XYZ/1"]
	paymentDoc := sampleValidPayment("RCPAY/001", "CUSTPAY001", "FT XYZ/1", invoiceDate.Format("2006-01-02"))
	paymentDoc.DocumentStatus.PaymentStatus = "X" // Invalid status
	p := &saft.Payments{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Payment: []saft.Payment{paymentDoc},
	}
	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	expectedMsg := "Payment[0 RefNo:RCPAY/001]: DocumentStatus.PaymentStatus 'X' is invalid"
	if !containsErrorMsgPay(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidatePayments_PaymentMethod_Invalid_PaymentMechanism(t *testing.T) {
	invoiceDate := mockValidInvoiceRefsPay["FT XYZ/1"]
	paymentDoc := sampleValidPayment("RCPAY/001", "CUSTPAY001", "FT XYZ/1", invoiceDate.Format("2006-01-02"))
	paymentDoc.PaymentMethod[0].PaymentMechanism = "INVALID"
	p := &saft.Payments{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Payment: []saft.Payment{paymentDoc},
	}
	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	expectedMsg := "Payment[0 RefNo:RCPAY/001] PaymentMethod[0]: PaymentMechanism 'INVALID' is invalid"
	if !containsErrorMsgPay(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidatePayments_Line_Invalid_SourceDocumentID_OriginatingONNotFound(t *testing.T) {
	invoiceDate := mockValidInvoiceRefsPay["FT XYZ/1"] // A valid date for setup, but OriginatingON will be wrong
	paymentDoc := sampleValidPayment("RCPAY/001", "CUSTPAY001", "FT UNKNOWN/7", invoiceDate.Format("2006-01-02"))
	p := &saft.Payments{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Payment: []saft.Payment{paymentDoc},
	}
	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	expectedMsg := "Payment[0 RefNo:RCPAY/001] Line[1] SourceDocumentID[0]: OriginatingON 'FT UNKNOWN/7' not found in valid invoice references"
	if !containsErrorMsgPay(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidatePayments_Line_Invalid_SourceDocumentID_InvoiceDateMismatch(t *testing.T) {
	// Correct OriginatingON but wrong date for it
	paymentDoc := sampleValidPayment("RCPAY/001", "CUSTPAY001", "FT XYZ/1", "2023-10-16") // Correct date is 2023-10-15
	p := &saft.Payments{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00",
		Payment: []saft.Payment{paymentDoc},
	}
	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	expectedMsgPrefix := "Payment[0 RefNo:RCPAY/001] Line[1] SourceDocumentID[0]: InvoiceDate '2023-10-16' does not match registered date '2023-10-15' for OriginatingON 'FT XYZ/1'"
	if !containsErrorMsgPrefixPay(errs, expectedMsgPrefix) {
		t.Errorf("Expected prefix '%s', got: %v", expectedMsgPrefix, errs)
	}
}


func TestValidatePayments_Line_MissingDebitAndCreditAmount(t *testing.T) {
	invoiceDate := mockValidInvoiceRefsPay["FT XYZ/1"]
	paymentDoc := sampleValidPayment("RCPAY/001", "CUSTPAY001", "FT XYZ/1", invoiceDate.Format("2006-01-02"))
	paymentDoc.Line[0].CreditAmount = "" // Also ensure DebitAmount is empty or not set
	paymentDoc.Line[0].DebitAmount = ""  // Explicitly empty
	p := &saft.Payments{
		NumberOfEntries: "1", TotalDebit: "123.00", TotalCredit: "123.00", // These would be affected, adjust if necessary for test focus
		Payment: []saft.Payment{paymentDoc},
	}
	// Recalculate totals for the test to focus on the line error
	p.TotalDebit = "0.00"
	p.TotalCredit = "0.00"

	errs := sourcedocuments.ValidatePayments(p, mockCustomerIDsPay, mockValidInvoiceRefsPay)
	expectedMsg := "Payment[0 RefNo:RCPAY/001] Line[1]: Either DebitAmount or CreditAmount must be provided and valid"
	if !containsErrorMsgPay(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

// Further tests could include:
// - Nil Payments input.
// - Optional fields being empty (Period, TransactionID, Description, SystemID, SettlementAmount in Line).
// - More DocumentStatus checks (SourceID, PaymentStatusDate, SourcePayment).
// - PaymentMethod: Empty PaymentAmount, invalid PaymentDate. Missing PaymentMethod block.
// - Line: Empty LineNumber, invalid LineNumber format. Missing SourceDocumentID block.
// - Line.Tax: Similar to SalesInvoices tax tests if a payment line can have its own tax implications.
// - DocumentTotals: Invalid formats for TaxPayable, NetTotal, GrossTotal. SettlementAmount format. Currency validation.
// - Multiple payments, some valid, some invalid.
// - Multiple lines per payment, multiple SourceDocumentIDs per line.
// - Multiple PaymentMethods per payment.
// - Cases where pointers (DocumentStatus, Tax, DocumentTotals, Settlement, Currency) are nil.
// - CustomerID not found.
