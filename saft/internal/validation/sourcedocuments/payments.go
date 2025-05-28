package validation

import (
	"fmt"
	"slices"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/common"
	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/shopspring/decimal"
)

func ValidatePayments(a *saft.AuditFile) error {

	if a.SourceDocuments == nil && a.SourceDocuments.Payments == nil {
		return nil
	}

	// Calculate the number of registered payments, total debits and total credits
	numPayments := uint64(0)
	var totalDebit, totalCredit decimal.Decimal
	for _, payment := range a.SourceDocuments.Payments.Payment {
		if len(payment.Line) == 0 {
			return fmt.Errorf("saft: missing Payment.Line")
		}

		numPayments++
		for _, line := range payment.Line {
			// Can't have both debit and credit
			if line.DebitAmount != nil && line.CreditAmount != nil {
				return fmt.Errorf("saft: invalid Payment.Line: both DebitAmount and CreditAmount present")
			}

			if line.DebitAmount != nil {
				totalDebit = totalDebit.Add((*line.DebitAmount).Decimal)
			}

			if line.CreditAmount != nil {
				totalCredit = totalCredit.Add(line.CreditAmount.Decimal)
			}
		}
	}

	// Check if the total debit and total credit are the same
	if a.SourceDocuments.Payments.NumberOfEntries != numPayments {
		return fmt.Errorf("saft: invalid NumberOfEntries: %d != calculated %d", a.SourceDocuments.Payments.NumberOfEntries, numPayments)
	}

	if a.SourceDocuments.Payments.TotalCredit.Cmp(totalCredit) != 0 {
		return fmt.Errorf("saft: invalid TotalCredit: %s != calculated %s", a.SourceDocuments.Payments.TotalCredit, totalCredit)
	}

	if a.SourceDocuments.Payments.TotalDebit.Cmp(totalDebit) != 0 {
		return fmt.Errorf("saft: invalid TotalDebit: %s != calculated %s", a.SourceDocuments.Payments.TotalDebit, totalDebit)
	}

	for _, payment := range a.SourceDocuments.Payments.Payment {
		if payment.PaymentRefNo == "" {
			return fmt.Errorf("saft: missing Payment.PaymentRefNo")
		}

		if payment.Atcud == "" {
			return fmt.Errorf("saft: missing Payment.Atcud")
		}

		// TODO: Validate ATCUD?

		// Transaction Id validation is done in the constraints

		if payment.TransactionDate == (saft.SafdateType{}) {
			return fmt.Errorf("saft: missing Payment.TransactionDate")
		}

		if payment.TransactionDate.After(time.Now()) {
			return fmt.Errorf("saft: invalid Payment.TransactionDate")
		}

		if payment.PaymentType != saft.SaftptpaymentTypeRC && payment.PaymentType != saft.SaftptpaymentTypeRG {
			return fmt.Errorf("saft: invalid Payment.PaymentType")
		}

		if payment.DocumentStatus == (saft.PaymentDocumentStatus{}) {
			return fmt.Errorf("saft: missing Payment.DocumentStatus")
		}

		if payment.DocumentStatus.PaymentStatus != saft.PaymentStatusNormal && payment.DocumentStatus.PaymentStatus != saft.PaymentStatusCancelled {
			return fmt.Errorf("saft: invalid Payment.DocumentStatus.PaymentStatus")
		}

		if payment.DocumentStatus.PaymentStatusDate == (saft.SafdateTimeType{}) {
			return fmt.Errorf("saft: missing Payment.DocumentStatus.PaymentStatusDate")
		}

		if time.Time(payment.DocumentStatus.PaymentStatusDate).After(time.Now()) {
			return fmt.Errorf("saft: invalid Payment.DocumentStatus.PaymentStatusDate")
		}

		if payment.DocumentStatus.SourceId == "" {
			return fmt.Errorf("saft: missing Payment.DocumentStatus.SourceId")
		}

		if payment.DocumentStatus.SourcePayment != saft.SaftptsourcePaymentP &&
			payment.DocumentStatus.SourcePayment != saft.SaftptsourcePaymentI &&
			payment.DocumentStatus.SourcePayment != saft.SaftptsourcePaymentM {
			return fmt.Errorf("saft: invalid Payment.DocumentStatus.SourcePayment")
		}

		if len(payment.PaymentMethod) == 0 {
			return fmt.Errorf("saft: missing Payment.PaymentMethod")
		}

		taxPayable := decimal.NewFromInt(0)
		for _, method := range payment.PaymentMethod {
			if method.PaymentMechanism != nil {
				switch *method.PaymentMechanism {
				case saft.PaymentMechanismCC, saft.PaymentMechanismCD, saft.PaymentMechanismCH, saft.PaymentMechanismCI, saft.PaymentMechanismCO, saft.PaymentMechanismCS, saft.PaymentMechanismDE, saft.PaymentMechanismLC, saft.PaymentMechanismMB, saft.PaymentMechanismNU, saft.PaymentMechanismOU, saft.PaymentMechanismPR, saft.PaymentMechanismTB, saft.PaymentMechanismTR:
					// Ignore
				default:
					return fmt.Errorf("saft: invalid PaymentMethod.PaymentMechanism: %s", *method.PaymentMechanism)
				}
			}

			if method.PaymentAmount.LessThan(decimal.NewFromInt(0)) {
				return fmt.Errorf("saft: invalid PaymentMethod.PaymentAmount: %s", method.PaymentAmount)
			}

			if method.PaymentDate == (saft.SafdateType{}) {
				return fmt.Errorf("saft: missing PaymentMethod.PaymentDate")
			}

			if method.PaymentDate.After(time.Now()) {
				return fmt.Errorf("saft: invalid PaymentMethod.PaymentDate")
			}

			if payment.SourceId == "" {
				return fmt.Errorf("saft: missing Payment.SourceId")
			}

			if payment.CustomerId == "" {
				return fmt.Errorf("saft: missing Payment.CustomerId")
			}

			for _, line := range payment.Line {
				// LineNumber can be 0 I guess

				if len(line.SourceDocumentId) == 0 {
					return fmt.Errorf("saft: missing PaymentLine.SourceDocumentId")
				}
				for _, sourceDocument := range line.SourceDocumentId {
					if sourceDocument.OriginatingOn == "" {
						return fmt.Errorf("saft: missing PaymentLine.SourceDocumentId.OriginatingOn")
					}
				}

				if line.DebitAmount != nil {
					if (*line.DebitAmount).LessThan(decimal.NewFromInt(0)) {
						return fmt.Errorf("saft: invalid PaymentLine.DebitAmount: %s", line.DebitAmount)
					}
				}

				if line.CreditAmount != nil {
					if (*line.CreditAmount).LessThan(decimal.NewFromInt(0)) {
						return fmt.Errorf("saft: invalid PaymentLine.CreditAmount: %s", line.CreditAmount)
					}
				}

				// If PaymentType is RC, then Tax must be present
				if payment.PaymentType == saft.SaftptpaymentTypeRC && line.Tax == nil {
					return fmt.Errorf("saft: missing mandatory PaymentLine.Tax due to PaymentType = RC")
				}

				// TODO: check later, field 4.4.4.14.6 of portaria
				// If PaymentType is RG, then Tax must not be present
				//if payment.PaymentType == SaftptpaymentTypeRG && line.Tax.TaxType ==  {
				//	return fmt.Errorf("saft: invalid PaymentLine.Tax")
				//}

				if line.Tax != nil {
					if line.Tax.TaxType != saft.TaxTypeIVA && line.Tax.TaxType != saft.TaxTypeIS && line.Tax.TaxType != saft.TaxTypeNS {
						return fmt.Errorf("saft: invalid PaymentLine.Tax.TaxType: %s", line.Tax.TaxType)
					}

					// Check if its a ISO 3166-1 alpha-2 country code + PT regions
					if !slices.Contains(common.CountryCodesPTRegions, line.Tax.TaxCountryRegion) {
						return fmt.Errorf("saft: invalid PaymentLine.Tax.TaxCountryRegion: %s", line.Tax.TaxCountryRegion)
					}

					// Cant have both TaxPercentage and TaxAmount
					if line.Tax.TaxPercentage != nil && line.Tax.TaxAmount != nil {
						return fmt.Errorf("saft: invalid PaymentLine.Tax: both TaxPercentage and TaxAmount present")
					}

					// TaxPercentage can only be zero if TaxCode is Ise or Na
					if line.Tax.TaxPercentage != nil &&
						(*line.Tax.TaxPercentage).Cmp(decimal.NewFromInt(0)) == 0 &&
						(line.Tax.TaxCode != saft.PaymentTaxCode(saft.TaxCodeIse) && line.Tax.TaxCode != saft.PaymentTaxCode(saft.TaxCodeNa)) {
						return fmt.Errorf("saft: invalid PaymentLine.Tax: TaxPercentage is zero but TaxCode is not Ise or Na")
					}

					// TaxPercentage must be 0 or greater
					if line.Tax.TaxPercentage != nil && (*line.Tax.TaxPercentage).LessThan(decimal.NewFromInt(0)) {
						return fmt.Errorf("saft: invalid PaymentLine.Tax.TaxPercentage: %s", line.Tax.TaxPercentage)
					}

					// Calculate the TaxPayable using the TaxPercentage
					if line.Tax.TaxPercentage != nil {
						taxMultiplier := (*line.Tax.TaxPercentage).Div(decimal.NewFromInt(100)).Add(decimal.NewFromInt(1))
						if line.DebitAmount != nil {
							debitAmount := (*line.DebitAmount)
							taxPayable = taxPayable.Add((debitAmount.Mul(taxMultiplier)).Sub(debitAmount.Decimal))
						}
					}

					// TaxAmount can only appear if TaxType is IS
					if line.Tax.TaxAmount != nil && line.Tax.TaxType != saft.TaxTypeIS {
						return fmt.Errorf("saft: invalid PaymentLine.Tax: TaxAmount present but TaxType is not IS")
					}

					if (*line.Tax.TaxAmount).LessThan(decimal.NewFromInt(0)) {
						return fmt.Errorf("saft: invalid PaymentLine.Tax.TaxAmount: %s", line.Tax.TaxAmount)
					}

					// Add the TaxAmount to the TaxPayable
					if line.Tax.TaxAmount != nil {
						taxPayable = taxPayable.Add(line.Tax.TaxAmount.Decimal)
					}

				}

				//// Cant have Tax and TaxExemptionReason and TaxExemptionCodesaft.
				//if line.Tax != nil && line.TaxExemptionReason != nil && line.TaxExemptionCode != nil {
				//	return errcodes.ErrPaymentLineTaxTaxExemption
				//}

				// TaxExemptionReason and TaxExemptionCode must exist when TaxPercentage/TaxAmount is 0
				if line.Tax != nil && line.Tax.TaxPercentage != nil && (*line.Tax.TaxPercentage).Cmp(decimal.NewFromInt(0)) == 0 {
					if line.TaxExemptionReason == nil || line.TaxExemptionCode == nil {
						return fmt.Errorf("saft: missing PaymentLine.TaxExemptionReason or PaymentLine.TaxExemptionCode")
					}
				}

				// Check if TaxExemption is valid
				if line.TaxExemptionCode != nil && line.TaxExemptionReason != nil {
					for _, exemption := range common.VatExemptionCodes {
						if saft.SafptportugueseTaxExemptionCode(exemption.Code) == *line.TaxExemptionCode && saft.SafptportugueseTaxExemptionReason(exemption.Description) == *line.TaxExemptionReason {
							break
						}
					}
					return fmt.Errorf("saft: invalid PaymentLine.TaxExemptionCode or PaymentLine.TaxExemptionReason")
				}
			}

		}

		if payment.DocumentTotals == (saft.PaymentDocumentTotals{}) {
			return fmt.Errorf("saft: missing Payment.DocumentTotals")
		}

		if payment.DocumentTotals.TaxPayable.Cmp(taxPayable) != 0 {
			return fmt.Errorf("saft: invalid Payment.DocumentTotals.TaxPayable: %s != calculated %s", payment.DocumentTotals.GrossTotal, totalDebit)
		}

		// Calculate NetTotal and GrossTotal
		netTotal := decimal.NewFromInt(0)
		grossTotal := decimal.NewFromInt(0)
		settlementTotal := decimal.NewFromInt(0)
		for _, line := range payment.Line {
			if line.DebitAmount != nil {
				netTotal = netTotal.Add(line.DebitAmount.Decimal)
				//grossTotal = grossTotal.Add(decimal.Decimal(*line.DebitAmount))
			}

			if line.CreditAmount != nil {
				netTotal = netTotal.Sub(line.CreditAmount.Decimal)
				//grossTotal = grossTotal.Sub(decimal.Decimal(*line.CreditAmount))
			}

			//  get the tax amount
			if line.Tax != nil && line.Tax.TaxPercentage != nil {
				taxMultiplier := (*line.Tax.TaxPercentage).Div(decimal.NewFromInt(100)).Add(decimal.NewFromInt(1))
				grossTotal = grossTotal.Add(netTotal.Div(taxMultiplier))
			} else if line.Tax != nil && line.Tax.TaxAmount != nil {
				grossTotal = grossTotal.Add(netTotal.Add(line.Tax.TaxAmount.Decimal))
			}

			if line.SettlementAmount != nil {
				settlementTotal = settlementTotal.Add(line.SettlementAmount.Decimal)
			}
		}

		// Check if the net total and gross total are correct
		if payment.DocumentTotals.NetTotal.Cmp(netTotal) != 0 {
			return fmt.Errorf("saft: invalid Payment.DocumentTotals.NetTotal: %s != calculated %s", payment.DocumentTotals.NetTotal, netTotal)
		}
		if payment.DocumentTotals.GrossTotal.Cmp(grossTotal) != 0 {
			return fmt.Errorf("saft: invalid Payment.DocumentTotals.GrossTotal: %s != calculated %s", payment.DocumentTotals.GrossTotal, grossTotal)
		}

		// Check Settlement
		if settlementTotal.GreaterThan(decimal.Zero) && payment.DocumentTotals.Settlement == nil {
			return fmt.Errorf("saft: missing Payment.DocumentTotals.Settlement when settlementTotal is greater than zero")
		}
		if payment.DocumentTotals.Settlement != nil {
			if payment.DocumentTotals.Settlement.SettlementAmount.Cmp(settlementTotal) != 0 {
				return fmt.Errorf("saft: invalid Payment.DocumentTotals.Settlement.SettlementAmount: %s != calculated %s", payment.DocumentTotals.Settlement.SettlementAmount, settlementTotal)
			}
		}

		// Check Currency
		if payment.DocumentTotals.Currency != nil && a.Header.CurrencyCode == "EUR" {
			return fmt.Errorf("saft: invalid Payment.DocumentTotals.Currency: %s, Header.CurrencyCode must not be EUR", payment.DocumentTotals.Currency.CurrencyCode)
		}

		if payment.DocumentTotals.Currency == nil && a.Header.CurrencyCode != "EUR" {
			return fmt.Errorf("saft: missing Payment.DocumentTotals.Currency when Header.CurrencyCode is not EUR")
		}

		if payment.DocumentTotals.Currency != nil {

		}
	}
	return nil
}
