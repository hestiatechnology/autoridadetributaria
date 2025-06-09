package validation // Assuming this is the correct package name based on other files.

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/saft" // Adjusted import path
)

const (
	saftDateFormat     = "2006-01-02"
	saftDateTimeFormat = "2006-01-02T15:04:05"
)

var (
	countryCodeRegex  = regexp.MustCompile(`^[A-Z]{2}$`)
	currencyCodeRegex = regexp.MustCompile(`^[A-Z]{3}$`)
)

// isValidDate checks if a string is a valid SAFT date.
func isValidDate(dateStr string) bool {
	if dateStr == "" {
		return false
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
func parseDecimal(s string) (float64, error) {
	if s == "" {
		return 0, errors.New("amount string is empty")
	}
	// SAFT often uses "." as decimal separator.
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

// isValidInteger checks if a string is a valid integer with options.
func isValidInteger(s string, allowZero bool, allowNegative bool) (int, bool) {
	if s == "" {
		return 0, false
	}
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, false
	}
	if !allowZero && i == 0 {
		return 0, false
	}
	if !allowNegative && i < 0 {
		return 0, false
	}
	return i, true
}

// ValidateSalesInvoices validates the SalesInvoices section.
func ValidateSalesInvoices(
	si *saft.SalesInvoices,
	productCodes map[string]bool, // Map of valid ProductCodes
	customerIDs map[string]bool,  // Map of valid CustomerIDs
	// taxTableEntries map[string]saft.TaxTableEntry, // For deeper tax validation if needed
) []error {
	var errs []error

	if si == nil {
		errs = append(errs, errors.New("SalesInvoices is nil"))
		return errs
	}

	actualNumberOfEntries := len(si.Invoice)
	calculatedTotalDebit := 0.0
	calculatedTotalCredit := 0.0

	for _, invoice := range si.Invoice {
		if invoice.DocumentTotals != nil {
			grossTotal, err := parseDecimal(invoice.DocumentTotals.GrossTotal)
			if err == nil {
				calculatedTotalCredit += grossTotal
				calculatedTotalDebit += grossTotal // For Sales, Debit to AR, Credit to Sales/VAT
			}
		}
	}

	numEntriesDeclared, err := strconv.Atoi(si.NumberOfEntries)
	if err != nil || numEntriesDeclared != actualNumberOfEntries {
		errs = append(errs, fmt.Errorf("mismatch in NumberOfEntries: declared %s, actual %d", si.NumberOfEntries, actualNumberOfEntries))
	}

	totalDebitDeclared, errDb := parseDecimal(si.TotalDebit)
	if errDb != nil {
		errs = append(errs, fmt.Errorf("invalid TotalDebit format: %s", si.TotalDebit))
	} else if totalDebitDeclared != calculatedTotalDebit { // Basic float comparison
		errs = append(errs, fmt.Errorf("mismatch in TotalDebit: declared %f, calculated %f", totalDebitDeclared, calculatedTotalDebit))
	}

	totalCreditDeclared, errCr := parseDecimal(si.TotalCredit)
	if errCr != nil {
		errs = append(errs, fmt.Errorf("invalid TotalCredit format: %s", si.TotalCredit))
	} else if totalCreditDeclared != calculatedTotalCredit { // Basic float comparison
		errs = append(errs, fmt.Errorf("mismatch in TotalCredit: declared %f, calculated %f", totalCreditDeclared, calculatedTotalCredit))
	}

	for invIdx, invoice := range si.Invoice {
		invIdentifier := fmt.Sprintf("Invoice[%d No:%s]", invIdx, invoice.InvoiceNo)

		if invoice.InvoiceNo == "" {
			errs = append(errs, fmt.Errorf("%s: InvoiceNo is required", invIdentifier))
		}
		if invoice.ATCUD == "" {
			errs = append(errs, fmt.Errorf("%s: ATCUD is required", invIdentifier))
		}

		// DocumentStatus Validation
		if invoice.DocumentStatus == nil {
			errs = append(errs, fmt.Errorf("%s: DocumentStatus is required", invIdentifier))
		} else {
			ds := invoice.DocumentStatus
			validInvoiceStatus := map[string]bool{"N": true, "S": true, "A": true, "R": true, "F": true, "T": true}
			if _, ok := validInvoiceStatus[ds.InvoiceStatus]; !ok {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.InvoiceStatus '%s' is invalid", invIdentifier, ds.InvoiceStatus))
			}
			if !isValidDateTime(ds.InvoiceStatusDate) {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.InvoiceStatusDate '%s' is invalid", invIdentifier, ds.InvoiceStatusDate))
			}
			if ds.SourceID == "" { // SourceID within DocumentStatus
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.SourceID is required", invIdentifier))
			}
			validSourceBilling := map[string]bool{"P": true, "I": true, "M": true}
			if _, ok := validSourceBilling[ds.SourceBilling]; !ok {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.SourceBilling '%s' is invalid", invIdentifier, ds.SourceBilling))
			}
		}

		if invoice.Hash == "" {
			errs = append(errs, fmt.Errorf("%s: Hash is required", invIdentifier))
		}
		if invoice.HashControl == "" {
			errs = append(errs, fmt.Errorf("%s: HashControl is required", invIdentifier))
		}

		if _, ok := isValidInteger(invoice.Period, false, false); !ok { // Period must be positive integer
			errs = append(errs, fmt.Errorf("%s: Period '%s' is not a valid positive integer", invIdentifier, invoice.Period))
		}
		if !isValidDate(invoice.InvoiceDate) {
			errs = append(errs, fmt.Errorf("%s: InvoiceDate '%s' is invalid", invIdentifier, invoice.InvoiceDate))
		}

		validInvoiceTypes := map[string]bool{
			"FT": true, "FS": true, "FR": true, "ND": true, "NC": true,
			"AA": true, "DA": true, "RP": true, "RE": true, "CS": true, "LD": true, "RA": true,
		}
		if _, ok := validInvoiceTypes[invoice.InvoiceType]; !ok {
			errs = append(errs, fmt.Errorf("%s: InvoiceType '%s' is invalid", invIdentifier, invoice.InvoiceType))
		}

		// SpecialRegimes Validation
		if invoice.SpecialRegimes == nil {
			errs = append(errs, fmt.Errorf("%s: SpecialRegimes is required", invIdentifier))
		} else {
			sr := invoice.SpecialRegimes
			if sr.SelfBillingIndicator != "0" && sr.SelfBillingIndicator != "1" {
				errs = append(errs, fmt.Errorf("%s: SpecialRegimes.SelfBillingIndicator '%s' is invalid (must be \"0\" or \"1\")", invIdentifier, sr.SelfBillingIndicator))
			}
			// Assuming CashVATSchemeIndicator is int 0 or 1. If string, adapt.
			if sr.CashVATSchemeIndicator != 0 && sr.CashVATSchemeIndicator != 1 {
				errs = append(errs, fmt.Errorf("%s: SpecialRegimes.CashVATSchemeIndicator '%d' is invalid (must be 0 or 1)", invIdentifier, sr.CashVATSchemeIndicator))
			}
			if sr.ThirdPartiesBillingIndicator != "0" && sr.ThirdPartiesBillingIndicator != "1" {
				errs = append(errs, fmt.Errorf("%s: SpecialRegimes.ThirdPartiesBillingIndicator '%s' is invalid (must be \"0\" or \"1\")", invIdentifier, sr.ThirdPartiesBillingIndicator))
			}
		}

		if invoice.SourceID == "" { // SourceID at Invoice level
			errs = append(errs, fmt.Errorf("%s: SourceID (at Invoice level) is required", invIdentifier))
		}
		if !isValidDateTime(invoice.SystemEntryDate) {
			errs = append(errs, fmt.Errorf("%s: SystemEntryDate '%s' is invalid", invIdentifier, invoice.SystemEntryDate))
		}
		if invoice.CustomerID == "" {
			errs = append(errs, fmt.Errorf("%s: CustomerID is required", invIdentifier))
		} else if _, ok := customerIDs[invoice.CustomerID]; !ok {
			errs = append(errs, fmt.Errorf("%s: CustomerID '%s' is invalid or not found", invIdentifier, invoice.CustomerID))
		}

		// Line Validation
		for lineIdx, line := range invoice.Line {
			lineIdentifier := fmt.Sprintf("%s Line[%d]", invIdentifier, lineIdx+1) // LineNumber is 1-based

			ln, ok := isValidInteger(line.LineNumber, false, false)
			if !ok || ln != lineIdx+1 { // Check sequence if LineNumber is string, or just positive if int
				errs = append(errs, fmt.Errorf("%s: LineNumber '%s' is invalid or out of sequence (expected %d)", lineIdentifier, line.LineNumber, lineIdx+1))
			}

			if line.ProductCode == "" {
				errs = append(errs, fmt.Errorf("%s: ProductCode is required", lineIdentifier))
			} else if _, ok := productCodes[line.ProductCode]; !ok {
				errs = append(errs, fmt.Errorf("%s: ProductCode '%s' is invalid or not found", lineIdentifier, line.ProductCode))
			}
			if line.ProductDescription == "" {
				errs = append(errs, fmt.Errorf("%s: ProductDescription is required", lineIdentifier))
			}
			if _, err := parseDecimal(line.Quantity); err != nil {
				errs = append(errs, fmt.Errorf("%s: Quantity '%s' is not a valid decimal: %v", lineIdentifier, line.Quantity, err))
			}
			if line.UnitOfMeasure == "" {
				errs = append(errs, fmt.Errorf("%s: UnitOfMeasure is required", lineIdentifier))
			}
			if _, err := parseDecimal(line.UnitPrice); err != nil {
				errs = append(errs, fmt.Errorf("%s: UnitPrice '%s' is not a valid decimal: %v", lineIdentifier, line.UnitPrice, err))
			}
			if !isValidDate(line.TaxPointDate) {
				errs = append(errs, fmt.Errorf("%s: TaxPointDate '%s' is invalid", lineIdentifier, line.TaxPointDate))
			}
			if line.Description == "" { // Line description
				errs = append(errs, fmt.Errorf("%s: Description (line level) is required", lineIdentifier))
			}

			// Tax Validation (Line.Tax)
			if line.Tax == nil {
				errs = append(errs, fmt.Errorf("%s: Tax information is required", lineIdentifier))
			} else {
				tax := line.Tax
				validTaxTypes := map[string]bool{"IVA": true, "IS": true, "NS": true}
				if _, ok := validTaxTypes[tax.TaxType]; !ok {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxType '%s' is invalid", lineIdentifier, tax.TaxType))
				}
				if !countryCodeRegex.MatchString(tax.TaxCountryRegion) {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxCountryRegion '%s' is not a valid country code", lineIdentifier, tax.TaxCountryRegion))
				}
				if tax.TaxCode == "" {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxCode is required", lineIdentifier))
				}

				hasTaxPercentage := tax.TaxPercentage != "" && tax.TaxPercentage != "0" // "0" might mean exempt or not applicable
				hasTaxAmount := tax.TaxAmount != "" && tax.TaxAmount != "0"

				if tax.TaxPercentage != "" {
					if _, err := parseDecimal(tax.TaxPercentage); err != nil {
						errs = append(errs, fmt.Errorf("%s: Tax.TaxPercentage '%s' is not a valid decimal: %v", lineIdentifier, tax.TaxPercentage, err))
						hasTaxPercentage = false // Mark as invalid for the check below
					}
				}
				if tax.TaxAmount != "" {
					 if _, err := parseDecimal(tax.TaxAmount); err != nil {
						errs = append(errs, fmt.Errorf("%s: Tax.TaxAmount '%s' is not a valid decimal: %v", lineIdentifier, tax.TaxAmount, err))
						hasTaxAmount = false // Mark as invalid
					}
				}

				// For some TaxTypes (e.g. NS or specific IS codes), percentage/amount might not be applicable.
				// This simplified check assumes one should be valid if the other is not, for IVA mostly.
				if tax.TaxType == "IVA" && !hasTaxPercentage && !hasTaxAmount {
                     errs = append(errs, fmt.Errorf("%s: For TaxType IVA, either TaxPercentage or TaxAmount must be provided and valid", lineIdentifier))
                } else if tax.TaxType == "IVA" && tax.TaxPercentage == "" && tax.TaxAmount == "" {
					// If both are empty strings (not just "0")
					errs = append(errs, fmt.Errorf("%s: For TaxType IVA, either TaxPercentage or TaxAmount must be provided", lineIdentifier))
				}


			}
		} // End Line Validation

		// DocumentTotals Validation
		if invoice.DocumentTotals == nil {
			errs = append(errs, fmt.Errorf("%s: DocumentTotals is required", invIdentifier))
		} else {
			dt := invoice.DocumentTotals
			if _, err := parseDecimal(dt.TaxPayable); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.TaxPayable '%s' is not a valid decimal: %v", invIdentifier, dt.TaxPayable, err))
			}
			if _, err := parseDecimal(dt.NetTotal); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.NetTotal '%s' is not a valid decimal: %v", invIdentifier, dt.NetTotal, err))
			}
			if _, err := parseDecimal(dt.GrossTotal); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.GrossTotal '%s' is not a valid decimal: %v", invIdentifier, dt.GrossTotal, err))
			}

			// Currency Validation (If Currency block exists and is not base currency)
			// Assuming saft.Currency is a pointer or a struct that can be checked for presence.
			// The exact structure of saft.Currency (e.g. if it's *saft.CurrencyType or saft.CurrencyType) matters.
			// For this example, let's assume if CurrencyCode is present, the block is active.
			if dt.Currency != nil && dt.Currency.CurrencyCode != "" { // Check if currency info is provided
				if !currencyCodeRegex.MatchString(dt.Currency.CurrencyCode) {
					errs = append(errs, fmt.Errorf("%s: DocumentTotals.Currency.CurrencyCode '%s' is invalid", invIdentifier, dt.Currency.CurrencyCode))
				}
				// CurrencyAmount would be GrossTotal in specified currency.
				// It should be a valid decimal.
				if _, err := parseDecimal(dt.Currency.CurrencyAmount); err != nil {
					errs = append(errs, fmt.Errorf("%s: DocumentTotals.Currency.CurrencyAmount '%s' is not a valid decimal: %v", invIdentifier, dt.Currency.CurrencyAmount, err))
				}
				// ExchangeRate also needs to be a valid decimal if present
				// if dt.Currency.ExchangeRate != "" { // Assuming ExchangeRate is a string
				//    if _, err := parseDecimal(dt.Currency.ExchangeRate); err != nil {
				//        errs = append(errs, fmt.Errorf("%s: DocumentTotals.Currency.ExchangeRate '%s' is not a valid decimal: %v", invIdentifier, dt.Currency.ExchangeRate, err))
				//    }
				// }
			}
		}
	}
	return errs
}

// ValidateWorkingDocuments validates the WorkingDocuments section.
func ValidateWorkingDocuments(
	wd *saft.WorkingDocuments,
	productCodes map[string]bool, // Map of valid ProductCodes
	customerIDs map[string]bool,  // Map of valid CustomerIDs
	// taxTableEntries map[string]saft.TaxTableEntry, // For deeper tax validation if needed
) []error {
	var errs []error

	if wd == nil {
		errs = append(errs, errors.New("WorkingDocuments is nil"))
		return errs
	}

	actualNumberOfEntries := len(wd.WorkDocument)
	calculatedTotalDebit := 0.0
	calculatedTotalCredit := 0.0

	for _, workDoc := range wd.WorkDocument {
		if workDoc.DocumentTotals != nil {
			grossTotal, err := parseDecimal(workDoc.DocumentTotals.GrossTotal)
			if err == nil {
				calculatedTotalCredit += grossTotal
				calculatedTotalDebit += grossTotal // Assuming similar to SalesInvoices for debit/credit summary
			}
		}
	}

	numEntriesDeclared, err := strconv.Atoi(wd.NumberOfEntries)
	if err != nil || numEntriesDeclared != actualNumberOfEntries {
		errs = append(errs, fmt.Errorf("mismatch in NumberOfEntries: declared %s, actual %d", wd.NumberOfEntries, actualNumberOfEntries))
	}

	totalDebitDeclared, errDb := parseDecimal(wd.TotalDebit)
	if errDb != nil {
		errs = append(errs, fmt.Errorf("invalid TotalDebit format: %s", wd.TotalDebit))
	} else if totalDebitDeclared != calculatedTotalDebit {
		errs = append(errs, fmt.Errorf("mismatch in TotalDebit: declared %f, calculated %f", totalDebitDeclared, calculatedTotalDebit))
	}

	totalCreditDeclared, errCr := parseDecimal(wd.TotalCredit)
	if errCr != nil {
		errs = append(errs, fmt.Errorf("invalid TotalCredit format: %s", wd.TotalCredit))
	} else if totalCreditDeclared != calculatedTotalCredit {
		errs = append(errs, fmt.Errorf("mismatch in TotalCredit: declared %f, calculated %f", totalCreditDeclared, calculatedTotalCredit))
	}

	for wdIdx, workDoc := range wd.WorkDocument {
		wdIdentifier := fmt.Sprintf("WorkDocument[%d No:%s]", wdIdx, workDoc.DocumentNumber)

		if workDoc.DocumentNumber == "" {
			errs = append(errs, fmt.Errorf("%s: DocumentNumber is required", wdIdentifier))
		}
		if workDoc.ATCUD == "" {
			errs = append(errs, fmt.Errorf("%s: ATCUD is required", wdIdentifier))
		}

		// DocumentStatus Validation
		if workDoc.DocumentStatus == nil {
			errs = append(errs, fmt.Errorf("%s: DocumentStatus is required", wdIdentifier))
		} else {
			ds := workDoc.DocumentStatus
			validWorkStatus := map[string]bool{"N": true, "A": true, "F": true, "C": true}
			if _, ok := validWorkStatus[ds.WorkStatus]; !ok {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.WorkStatus '%s' is invalid", wdIdentifier, ds.WorkStatus))
			}
			if !isValidDateTime(ds.WorkStatusDate) {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.WorkStatusDate '%s' is invalid", wdIdentifier, ds.WorkStatusDate))
			}
			if ds.SourceID == "" {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.SourceID is required", wdIdentifier))
			}
			validSourceBilling := map[string]bool{"P": true, "I": true, "M": true} // Reused
			if _, ok := validSourceBilling[ds.SourceBilling]; !ok {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.SourceBilling '%s' is invalid", wdIdentifier, ds.SourceBilling))
			}
		}

		if workDoc.Hash == "" {
			errs = append(errs, fmt.Errorf("%s: Hash is required", wdIdentifier))
		}
		if workDoc.HashControl == "" {
			errs = append(errs, fmt.Errorf("%s: HashControl is required", wdIdentifier))
		}

		if _, ok := isValidInteger(workDoc.Period, false, false); !ok {
			errs = append(errs, fmt.Errorf("%s: Period '%s' is not a valid positive integer", wdIdentifier, workDoc.Period))
		}
		if !isValidDate(workDoc.WorkDate) {
			errs = append(errs, fmt.Errorf("%s: WorkDate '%s' is invalid", wdIdentifier, workDoc.WorkDate))
		}

		validWorkTypes := map[string]bool{
			"CM": true, "CC": true, "FC": true, "FO": true, "NE": true, "OU": true,
			"OR": true, "PF": true, "DC": true, "RP": true, "RE": true, "CS": true,
			"GD": true, "GT": true, "GC": true,
		}
		if _, ok := validWorkTypes[workDoc.WorkType]; !ok {
			errs = append(errs, fmt.Errorf("%s: WorkType '%s' is invalid", wdIdentifier, workDoc.WorkType))
		}

		if workDoc.SourceID == "" { // SourceID at WorkDocument level
			errs = append(errs, fmt.Errorf("%s: SourceID (at WorkDocument level) is required", wdIdentifier))
		}
		if !isValidDateTime(workDoc.SystemEntryDate) {
			errs = append(errs, fmt.Errorf("%s: SystemEntryDate '%s' is invalid", wdIdentifier, workDoc.SystemEntryDate))
		}
		if workDoc.CustomerID == "" {
			errs = append(errs, fmt.Errorf("%s: CustomerID is required", wdIdentifier))
		} else if _, ok := customerIDs[workDoc.CustomerID]; !ok {
			errs = append(errs, fmt.Errorf("%s: CustomerID '%s' is invalid or not found", wdIdentifier, workDoc.CustomerID))
		}

		// Line Validation (similar to SalesInvoices.Invoice.Line)
		for lineIdx, line := range workDoc.Line {
			lineIdentifier := fmt.Sprintf("%s Line[%d]", wdIdentifier, lineIdx+1)

			ln, ok := isValidInteger(line.LineNumber, false, false)
			if !ok || ln != lineIdx+1 {
				errs = append(errs, fmt.Errorf("%s: LineNumber '%s' is invalid or out of sequence (expected %d)", lineIdentifier, line.LineNumber, lineIdx+1))
			}

			if line.ProductCode == "" {
				errs = append(errs, fmt.Errorf("%s: ProductCode is required", lineIdentifier))
			} else if _, ok := productCodes[line.ProductCode]; !ok {
				errs = append(errs, fmt.Errorf("%s: ProductCode '%s' is invalid or not found", lineIdentifier, line.ProductCode))
			}
			if line.ProductDescription == "" {
				errs = append(errs, fmt.Errorf("%s: ProductDescription is required", lineIdentifier))
			}
			if _, err := parseDecimal(line.Quantity); err != nil {
				errs = append(errs, fmt.Errorf("%s: Quantity '%s' is not a valid decimal: %v", lineIdentifier, line.Quantity, err))
			}
			if line.UnitOfMeasure == "" {
				errs = append(errs, fmt.Errorf("%s: UnitOfMeasure is required", lineIdentifier))
			}
			if _, err := parseDecimal(line.UnitPrice); err != nil {
				errs = append(errs, fmt.Errorf("%s: UnitPrice '%s' is not a valid decimal: %v", lineIdentifier, line.UnitPrice, err))
			}
			if !isValidDate(line.TaxPointDate) {
				errs = append(errs, fmt.Errorf("%s: TaxPointDate '%s' is invalid", lineIdentifier, line.TaxPointDate))
			}
			if line.Description == "" {
				errs = append(errs, fmt.Errorf("%s: Description (line level) is required", lineIdentifier))
			}

			// Tax Validation (Line.Tax) - if provided
			if line.Tax != nil {
				tax := line.Tax
				validTaxTypes := map[string]bool{"IVA": true, "IS": true, "NS": true}
				if _, ok := validTaxTypes[tax.TaxType]; !ok {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxType '%s' is invalid", lineIdentifier, tax.TaxType))
				}
				if !countryCodeRegex.MatchString(tax.TaxCountryRegion) {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxCountryRegion '%s' is not a valid country code", lineIdentifier, tax.TaxCountryRegion))
				}
				if tax.TaxCode == "" {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxCode is required", lineIdentifier))
				}

				// Copied from SalesInvoices validation, as requirements are same for Line Tax
				hasTaxPercentage := tax.TaxPercentage != "" && tax.TaxPercentage != "0"
				hasTaxAmount := tax.TaxAmount != "" && tax.TaxAmount != "0"

				if tax.TaxPercentage != "" {
					if _, err := parseDecimal(tax.TaxPercentage); err != nil {
						errs = append(errs, fmt.Errorf("%s: Tax.TaxPercentage '%s' is not a valid decimal: %v", lineIdentifier, tax.TaxPercentage, err))
						hasTaxPercentage = false
					}
				}
				if tax.TaxAmount != "" {
					 if _, err := parseDecimal(tax.TaxAmount); err != nil {
						errs = append(errs, fmt.Errorf("%s: Tax.TaxAmount '%s' is not a valid decimal: %v", lineIdentifier, tax.TaxAmount, err))
						hasTaxAmount = false
					}
				}

				if tax.TaxType == "IVA" && !hasTaxPercentage && !hasTaxAmount {
                     errs = append(errs, fmt.Errorf("%s: For TaxType IVA, either TaxPercentage or TaxAmount must be provided and valid", lineIdentifier))
                } else if tax.TaxType == "IVA" && tax.TaxPercentage == "" && tax.TaxAmount == "" {
					errs = append(errs, fmt.Errorf("%s: For TaxType IVA, either TaxPercentage or TaxAmount must be provided", lineIdentifier))
				}
			}
		} // End Line Validation

		// DocumentTotals Validation (similar to SalesInvoices.Invoice.DocumentTotals)
		if workDoc.DocumentTotals == nil {
			errs = append(errs, fmt.Errorf("%s: DocumentTotals is required", wdIdentifier))
		} else {
			dt := workDoc.DocumentTotals
			if _, err := parseDecimal(dt.TaxPayable); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.TaxPayable '%s' is not a valid decimal: %v", wdIdentifier, dt.TaxPayable, err))
			}
			if _, err := parseDecimal(dt.NetTotal); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.NetTotal '%s' is not a valid decimal: %v", wdIdentifier, dt.NetTotal, err))
			}
			if _, err := parseDecimal(dt.GrossTotal); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.GrossTotal '%s' is not a valid decimal: %v", wdIdentifier, dt.GrossTotal, err))
			}

			if dt.Currency != nil && dt.Currency.CurrencyCode != "" {
				if !currencyCodeRegex.MatchString(dt.Currency.CurrencyCode) {
					errs = append(errs, fmt.Errorf("%s: DocumentTotals.Currency.CurrencyCode '%s' is invalid", wdIdentifier, dt.Currency.CurrencyCode))
				}
				if _, err := parseDecimal(dt.Currency.CurrencyAmount); err != nil {
					errs = append(errs, fmt.Errorf("%s: DocumentTotals.Currency.CurrencyAmount '%s' is not a valid decimal: %v", wdIdentifier, dt.Currency.CurrencyAmount, err))
				}
			}
		}
	}
	return errs
}

// ValidateMovementOfGoods validates the MovementOfGoods section.
func ValidateMovementOfGoods(
	mog *saft.MovementOfGoods,
	productCodes map[string]bool, // Map of valid ProductCodes
	customerIDs map[string]bool,  // Map of valid CustomerIDs
	supplierIDs map[string]bool,  // Map of valid SupplierIDs
	// taxTableEntries map[string]saft.TaxTableEntry, // For deeper tax validation if needed for lines
) []error {
	var errs []error

	if mog == nil {
		errs = append(errs, errors.New("MovementOfGoods is nil"))
		return errs
	}

	actualNumberOfMovementLines := 0
	calculatedTotalQuantityIssued := 0.0

	for _, sm := range mog.StockMovement {
		actualNumberOfMovementLines += len(sm.Line)
		for _, line := range sm.Line {
			quantity, err := parseDecimal(line.Quantity)
			if err == nil {
				calculatedTotalQuantityIssued += quantity
			}
		}
	}

	numLinesDeclared, err := strconv.Atoi(mog.NumberOfMovementLines)
	if err != nil || numLinesDeclared != actualNumberOfMovementLines {
		errs = append(errs, fmt.Errorf("mismatch in NumberOfMovementLines: declared %s, actual %d", mog.NumberOfMovementLines, actualNumberOfMovementLines))
	}

	totalQuantityIssuedDeclared, errQty := parseDecimal(mog.TotalQuantityIssued)
	if errQty != nil {
		errs = append(errs, fmt.Errorf("invalid TotalQuantityIssued format: %s", mog.TotalQuantityIssued))
	} else if totalQuantityIssuedDeclared != calculatedTotalQuantityIssued { // Basic float comparison
		errs = append(errs, fmt.Errorf("mismatch in TotalQuantityIssued: declared %f, calculated %f", totalQuantityIssuedDeclared, calculatedTotalQuantityIssued))
	}

	for smIdx, sm := range mog.StockMovement {
		smIdentifier := fmt.Sprintf("StockMovement[%d DocNum:%s]", smIdx, sm.DocumentNumber)

		if sm.DocumentNumber == "" {
			errs = append(errs, fmt.Errorf("%s: DocumentNumber is required", smIdentifier))
		}
		if sm.ATCUD == "" {
			errs = append(errs, fmt.Errorf("%s: ATCUD is required", smIdentifier))
		}

		// DocumentStatus Validation
		if sm.DocumentStatus == nil {
			errs = append(errs, fmt.Errorf("%s: DocumentStatus is required", smIdentifier))
		} else {
			ds := sm.DocumentStatus
			validMovementStatus := map[string]bool{"N": true, "A": true, "F": true, "T": true, "R": true}
			if _, ok := validMovementStatus[ds.MovementStatus]; !ok {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.MovementStatus '%s' is invalid", smIdentifier, ds.MovementStatus))
			}
			if !isValidDateTime(ds.MovementStatusDate) {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.MovementStatusDate '%s' is invalid", smIdentifier, ds.MovementStatusDate))
			}
			if ds.SourceID == "" {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.SourceID is required", smIdentifier))
			}
			validSourceBilling := map[string]bool{"P": true, "I": true, "M": true} // Reused from SalesInvoices
			if _, ok := validSourceBilling[ds.SourceBilling]; !ok {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.SourceBilling '%s' is invalid", smIdentifier, ds.SourceBilling))
			}
		}

		if sm.Hash == "" {
			errs = append(errs, fmt.Errorf("%s: Hash is required", smIdentifier))
		}
		if sm.HashControl == "" {
			errs = append(errs, fmt.Errorf("%s: HashControl is required", smIdentifier))
		}

		if _, ok := isValidInteger(sm.Period, false, false); !ok {
			errs = append(errs, fmt.Errorf("%s: Period '%s' is not a valid positive integer", smIdentifier, sm.Period))
		}
		if !isValidDate(sm.MovementDate) {
			errs = append(errs, fmt.Errorf("%s: MovementDate '%s' is invalid", smIdentifier, sm.MovementDate))
		}

		validMovementTypes := map[string]bool{
			"GR": true, "GT": true, "GA": true, "GC": true, "GD": true, "CM": true, "DM": true,
		}
		if _, ok := validMovementTypes[sm.MovementType]; !ok {
			errs = append(errs, fmt.Errorf("%s: MovementType '%s' is invalid", smIdentifier, sm.MovementType))
		}

		if !isValidDateTime(sm.SystemEntryDate) {
			errs = append(errs, fmt.Errorf("%s: SystemEntryDate '%s' is invalid", smIdentifier, sm.SystemEntryDate))
		}
		if sm.SourceID == "" { // SourceID at StockMovement level
			errs = append(errs, fmt.Errorf("%s: SourceID (at StockMovement level) is required", smIdentifier))
		}

		if !isValidDateTime(sm.MovementStartTime) {
			errs = append(errs, fmt.Errorf("%s: MovementStartTime '%s' is invalid", smIdentifier, sm.MovementStartTime))
		}
		if sm.ATDocCodeID != "" {
			// Just check not empty if provided. Specific format validation might be needed.
		} else if sm.ATDocCodeID == "" && (sm.MovementType == "GT" || sm.MovementType == "GA" || sm.MovementType == "GR" ) { // Example: some types might require ATDocCodeID
			// errs = append(errs, fmt.Errorf("%s: ATDocCodeID is required for MovementType %s", smIdentifier, sm.MovementType))
			// This rule ("if provided") is simple. More complex rules (e.g. required for specific MovementTypes) are not in scope here.
		}


		// CustomerID / SupplierID validation (if applicable)
		if sm.CustomerID != "" {
			if _, ok := customerIDs[sm.CustomerID]; !ok {
				errs = append(errs, fmt.Errorf("%s: CustomerID '%s' is invalid or not found", smIdentifier, sm.CustomerID))
			}
		}
		if sm.SupplierID != "" {
			if _, ok := supplierIDs[sm.SupplierID]; !ok {
				errs = append(errs, fmt.Errorf("%s: SupplierID '%s' is invalid or not found", smIdentifier, sm.SupplierID))
			}
		}


		// Line Validation
		for lineIdx, line := range sm.Line {
			lineIdentifier := fmt.Sprintf("%s Line[%d]", smIdentifier, lineIdx+1)

			ln, ok := isValidInteger(line.LineNumber, false, false)
			if !ok || ln != lineIdx+1 {
				errs = append(errs, fmt.Errorf("%s: LineNumber '%s' is invalid or out of sequence (expected %d)", lineIdentifier, line.LineNumber, lineIdx+1))
			}

			if line.ProductCode == "" {
				errs = append(errs, fmt.Errorf("%s: ProductCode is required", lineIdentifier))
			} else if _, ok := productCodes[line.ProductCode]; !ok {
				errs = append(errs, fmt.Errorf("%s: ProductCode '%s' is invalid or not found", lineIdentifier, line.ProductCode))
			}
			if line.ProductDescription == "" {
				errs = append(errs, fmt.Errorf("%s: ProductDescription is required", lineIdentifier))
			}

			if _, err := parseDecimal(line.Quantity); err != nil {
				errs = append(errs, fmt.Errorf("%s: Quantity '%s' is not a valid decimal: %v", lineIdentifier, line.Quantity, err))
			}
			if line.UnitOfMeasure == "" {
				errs = append(errs, fmt.Errorf("%s: UnitOfMeasure is required", lineIdentifier))
			}
			if _, err := parseDecimal(line.UnitPrice); err != nil { // UnitPrice might be 0 for non-billed movements
				errs = append(errs, fmt.Errorf("%s: UnitPrice '%s' is not a valid decimal: %v", lineIdentifier, line.UnitPrice, err))
			}
			if line.Description == "" {
				errs = append(errs, fmt.Errorf("%s: Description (line level) is required", lineIdentifier))
			}

			// Tax Validation (Line.Tax) - if provided
			if line.Tax != nil {
				tax := line.Tax
				validTaxTypes := map[string]bool{"IVA": true, "IS": true, "NS": true}
				if _, ok := validTaxTypes[tax.TaxType]; !ok {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxType '%s' is invalid", lineIdentifier, tax.TaxType))
				}
				if !countryCodeRegex.MatchString(tax.TaxCountryRegion) {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxCountryRegion '%s' is not a valid country code", lineIdentifier, tax.TaxCountryRegion))
				}
				if tax.TaxCode == "" {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxCode is required", lineIdentifier))
				}
				if tax.TaxPercentage != "" { // Only TaxPercentage listed in requirements for MovementOfGoods Line Tax
					if _, err := parseDecimal(tax.TaxPercentage); err != nil {
						errs = append(errs, fmt.Errorf("%s: Tax.TaxPercentage '%s' is not a valid decimal: %v", lineIdentifier, tax.TaxPercentage, err))
					}
				} else if tax.TaxType == "IVA" { // If IVA, percentage is usually expected unless exempt via TaxCode
					// This could be more nuanced: some IVA codes imply 0%.
					// For now, if TaxType is IVA and TaxPercentage is empty, it's suspicious.
					 errs = append(errs, fmt.Errorf("%s: Tax.TaxPercentage is required for TaxType IVA", lineIdentifier))
				}
			}
		} // End Line Validation

		// DocumentTotals Validation
		if sm.DocumentTotals == nil {
			errs = append(errs, fmt.Errorf("%s: DocumentTotals is required", smIdentifier))
		} else {
			dt := sm.DocumentTotals
			if _, err := parseDecimal(dt.TaxPayable); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.TaxPayable '%s' is not a valid decimal: %v", smIdentifier, dt.TaxPayable, err))
			}
			if _, err := parseDecimal(dt.NetTotal); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.NetTotal '%s' is not a valid decimal: %v", smIdentifier, dt.NetTotal, err))
			}
			if _, err := parseDecimal(dt.GrossTotal); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.GrossTotal '%s' is not a valid decimal: %v", smIdentifier, dt.GrossTotal, err))
			}

			if dt.Currency != nil && dt.Currency.CurrencyCode != "" {
				if !currencyCodeRegex.MatchString(dt.Currency.CurrencyCode) {
					errs = append(errs, fmt.Errorf("%s: DocumentTotals.Currency.CurrencyCode '%s' is invalid", smIdentifier, dt.Currency.CurrencyCode))
				}
				if _, err := parseDecimal(dt.Currency.CurrencyAmount); err != nil {
					errs = append(errs, fmt.Errorf("%s: DocumentTotals.Currency.CurrencyAmount '%s' is not a valid decimal: %v", smIdentifier, dt.Currency.CurrencyAmount, err))
				}
			}
		}
	}
	return errs
}
