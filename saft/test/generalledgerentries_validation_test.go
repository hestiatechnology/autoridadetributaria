package test

import (
	"testing"
	"time" // For time.Parse in test setup if needed for validInvoiceReferences

	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/generalledgerentries" // Assuming this is the package path
)

// Helper functions (can be shared)
func containsErrorMsgGLE(errs []error, msg string) bool {
	for _, err := range errs {
		if err.Error() == msg {
			return true
		}
	}
	return false
}

func containsErrorMsgPrefixGLE(errs []error, prefix string) bool {
	for _, err := range errs {
		// Using strings.HasPrefix might be cleaner if allowed/imported
		if len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// Mock data for dependencies
var (
	mockAccountIDsGLE = map[string]bool{
		"111": true, "112": true, "211": true, "611": true, "711": true,
	}
	mockCustomerIDsGLE = map[string]bool{"CUST001": true}
	mockSupplierIDsGLE = map[string]bool{"SUP001": true}
)


func TestValidateGeneralLedgerEntries_Valid(t *testing.T) {
	gle := &saft.GeneralLedgerEntries{
		NumberOfEntries: "1", // Number of Journals
		TotalDebit:      "150.00",
		TotalCredit:     "150.00",
		Journal: []saft.Journal{
			{
				JournalID:   "J001",
				Description: "Monthly Operations",
				Transaction: []saft.Transaction{
					{
						TransactionID:     "T202300001",
						Period:            "10",
						TransactionDate:   "2023-10-28",
						SourceID:          "UserXYZ",
						Description:       "Cash Sale",
						DocArchivalNumber: "DOC001",
						TransactionType:   "N", // Normal
						GLPostingDate:     "2023-10-28",
						CustomerID:        "CUST001",
						Lines: &saft.TransactionLines{
							DebitLine: []saft.DebitLine{
								{RecordID: "REC001D", AccountID: "111", SystemEntryDate: "2023-10-28T10:00:00", Description: "Cash debit", DebitAmount: "150.00"},
							},
							CreditLine: []saft.CreditLine{
								{RecordID: "REC001C", AccountID: "711", SystemEntryDate: "2023-10-28T10:00:00", Description: "Sales credit", CreditAmount: "150.00"},
							},
						},
					},
				},
			},
		},
	}

	errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid GeneralLedgerEntries, got %d: %v", len(errs), errs)
	}
}

func TestValidateGeneralLedgerEntries_Invalid_NumberOfEntriesMismatch(t *testing.T) {
	gle := &saft.GeneralLedgerEntries{
		NumberOfEntries: "2", // Mismatch, only 1 journal
		TotalDebit:      "150.00",
		TotalCredit:     "150.00",
		Journal: []saft.Journal{
			{ JournalID: "J001", Description: "Desc", Transaction: []saft.Transaction{/*...*/}},
		},
	}
	errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
	if !containsErrorMsgPrefixGLE(errs, "mismatch in NumberOfEntries: declared 2, actual 1") {
		t.Errorf("Expected NumberOfEntries mismatch error, got: %v", errs)
	}
}

func TestValidateGeneralLedgerEntries_Invalid_TotalDebitMismatch(t *testing.T) {
	gle := &saft.GeneralLedgerEntries{
		NumberOfEntries: "1",
		TotalDebit:      "100.00", // Mismatch, actual is 150.00 from lines below
		TotalCredit:     "150.00",
		Journal: []saft.Journal{
			{
				JournalID: "J001", Description: "Desc",
				Transaction: []saft.Transaction{
					{ /* Assume valid transaction structure from TestValidateGeneralLedgerEntries_Valid */
						TransactionID: "T001", Period: "1", TransactionDate: "2023-01-01", SourceID: "S1", Description: "D1", DocArchivalNumber: "DAN1", TransactionType: "N", GLPostingDate: "2023-01-01",
						Lines: &saft.TransactionLines{ DebitLine: []saft.DebitLine{{RecordID: "R1", AccountID: "111", SystemEntryDate: "2023-01-01T00:00:00", Description: "DL1", DebitAmount: "150.00"}}, CreditLine: []saft.CreditLine{{RecordID: "R2", AccountID: "711", SystemEntryDate: "2023-01-01T00:00:00", Description: "CL1", CreditAmount: "150.00"}}},
					},
				},
			},
		},
	}
	errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
	if !containsErrorMsgPrefixGLE(errs, "mismatch in TotalDebit: declared 100.000000, calculated 150.000000") {
		t.Errorf("Expected TotalDebit mismatch error, got: %v", errs)
	}
}

func TestValidateGeneralLedgerEntries_Invalid_Journal_EmptyID(t *testing.T) {
	gle := &saft.GeneralLedgerEntries{
		NumberOfEntries: "1", TotalDebit: "0.00", TotalCredit: "0.00",
		Journal: []saft.Journal{{JournalID: "", Description: "Desc"}},
	}
	errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
	if !containsErrorMsgPrefixGLE(errs, "Journal[0 ID:]: JournalID is required") {
		t.Errorf("Expected JournalID required error, got: %v", errs)
	}
}

func TestValidateGeneralLedgerEntries_Invalid_Transaction_InvalidDate(t *testing.T) {
	gle := &saft.GeneralLedgerEntries{
		NumberOfEntries: "1", TotalDebit: "0.00", TotalCredit: "0.00",
		Journal: []saft.Journal{
			{ JournalID: "J1", Description: "JDesc",
				Transaction: []saft.Transaction{
					{TransactionID: "T1", Period: "1", TransactionDate: "2023-13-01", /* Invalid month */ SourceID: "S1", Description: "D1", DocArchivalNumber: "DAN1", TransactionType: "N", GLPostingDate: "2023-01-01", Lines: &saft.TransactionLines{DebitLine:[]saft.DebitLine{}, CreditLine:[]saft.CreditLine{}}},
				},
			},
		},
	}
	// Add a dummy line to avoid "Lines must not be empty" error, focusing on date error
	gle.Journal[0].Transaction[0].Lines.DebitLine = append(gle.Journal[0].Transaction[0].Lines.DebitLine, saft.DebitLine{RecordID:"temp", AccountID:"111", SystemEntryDate:"2023-10-28T10:00:00", Description:"temp", DebitAmount:"0.00"})

	errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
	if !containsErrorMsgPrefixGLE(errs, "Journal[0 ID:J1] Transaction[0 ID:T1]: TransactionDate '2023-13-01' is invalid") {
		t.Errorf("Expected TransactionDate invalid error, got: %v", errs)
	}
}

func TestValidateGeneralLedgerEntries_Invalid_Transaction_InvalidType(t *testing.T) {
	gle := &saft.GeneralLedgerEntries{
		NumberOfEntries: "1", TotalDebit: "0.00", TotalCredit: "0.00",
		Journal: []saft.Journal{
			{ JournalID: "J1", Description: "JDesc",
				Transaction: []saft.Transaction{
					{TransactionID: "T1", Period: "1", TransactionDate: "2023-10-01", SourceID: "S1", Description: "D1", DocArchivalNumber: "DAN1", TransactionType: "INVALID", GLPostingDate: "2023-10-01", Lines: &saft.TransactionLines{DebitLine:[]saft.DebitLine{}, CreditLine:[]saft.CreditLine{}}},
				},
			},
		},
	}
	gle.Journal[0].Transaction[0].Lines.DebitLine = append(gle.Journal[0].Transaction[0].Lines.DebitLine, saft.DebitLine{RecordID:"temp", AccountID:"111", SystemEntryDate:"2023-10-28T10:00:00", Description:"temp", DebitAmount:"0.00"})

	errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
	if !containsErrorMsgPrefixGLE(errs, "Journal[0 ID:J1] Transaction[0 ID:T1]: TransactionType 'INVALID' is invalid") {
		t.Errorf("Expected TransactionType invalid error, got: %v", errs)
	}
}

func TestValidateGeneralLedgerEntries_Invalid_Lines_Empty(t *testing.T) {
	gle := &saft.GeneralLedgerEntries{
		NumberOfEntries: "1", TotalDebit: "0.00", TotalCredit: "0.00",
		Journal: []saft.Journal{
			{ JournalID: "J1", Description: "JDesc",
				Transaction: []saft.Transaction{
					{TransactionID: "T1", Period: "1", TransactionDate: "2023-10-01", SourceID: "S1", Description: "D1", DocArchivalNumber: "DAN1", TransactionType: "N", GLPostingDate: "2023-10-01", Lines: &saft.TransactionLines{}}, // Empty lines
				},
			},
		},
	}
	errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
	if !containsErrorMsgPrefixGLE(errs, "Journal[0 ID:J1] Transaction[0 ID:T1]: Lines must not be empty and contain at least one DebitLine or CreditLine") {
		t.Errorf("Expected Lines empty error, got: %v", errs)
	}
}

func TestValidateGeneralLedgerEntries_Invalid_DebitLine_InvalidAccountID(t *testing.T) {
	gle := &saft.GeneralLedgerEntries{
		NumberOfEntries: "1", TotalDebit: "10.00", TotalCredit: "0.00", // Adjusted for the single debit line
		Journal: []saft.Journal{
			{ JournalID: "J1", Description: "JDesc",
				Transaction: []saft.Transaction{
					{TransactionID: "T1", Period: "1", TransactionDate: "2023-10-01", SourceID: "S1", Description: "D1", DocArchivalNumber: "DAN1", TransactionType: "N", GLPostingDate: "2023-10-01",
						Lines: &saft.TransactionLines{DebitLine: []saft.DebitLine{{RecordID: "R1", AccountID: "INVALIDACC", SystemEntryDate: "2023-10-01T12:00:00", Description: "DL1", DebitAmount: "10.00"}}},
					},
				},
			},
		},
	}
	errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
	expectedMsg := "Journal[0 ID:J1] Transaction[0 ID:T1] DebitLine[0 RecordID:R1]: AccountID 'INVALIDACC' is invalid or not found in GeneralLedgerAccounts"
	if !containsErrorMsgPrefixGLE(errs, expectedMsg) {
		t.Errorf("Expected DebitLine AccountID invalid error, got: %v", errs)
	}
}


func TestValidateGeneralLedgerEntries_Invalid_CreditLine_InvalidSystemEntryDate(t *testing.T) {
	gle := &saft.GeneralLedgerEntries{
		NumberOfEntries: "1", TotalDebit: "0.00", TotalCredit: "20.00", // Adjusted
		Journal: []saft.Journal{
			{ JournalID: "J1", Description: "JDesc",
				Transaction: []saft.Transaction{
					{TransactionID: "T1", Period: "1", TransactionDate: "2023-10-01", SourceID: "S1", Description: "D1", DocArchivalNumber: "DAN1", TransactionType: "N", GLPostingDate: "2023-10-01",
						Lines: &saft.TransactionLines{CreditLine: []saft.CreditLine{{RecordID: "R1", AccountID: "711", SystemEntryDate: "2023-10-01T25:00:00" /*Invalid hour*/, Description: "CL1", CreditAmount: "20.00"}}},
					},
				},
			},
		},
	}
	errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
	expectedMsg := "Journal[0 ID:J1] Transaction[0 ID:T1] CreditLine[0 RecordID:R1]: SystemEntryDate '2023-10-01T25:00:00' is invalid"
	if !containsErrorMsgPrefixGLE(errs, expectedMsg) {
		t.Errorf("Expected CreditLine SystemEntryDate invalid error, got: %v", errs)
	}
}


func TestValidateGeneralLedgerEntries_Invalid_Transaction_InvalidCustomerID(t *testing.T) {
    gle := &saft.GeneralLedgerEntries{
        NumberOfEntries: "1", TotalDebit: "0.00", TotalCredit: "0.00",
        Journal: []saft.Journal{
            {JournalID: "J1", Description: "JDesc",
                Transaction: []saft.Transaction{
                    {
                        TransactionID: "T1", Period: "1", TransactionDate: "2023-10-01", SourceID: "S1",
                        Description: "D1", DocArchivalNumber: "DAN1", TransactionType: "N", GLPostingDate: "2023-10-01",
                        CustomerID: "INVALIDCUST", // Invalid CustomerID
                        Lines: &saft.TransactionLines{
                            DebitLine:  []saft.DebitLine{{RecordID: "tempD", AccountID: "111", SystemEntryDate: "2023-10-28T10:00:00", Description: "tempD", DebitAmount: "0.00"}},
                            CreditLine: []saft.CreditLine{{RecordID: "tempC", AccountID: "111", SystemEntryDate: "2023-10-28T10:00:00", Description: "tempC", CreditAmount: "0.00"}},
                        },
                    },
                },
            },
        },
    }
    errs := generalledgerentries.ValidateGeneralLedgerEntries(gle, mockAccountIDsGLE, mockCustomerIDsGLE, mockSupplierIDsGLE)
    expectedMsg := "Journal[0 ID:J1] Transaction[0 ID:T1]: CustomerID 'INVALIDCUST' is invalid or not found in MasterFiles"
    if !containsErrorMsgPrefixGLE(errs, expectedMsg) {
        t.Errorf("Expected invalid CustomerID error, got: %v", errs)
    }
}

// More tests would cover:
// - Nil GLE
// - Invalid TotalCredit format
// - Invalid DebitAmount / CreditAmount format
// - Empty required fields in Journal, Transaction, DebitLine, CreditLine (Description, RecordID etc.)
// - Invalid Period format (not an integer)
// - Invalid SupplierID
// - Combinations of errors
// - Edge cases for dates/datetimes (e.g. leap years, different timezones if applicable, though SAFT usually implies local time)
// - Transactions with only DebitLine or only CreditLine (which is valid)
// - Multiple journal entries, multiple transactions.
// - Correct calculation of totals with multiple lines/transactions/journals.
// - Validation of SourceDocumentID in lines (if provided, not empty - current validator is simple for this).
// - The "either CustomerID or SupplierID is provided if applicable" rule is hard to test without specific "applicability" criteria.
//   Current validator only checks them if present.
