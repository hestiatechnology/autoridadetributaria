package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/saft"
)

// Copied helper constants and functions (assuming not in a shared package for this context)
const (
	saftDateFormat     = "2006-01-02"
	saftDateTimeFormat = "2006-01-02T15:04:05"
)

var (
	countryCodeRegexPayments  = regexp.MustCompile(`^[A-Z]{2}$`)
	currencyCodeRegexPayments = regexp.MustCompile(`^[A-Z]{3}$`)
)

func isValidDatePayments(dateStr string) bool {
	if dateStr == "" { return false }
	_, err := time.Parse(saftDateFormat, dateStr)
	return err == nil
}

func isValidDateTimePayments(dateTimeStr string) bool {
	if dateTimeStr == "" { return false }
	_, err := time.Parse(saftDateTimeFormat, dateTimeStr)
	return err == nil
}

func parseDecimalPayments(s string) (float64, error) {
	if s == "" { return 0, errors.New("amount string is empty") }
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

func isValidIntegerPayments(s string, allowZero bool, allowNegative bool) (int, bool) {
	if s == "" { return 0, false }
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil { return 0, false }
	if !allowZero && i == 0 { return 0, false }
	if !allowNegative && i < 0 { return 0, false }
	return i, true
}

// ValidatePayments validates the Payments section.
func ValidatePayments(
	p *saft.Payments,
	customerIDs map[string]bool, // Map of valid CustomerIDs
	// Map of valid invoice numbers to their dates for SourceDocumentID validation.
	// Key: OriginatingON (InvoiceNo), Value: InvoiceDate (as time.Time for accurate comparison)
	validInvoiceReferences map[string]time.Time,
) []error {
	var errs []error

	if p == nil {
		errs = append(errs, errors.New("Payments section is nil"))
		return errs
	}

	actualNumberOfEntries := len(p.Payment)
	calculatedTotalDebit := 0.0
	calculatedTotalCredit := 0.0

	for _, paymentDoc := range p.Payment {
		if paymentDoc.DocumentTotals != nil {
			// For payments (RC), GrossTotal typically debits cash and credits AR.
			// So, TotalDebit and TotalCredit in the summary might reflect this.
			// This interpretation can vary. Using GrossTotal for both as a common summary approach.
			grossTotal, err := parseDecimalPayments(paymentDoc.DocumentTotals.GrossTotal)
			if err == nil {
				calculatedTotalCredit += grossTotal
				calculatedTotalDebit += grossTotal
			}
		}
	}

	numEntriesDeclared, err := strconv.Atoi(p.NumberOfEntries)
	if err != nil || numEntriesDeclared != actualNumberOfEntries {
		errs = append(errs, fmt.Errorf("mismatch in NumberOfEntries: declared %s, actual %d", p.NumberOfEntries, actualNumberOfEntries))
	}

	totalDebitDeclared, errDb := parseDecimalPayments(p.TotalDebit)
	if errDb != nil {
		errs = append(errs, fmt.Errorf("invalid TotalDebit format: %s", p.TotalDebit))
	} else if totalDebitDeclared != calculatedTotalDebit {
		errs = append(errs, fmt.Errorf("mismatch in TotalDebit: declared %.2f, calculated %.2f", totalDebitDeclared, calculatedTotalDebit))
	}

	totalCreditDeclared, errCr := parseDecimalPayments(p.TotalCredit)
	if errCr != nil {
		errs = append(errs, fmt.Errorf("invalid TotalCredit format: %s", p.TotalCredit))
	} else if totalCreditDeclared != calculatedTotalCredit {
		errs = append(errs, fmt.Errorf("mismatch in TotalCredit: declared %.2f, calculated %.2f", totalCreditDeclared, calculatedTotalCredit))
	}

	for pIdx, payment := range p.Payment {
		pIdentifier := fmt.Sprintf("Payment[%d RefNo:%s]", pIdx, payment.PaymentRefNo)

		if payment.PaymentRefNo == "" {
			errs = append(errs, fmt.Errorf("%s: PaymentRefNo is required", pIdentifier))
		}
		if payment.ATCUD == "" { // ATCUD is usually mandatory for documents subject to certification
			errs = append(errs, fmt.Errorf("%s: ATCUD is required", pIdentifier))
		}

		if payment.Period != "" { // Optional field
			if _, ok := isValidIntegerPayments(payment.Period, false, false); !ok {
				errs = append(errs, fmt.Errorf("%s: Period '%s' is not a valid positive integer", pIdentifier, payment.Period))
			}
		}
		if payment.TransactionID != "" { // Optional field
			// No specific format, just not empty if provided.
		}
		if !isValidDatePayments(payment.TransactionDate) {
			errs = append(errs, fmt.Errorf("%s: TransactionDate '%s' is invalid", pIdentifier, payment.TransactionDate))
		}

		validPaymentTypes := map[string]bool{"RC": true, "RG": true}
		if _, ok := validPaymentTypes[payment.PaymentType]; !ok {
			errs = append(errs, fmt.Errorf("%s: PaymentType '%s' is invalid", pIdentifier, payment.PaymentType))
		}
		if payment.Description != "" { // Optional field
			// No specific format, just not empty if provided.
		}
		if payment.SystemID != "" { // Optional field
			// No specific format, just not empty if provided.
		}

		// DocumentStatus Validation
		if payment.DocumentStatus == nil {
			errs = append(errs, fmt.Errorf("%s: DocumentStatus is required", pIdentifier))
		} else {
			ds := payment.DocumentStatus
			validPaymentStatus := map[string]bool{"N": true, "A": true, "P": true}
			if _, ok := validPaymentStatus[ds.PaymentStatus]; !ok {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.PaymentStatus '%s' is invalid", pIdentifier, ds.PaymentStatus))
			}
			if !isValidDateTimePayments(ds.PaymentStatusDate) {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.PaymentStatusDate '%s' is invalid", pIdentifier, ds.PaymentStatusDate))
			}
			if ds.SourceID == "" {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.SourceID is required", pIdentifier))
			}
			validSourcePayment := map[string]bool{"P": true, "I": true, "M": true}
			if _, ok := validSourcePayment[ds.SourcePayment]; !ok {
				errs = append(errs, fmt.Errorf("%s: DocumentStatus.SourcePayment '%s' is invalid", pIdentifier, ds.SourcePayment))
			}
		}

		// PaymentMethod Validation
		if payment.PaymentMethod == nil || len(payment.PaymentMethod) == 0 {
             errs = append(errs, fmt.Errorf("%s: At least one PaymentMethod is required", pIdentifier))
        }
		for pmIdx, pm := range payment.PaymentMethod {
			pmIdentifier := fmt.Sprintf("%s PaymentMethod[%d]", pIdentifier, pmIdx)
			validMechanisms := map[string]bool{
				"CC": true, "CD": true, "CH": true, "CI": true, "CO": true, "CS": true, "DE": true,
				"LC": true, "MB": true, "NU": true, "OU": true, "PR": true, "TB": true, "TR": true,
			}
			if pm.PaymentMechanism != "" { // Optional field
				if _, ok := validMechanisms[pm.PaymentMechanism]; !ok {
					errs = append(errs, fmt.Errorf("%s: PaymentMechanism '%s' is invalid", pmIdentifier, pm.PaymentMechanism))
				}
			}
			if _, err := parseDecimalPayments(pm.PaymentAmount); err != nil {
				errs = append(errs, fmt.Errorf("%s: PaymentAmount '%s' is not a valid decimal: %v", pmIdentifier, pm.PaymentAmount, err))
			}
			if !isValidDatePayments(pm.PaymentDate) {
				errs = append(errs, fmt.Errorf("%s: PaymentDate '%s' is invalid", pmIdentifier, pm.PaymentDate))
			}
		}

		if payment.SourceID == "" { // SourceID at Payment level
			errs = append(errs, fmt.Errorf("%s: SourceID (at Payment level) is required", pIdentifier))
		}
		if !isValidDateTimePayments(payment.SystemEntryDate) {
			errs = append(errs, fmt.Errorf("%s: SystemEntryDate '%s' is invalid", pIdentifier, payment.SystemEntryDate))
		}
		if payment.CustomerID == "" {
			errs = append(errs, fmt.Errorf("%s: CustomerID is required", pIdentifier))
		} else if _, ok := customerIDs[payment.CustomerID]; !ok {
			errs = append(errs, fmt.Errorf("%s: CustomerID '%s' is invalid or not found", pIdentifier, payment.CustomerID))
		}

		// Line Validation
		if payment.Line == nil || len(payment.Line) == 0 {
             errs = append(errs, fmt.Errorf("%s: At least one Line is required", pIdentifier))
        }
		for lineIdx, line := range payment.Line {
			lineIdentifier := fmt.Sprintf("%s Line[%d]", pIdentifier, lineIdx+1)

			ln, ok := isValidIntegerPayments(line.LineNumber, false, false)
			if !ok || ln != lineIdx+1 {
				errs = append(errs, fmt.Errorf("%s: LineNumber '%s' is invalid or out of sequence (expected %d)", lineIdentifier, line.LineNumber, lineIdx+1))
			}

			// SourceDocumentID Validation (links to invoices)
			if line.SourceDocumentID == nil || len(line.SourceDocumentID) == 0 {
                 errs = append(errs, fmt.Errorf("%s: At least one SourceDocumentID is required per Line", lineIdentifier))
            }
			for sdIdx, sd := range line.SourceDocumentID {
				sdIdentifier := fmt.Sprintf("%s SourceDocumentID[%d]", lineIdentifier, sdIdx)
				if sd.OriginatingON == "" {
					errs = append(errs, fmt.Errorf("%s: OriginatingON is required", sdIdentifier))
				}
				if !isValidDatePayments(sd.InvoiceDate) { // This is InvoiceDate of the original document
					errs = append(errs, fmt.Errorf("%s: InvoiceDate '%s' is invalid", sdIdentifier, sd.InvoiceDate))
				}
				// Deeper validation against validInvoiceReferences map
				if sd.OriginatingON != "" && validInvoiceReferences != nil {
					expectedInvoiceDate, refOk := validInvoiceReferences[sd.OriginatingON]
					if !refOk {
						errs = append(errs, fmt.Errorf("%s: OriginatingON '%s' not found in valid invoice references", sdIdentifier, sd.OriginatingON))
					} else {
						parsedInvoiceDate, _ := time.Parse(saftDateFormat, sd.InvoiceDate)
						if !parsedInvoiceDate.Equal(expectedInvoiceDate) {
							errs = append(errs, fmt.Errorf("%s: InvoiceDate '%s' does not match registered date '%s' for OriginatingON '%s'", sdIdentifier, sd.InvoiceDate, expectedInvoiceDate.Format(saftDateFormat), sd.OriginatingON))
						}
					}
				}
			}

			if line.SettlementAmount != "" { // Optional
				if _, err := parseDecimalPayments(line.SettlementAmount); err != nil {
					errs = append(errs, fmt.Errorf("%s: SettlementAmount '%s' is not a valid decimal: %v", lineIdentifier, line.SettlementAmount, err))
				}
			}

			hasDebit := line.DebitAmount != ""
			hasCredit := line.CreditAmount != ""

			if hasDebit {
				if _, err := parseDecimalPayments(line.DebitAmount); err != nil {
					errs = append(errs, fmt.Errorf("%s: DebitAmount '%s' is not a valid decimal: %v", lineIdentifier, line.DebitAmount, err))
					hasDebit = false // Consider it not validly provided
				}
			}
			if hasCredit {
				if _, err := parseDecimalPayments(line.CreditAmount); err != nil {
					errs = append(errs, fmt.Errorf("%s: CreditAmount '%s' is not a valid decimal: %v", lineIdentifier, line.CreditAmount, err))
					hasCredit = false // Consider it not validly provided
				}
			}

			if !hasDebit && !hasCredit {
				errs = append(errs, fmt.Errorf("%s: Either DebitAmount or CreditAmount must be provided and valid", lineIdentifier))
			}
			// Note: SAFT rules might restrict having both Debit and Credit amount on the same payment line. Not checked here.

			// Tax Validation (Line.Tax) - if provided
			if line.Tax != nil {
				tax := line.Tax
				validTaxTypes := map[string]bool{"IVA": true, "IS": true, "NS": true}
				if _, ok := validTaxTypes[tax.TaxType]; !ok {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxType '%s' is invalid", lineIdentifier, tax.TaxType))
				}
				if !countryCodeRegexPayments.MatchString(tax.TaxCountryRegion) {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxCountryRegion '%s' is not a valid country code", lineIdentifier, tax.TaxCountryRegion))
				}
				if tax.TaxCode == "" {
					errs = append(errs, fmt.Errorf("%s: Tax.TaxCode is required", lineIdentifier))
				}

				hasTaxPercentage := tax.TaxPercentage != "" && tax.TaxPercentage != "0"
				hasTaxAmount := tax.TaxAmount != "" && tax.TaxAmount != "0"

				if tax.TaxPercentage != "" {
					if _, err := parseDecimalPayments(tax.TaxPercentage); err != nil {
						errs = append(errs, fmt.Errorf("%s: Tax.TaxPercentage '%s' is not a valid decimal: %v", lineIdentifier, tax.TaxPercentage, err))
						hasTaxPercentage = false
					}
				}
				if tax.TaxAmount != "" {
					 if _, err := parseDecimalPayments(tax.TaxAmount); err != nil {
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

		// DocumentTotals Validation
		if payment.DocumentTotals == nil {
			errs = append(errs, fmt.Errorf("%s: DocumentTotals is required", pIdentifier))
		} else {
			dt := payment.DocumentTotals
			if _, err := parseDecimalPayments(dt.TaxPayable); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.TaxPayable '%s' is not a valid decimal: %v", pIdentifier, dt.TaxPayable, err))
			}
			if _, err := parseDecimalPayments(dt.NetTotal); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.NetTotal '%s' is not a valid decimal: %v", pIdentifier, dt.NetTotal, err))
			}
			if _, err := parseDecimalPayments(dt.GrossTotal); err != nil {
				errs = append(errs, fmt.Errorf("%s: DocumentTotals.GrossTotal '%s' is not a valid decimal: %v", pIdentifier, dt.GrossTotal, err))
			}

			if dt.Settlement != nil && dt.Settlement.SettlementAmount != "" { // If Settlement block exists and amount is provided
				if _, err := parseDecimalPayments(dt.Settlement.SettlementAmount); err != nil {
					errs = append(errs, fmt.Errorf("%s: DocumentTotals.Settlement.SettlementAmount '%s' is not a valid decimal: %v", pIdentifier, dt.Settlement.SettlementAmount, err))
				}
			}

			if dt.Currency != nil && dt.Currency.CurrencyCode != "" {
				if !currencyCodeRegexPayments.MatchString(dt.Currency.CurrencyCode) {
					errs = append(errs, fmt.Errorf("%s: DocumentTotals.Currency.CurrencyCode '%s' is invalid", pIdentifier, dt.Currency.CurrencyCode))
				}
				if _, err := parseDecimalPayments(dt.Currency.CurrencyAmount); err != nil {
					errs = append(errs, fmt.Errorf("%s: DocumentTotals.Currency.CurrencyAmount '%s' is not a valid decimal: %v", pIdentifier, dt.Currency.CurrencyAmount, err))
				}
			}
		}
	}
	return errs
}
