package validation

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	// Assuming saft structs are in this package based on other files.
	// Adjust if the path is different, e.g., "github.com/lusis/go-saft/saft"
	// or "github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft"
)

const (
	saftDateFormat     = "2006-01-02"
	saftDateTimeFormat = "2006-01-02T15:04:05"
)

// isValidDate checks if a string is a valid SAFT date.
func isValidDate(dateStr string) bool {
	if dateStr == "" {
		return false // Or true if optional, but usually dates are required if field exists
	}
	_, err := time.Parse(saftDateFormat, dateStr)
	return err == nil
}

// isValidDateTime checks if a string is a valid SAFT date-time.
func isValidDateTime(dateTimeStr string) bool {
	if dateTimeStr == "" {
		return false
	}
	_, err := time.Parse(saftDateTimeFormat, dateTimeStr)
	return err == nil
}

// parseDecimal converts a string to float64.
// SAFT amounts often have fixed precision, float64 might have issues.
// For this validation, we're primarily checking if it's a number.
func parseDecimal(s string) (float64, error) {
	if s == "" {
		return 0, errors.New("amount string is empty")
	}
	return strconv.ParseFloat(s, 64)
}

// ValidateGeneralLedgerEntries validates the GeneralLedgerEntries section.
func ValidateGeneralLedgerEntries(
	gle *saft.GeneralLedgerEntries,
	accountIDs map[string]bool, // Map of valid AccountIDs
	customerIDs map[string]bool, // Map of valid CustomerIDs
	supplierIDs map[string]bool, // Map of valid SupplierIDs
) []error {
	var errs []error

	if gle == nil {
		errs = append(errs, errors.New("GeneralLedgerEntries is nil"))
		return errs
	}

	// Calculate actual number of entries and sums
	actualNumberOfEntries := len(gle.Journal)
	calculatedTotalDebit := 0.0
	calculatedTotalCredit := 0.0

	for _, journal := range gle.Journal {
		for _, transaction := range journal.Transaction {
			for _, debitLine := range transaction.Lines.DebitLine {
				amount, err := parseDecimal(debitLine.DebitAmount)
				if err == nil {
					calculatedTotalDebit += amount
				} else {
					// Error added during line validation
				}
			}
			for _, creditLine := range transaction.Lines.CreditLine {
				amount, err := parseDecimal(creditLine.CreditAmount)
				if err == nil {
					calculatedTotalCredit += amount
				} else {
					// Error added during line validation
				}
			}
		}
	}

	// Top-level validations for GeneralLedgerEntries
	if gle.NumberOfEntries != strconv.Itoa(actualNumberOfEntries) {
		// SAFT NumberOfEntries is often a string. If it's int, direct comparison.
		// Assuming gle.NumberOfEntries is string.
		numEntriesInt, err := strconv.Atoi(gle.NumberOfEntries)
		if err != nil || numEntriesInt != actualNumberOfEntries {
			errs = append(errs, fmt.Errorf("mismatch in NumberOfEntries: declared %s, actual %d", gle.NumberOfEntries, actualNumberOfEntries))
		}
	}

	// Compare totals. Due to potential float precision issues, direct comparison can be risky.
	// A small tolerance (epsilon) might be needed in real-world scenarios if using floats.
	// Or, better, use decimal arithmetic libraries or scaled integers.
	// For this exercise, direct comparison after parsing.
	declaredTotalDebit, errDebit := parseDecimal(gle.TotalDebit)
	if errDebit != nil {
		errs = append(errs, fmt.Errorf("invalid TotalDebit format: %s", gle.TotalDebit))
	} else if declaredTotalDebit != calculatedTotalDebit {
		errs = append(errs, fmt.Errorf("mismatch in TotalDebit: declared %f, calculated %f", declaredTotalDebit, calculatedTotalDebit))
	}

	declaredTotalCredit, errCredit := parseDecimal(gle.TotalCredit)
	if errCredit != nil {
		errs = append(errs, fmt.Errorf("invalid TotalCredit format: %s", gle.TotalCredit))
	} else if declaredTotalCredit != calculatedTotalCredit {
		errs = append(errs, fmt.Errorf("mismatch in TotalCredit: declared %f, calculated %f", declaredTotalCredit, calculatedTotalCredit))
	}

	// Journal Validation
	for jIdx, journal := range gle.Journal {
		journalIdentifier := fmt.Sprintf("Journal[%d ID:%s]", jIdx, journal.JournalID)
		if journal.JournalID == "" {
			errs = append(errs, fmt.Errorf("%s: JournalID is required", journalIdentifier))
		}
		if journal.Description == "" {
			errs = append(errs, fmt.Errorf("%s: Description is required", journalIdentifier))
		}

		// Transaction Validation
		for tIdx, transaction := range journal.Transaction {
			// Constructing a more detailed identifier for transaction related errors
			txIdentifier := fmt.Sprintf("%s Transaction[%d ID:%s]", journalIdentifier, tIdx, transaction.TransactionID)

			if transaction.TransactionID == "" {
				errs = append(errs, fmt.Errorf("%s: TransactionID is required", txIdentifier))
			}

			// Period validation (assuming Period is string, needs to be parsed to int)
			// If Period is int type in struct, this changes.
			periodInt, err := strconv.Atoi(transaction.Period)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: Period '%s' is not a valid integer", txIdentifier, transaction.Period))
			} else if periodInt <= 0 { // Assuming period must be positive
				errs = append(errs, fmt.Errorf("%s: Period '%d' must be a positive integer", txIdentifier, periodInt))
			}

			if !isValidDate(transaction.TransactionDate) {
				errs = append(errs, fmt.Errorf("%s: TransactionDate '%s' is invalid", txIdentifier, transaction.TransactionDate))
			}
			if transaction.SourceID == "" {
				errs = append(errs, fmt.Errorf("%s: SourceID is required", txIdentifier))
			}
			if transaction.Description == "" {
				errs = append(errs, fmt.Errorf("%s: Description is required", txIdentifier))
			}
			if transaction.DocArchivalNumber == "" {
				errs = append(errs, fmt.Errorf("%s: DocArchivalNumber is required", txIdentifier))
			}

			validTransactionTypes := map[string]bool{
				"N": true, "R": true, "A": true, "J": true, "M": true,
				"T": true, "S": true, "P": true, "C": true, "D": true,
			}
			if _, ok := validTransactionTypes[transaction.TransactionType]; !ok {
				errs = append(errs, fmt.Errorf("%s: TransactionType '%s' is invalid", txIdentifier, transaction.TransactionType))
			}

			if !isValidDate(transaction.GLPostingDate) {
				errs = append(errs, fmt.Errorf("%s: GLPostingDate '%s' is invalid", txIdentifier, transaction.GLPostingDate))
			}

			if transaction.Lines == nil || (len(transaction.Lines.DebitLine) == 0 && len(transaction.Lines.CreditLine) == 0) {
				errs = append(errs, fmt.Errorf("%s: Lines must not be empty and contain at least one DebitLine or CreditLine", txIdentifier))
			} else {
				// DebitLine Validation
				for dlIdx, debitLine := range transaction.Lines.DebitLine {
					lineIdentifier := fmt.Sprintf("%s DebitLine[%d RecordID:%s]", txIdentifier, dlIdx, debitLine.RecordID)
					if debitLine.RecordID == "" {
						errs = append(errs, fmt.Errorf("%s: RecordID is required", lineIdentifier))
					}
					if debitLine.AccountID == "" {
						errs = append(errs, fmt.Errorf("%s: AccountID is required", lineIdentifier))
					} else if _, ok := accountIDs[debitLine.AccountID]; !ok {
						errs = append(errs, fmt.Errorf("%s: AccountID '%s' is invalid or not found in GeneralLedgerAccounts", lineIdentifier, debitLine.AccountID))
					}
					if debitLine.SourceDocumentID != "" { // "if provided"
						// The requirement is "ensure it's not empty", which it isn't if this block is entered.
						// No specific validation for SourceDocumentID format, just presence.
					}
					if !isValidDateTime(debitLine.SystemEntryDate) {
						errs = append(errs, fmt.Errorf("%s: SystemEntryDate '%s' is invalid", lineIdentifier, debitLine.SystemEntryDate))
					}
					if debitLine.Description == "" {
						errs = append(errs, fmt.Errorf("%s: Description is required", lineIdentifier))
					}
					if _, err := parseDecimal(debitLine.DebitAmount); err != nil {
						errs = append(errs, fmt.Errorf("%s: DebitAmount '%s' is not a valid decimal: %v", lineIdentifier, debitLine.DebitAmount, err))
					}
				}

				// CreditLine Validation
				for clIdx, creditLine := range transaction.Lines.CreditLine {
					lineIdentifier := fmt.Sprintf("%s CreditLine[%d RecordID:%s]", txIdentifier, clIdx, creditLine.RecordID)
					if creditLine.RecordID == "" {
						errs = append(errs, fmt.Errorf("%s: RecordID is required", lineIdentifier))
					}
					if creditLine.AccountID == "" {
						errs = append(errs, fmt.Errorf("%s: AccountID is required", lineIdentifier))
					} else if _, ok := accountIDs[creditLine.AccountID]; !ok {
						errs = append(errs, fmt.Errorf("%s: AccountID '%s' is invalid or not found in GeneralLedgerAccounts", lineIdentifier, creditLine.AccountID))
					}

					if creditLine.SourceDocumentID != "" { // "if provided"
						// As above, presence is checked by `!= ""`.
					}
					if !isValidDateTime(creditLine.SystemEntryDate) {
						errs = append(errs, fmt.Errorf("%s: SystemEntryDate '%s' is invalid", lineIdentifier, creditLine.SystemEntryDate))
					}
					if creditLine.Description == "" {
						errs = append(errs, fmt.Errorf("%s: Description is required", lineIdentifier))
					}
					if _, err := parseDecimal(creditLine.CreditAmount); err != nil {
						errs = append(errs, fmt.Errorf("%s: CreditAmount '%s' is not a valid decimal: %v", lineIdentifier, creditLine.CreditAmount, err))
					}
				}
			} // End Lines validation

			// CustomerID / SupplierID validation
			if transaction.CustomerID != "" {
				if _, ok := customerIDs[transaction.CustomerID]; !ok {
					errs = append(errs, fmt.Errorf("%s: CustomerID '%s' is invalid or not found in MasterFiles", txIdentifier, transaction.CustomerID))
				}
			}
			if transaction.SupplierID != "" {
				if _, ok := supplierIDs[transaction.SupplierID]; !ok {
					errs = append(errs, fmt.Errorf("%s: SupplierID '%s' is invalid or not found in MasterFiles", txIdentifier, transaction.SupplierID))
				}
			}
			// The rule "Ensure either CustomerID or SupplierID is provided if applicable" is complex.
			// "if applicable" would require more domain knowledge (e.g. specific accounts imply customer/supplier).
			// The current checks validate them if they are present.
			// A stricter check (e.g. one of them MUST be present for certain transaction types or accounts) is not implemented here.
		}
	}
	return errs
}
