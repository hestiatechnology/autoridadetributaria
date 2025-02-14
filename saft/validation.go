package saft

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/common"
	"github.com/shopspring/decimal"
)

func (a *AuditFile) Validate() error {

	if err := a.checkCommon(); err != nil {
		return err
	}

	// Check the tests of the AuditFile
	if a.MasterFiles.GeneralLedgerAccounts != nil {
		for _, account := range a.MasterFiles.GeneralLedgerAccounts.Account {
			// Test 1
			if (account.GroupingCategory == "GM" && account.TaxonomyCode == nil) || (account.GroupingCategory != "GM" && account.TaxonomyCode != nil) {
				// logger.DebugLogger.Println("Err: Grouping category GM and taxonomy code must be present together")
				return fmt.Errorf("saft: grouping category GM and taxonomy code must be present together")
			}
			// Test 2
			if (account.GroupingCategory == "GR" && account.GroupingCode != nil) ||
				(account.GroupingCategory == "AR" && account.GroupingCode != nil) ||
				(account.GroupingCategory == "GA" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "AA" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "GM" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "AM" && account.GroupingCode == nil) {
				// logger.DebugLogger.Println("Err: Invalid GroupingCategory and GroupingCode combination: ", account.GroupingCategory, *account.GroupingCode)
				return fmt.Errorf("saft: invalid GroupingCategory and GroupingCode combination: %s %s", account.GroupingCategory, *account.GroupingCode)
			}
		}
	}

	for _, customer := range a.MasterFiles.Customer {
		r := regexp.MustCompile(`([^^]*)`)
		// Check if it doesnt match Desconhecido or the regex
		if customer.AccountId != "Desconhecido" && !r.MatchString(customer.AccountId) {
			return fmt.Errorf("saft: invalid AccountId: %s", customer.AccountId)
		}
	}

	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		if invoice.SpecialRegimes.CashVatschemeIndicator != CashVatschemeIndicatorNo && invoice.SpecialRegimes.CashVatschemeIndicator != CashVatschemeIndicatorYes {
			return fmt.Errorf("saft: invalid CashVatschemeIndicator: %d", invoice.SpecialRegimes.CashVatschemeIndicator)
		}
	}

	if err := a.checkConstraints(); err != nil {
		return err
	}
	return nil
}

// Checks the common fields of the AuditFile
// 1. – Cabeçalho (Header);
// 2.2. – Tabela de clientes (Customer);
// 2.5. – Tabela de impostos (TaxTable); e
// 4.4. – Documentos de recibos emitidos (Payments), quando deva existir.
func (a *AuditFile) checkCommon() error {
	if err := a.checkHeader(); err != nil {
		return err
	}

	if err := a.checkCustomers(); err != nil {
		return err
	}

	// TODO: Check TaxTable

	if err := a.checkPayments(); err != nil {
		return err
	}
	return nil
}

// Check Header mandatory fields
func (a *AuditFile) checkHeader() error {
	// 1.04_01 is the only supported version
	if a.Header.AuditFileVersion != "1.04_01" {
		return fmt.Errorf("saft: unsupported SAFT version: %s", a.Header.AuditFileVersion)
	}

	// CompanyId must be either a NIF (9 numbers) or Conservatória (string) + NIF
	r := regexp.MustCompile(`([0-9]{9})+|([^^]+ [0-9/]+)`)
	if !r.MatchString(a.Header.CompanyId) {
		return fmt.Errorf("saft: invalid Header.CompanyId: %s", a.Header.CompanyId)
	}

	// Check if TaxRegistrationNumber is a NIF (9 numbers)
	taxNumber := strconv.FormatUint(uint64(a.Header.TaxRegistrationNumber), 10)
	ok := common.ValidateNIFPT(taxNumber)
	if !ok {
		return fmt.Errorf("saft: invalid NIF in Header.TaxRegistrationNumber: %d", a.Header.TaxRegistrationNumber)
	}

	switch a.Header.TaxAccountingBasis {
	case SaftAccounting, SaftInvoicingThirdParties, SaftInvoicing, SaftIntegrated, SaftInvoicingParcial, SaftPayments, SaftSelfBilling, SaftTransportDocuments:
		break
	default:
		return fmt.Errorf("saft: invalid Header.TaxAccountingBasis: %s", a.Header.TaxAccountingBasis)
	}

	if a.Header.CompanyName == "" {
		return fmt.Errorf("saft: missing Header.CompanyName")
	}

	// Check if CompanyAddress is empty
	if a.Header.CompanyAddress == (AddressStructure{}) {
		return fmt.Errorf("saft: missing Header.CompanyAddress")
	}

	if a.Header.CompanyAddress.AddressDetail == "" {
		return fmt.Errorf("saft: missing Header.CompanyAddress.AddressDetail")
	}

	if a.Header.CompanyAddress.City == "" {
		return fmt.Errorf("saft: missing Header.CompanyAddress.City")
	}

	if a.Header.CompanyAddress.PostalCode == "" {
		return fmt.Errorf("saft: missing Header.CompanyAddress.PostalCode")
	}

	if a.Header.CompanyAddress.Country != "PT" {
		return fmt.Errorf("saft: invalid Header.CompanyAddress.Country, must be PR, current value: %s", a.Header.CompanyAddress.Country)
	}

	if a.Header.FiscalYear == "" {
		return fmt.Errorf("saft: missing Header.FiscalYear")
	}

	// Check if its a valid year
	year, err := strconv.Atoi(a.Header.FiscalYear)
	if err != nil {
		return fmt.Errorf("saft: invalid Header.FiscalYear: %s", a.Header.FiscalYear)
	}

	// check if its higher than current year
	currentYear := time.Now().Year()
	if year < 2000 && year > currentYear {
		return fmt.Errorf("saft: invalid Header.FiscalYear: %s", a.Header.FiscalYear)

	}

	if a.Header.StartDate == (SafptdateSpan{}) {
		return fmt.Errorf("saft: missing Header.StartDate")
	}

	if a.Header.EndDate == (SafptdateSpan{}) {
		return fmt.Errorf("saft: missing Header.EndDate")
	}

	if time.Time(a.Header.StartDate).After(time.Time(a.Header.EndDate)) {
		return fmt.Errorf("saft: invalid Header.StartDate and Header.EndDate")
	}

	// Currency must be EUR, if there was a another currency
	// it must be EUR+Exchange rate
	if a.Header.CurrencyCode != "EUR" {
		return fmt.Errorf("saft: invalid Header.CurrencyCode: %s", a.Header.CurrencyCode)
	}

	if a.Header.DateCreated == (SafptdateSpan{}) {
		return fmt.Errorf("saft: missing Header.DateCreated")
	}

	// Must be a establishment, else Global
	// Must be Sede if its a integrated accounting file
	if a.Header.TaxEntity == "" {
		return fmt.Errorf("saft: missing Header.TaxEntity")
	}

	// TaxId of the company who produced the software
	if a.Header.ProductCompanyTaxId == "" {
		return fmt.Errorf("saft: missing Header.ProductCompanyTaxId")
	}

	if len(a.Header.ProductCompanyTaxId) != 9 {
		return fmt.Errorf("saft: invalid NIF in Header.ProductCompanyTaxId: %s", a.Header.ProductCompanyTaxId)
	}

	ok = common.ValidateNIFPT(string(a.Header.ProductCompanyTaxId))
	if !ok {
		return fmt.Errorf("saft: invalid NIF in Header.ProductCompanyTaxId: %s", a.Header.ProductCompanyTaxId)
	}

	//SoftwareCertificateNumber is by default 0, which is a valid value

	r = regexp.MustCompile(`[^/]+\/[^/]+`)
	if !r.MatchString(string(a.Header.ProductId)) {
		return fmt.Errorf("saft: invalid Header.ProductId: %s", a.Header.ProductId)
	}

	if a.Header.ProductVersion == "" {
		return fmt.Errorf("saft: missing Header.ProductVersion")
	}

	return nil
}

func (a *AuditFile) checkCustomers() error {
	for _, customer := range a.MasterFiles.Customer {
		if customer.CustomerId == "" {
			return fmt.Errorf("saft: missing Customer.CustomerId")
		}

		if customer.AccountId == "" {
			return fmt.Errorf("saft: missing Customer.AccountId")
		}

		if customer.CustomerTaxId == "" {
			return fmt.Errorf("saft: missing Customer.CustomerTaxId")
		}

		if customer.CompanyName == "" {
			return fmt.Errorf("saft: missing Customer.CompanyName")
		}

		if customer.BillingAddress == (CustomerAddressStructure{}) {
			return fmt.Errorf("saft: missing Customer.BillingAddress")
		}

		if customer.BillingAddress.AddressDetail == "" {
			return fmt.Errorf("saft: missing Customer.BillingAddress.AddressDetail")
		}

		if customer.BillingAddress.City == "" {
			return fmt.Errorf("saft: missing Customer.BillingAddress.City")
		}

		if customer.BillingAddress.PostalCode == "" {
			return fmt.Errorf("saft: missing Customer.BillingAddress.PostalCode")
		}

		if customer.BillingAddress.Country == "" {
			return fmt.Errorf("saft: missing Customer.BillingAddress.Country")
		}

		if len(customer.CustomerTaxId) == 9 && customer.BillingAddress.Country == "PT" {
			ok := common.ValidateNIFPT(string(customer.CustomerTaxId))
			if !ok {
				return fmt.Errorf("saft: invalid NIF in Customer.CustomerTaxId: %s", customer.CustomerTaxId)
			}
		}

		for _, shipTo := range customer.ShipToAddress {
			if shipTo.AddressDetail == "" {
				return fmt.Errorf("saft: missing ShipTo.AddressDetail")
			}

			if shipTo.City == "" {
				return fmt.Errorf("saft: missing ShipTo.City")
			}

			if shipTo.PostalCode == "" {
				return fmt.Errorf("saft: missing ShipTo.PostalCode")
			}

			if shipTo.Country == "" {
				return fmt.Errorf("saft: missing ShipTo.Country")
			}
		}

		if customer.SelfBillingIndicator != SelfBillingIndicatorNo && customer.SelfBillingIndicator != SelfBillingIndicatorYes {
			return fmt.Errorf("saft: invalid Customer.SelfBillingIndicator: %d", customer.SelfBillingIndicator)
		}
	}

	return nil
}

func (a *AuditFile) checkPayments() error {
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
				amount := decimal.Decimal(*line.DebitAmount)
				totalDebit = totalDebit.Add(amount)
			}

			if line.CreditAmount != nil {
				amount := decimal.Decimal(*line.CreditAmount)
				totalCredit = totalCredit.Add(amount)
			}
		}
	}

	// Check if the total debit and total credit are the same
	if a.SourceDocuments.Payments.NumberOfEntries != numPayments {
		return fmt.Errorf("saft: invalid NumberOfEntries: %d != calculated %d", a.SourceDocuments.Payments.NumberOfEntries, numPayments)
	}

	if decimal.Decimal(a.SourceDocuments.Payments.TotalCredit).Cmp(totalCredit) != 0 {
		return fmt.Errorf("saft: invalid TotalCredit: %s != calculated %s", a.SourceDocuments.Payments.TotalCredit, totalCredit)
	}

	if decimal.Decimal(a.SourceDocuments.Payments.TotalDebit).Cmp(totalDebit) != 0 {
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

		if payment.TransactionDate == (SafdateType{}) {
			return fmt.Errorf("saft: missing Payment.TransactionDate")
		}

		if time.Time(payment.TransactionDate).After(time.Now()) {
			return fmt.Errorf("saft: invalid Payment.TransactionDate")
		}

		if payment.PaymentType != SaftptpaymentTypeRC && payment.PaymentType != SaftptpaymentTypeRG {
			return fmt.Errorf("saft: invalid Payment.PaymentType")
		}

		if payment.DocumentStatus == (PaymentDocumentStatus{}) {
			return fmt.Errorf("saft: missing Payment.DocumentStatus")
		}

		if payment.DocumentStatus.PaymentStatus != PaymentStatusNormal && payment.DocumentStatus.PaymentStatus != PaymentStatusCancelled {
			return fmt.Errorf("saft: invalid Payment.DocumentStatus.PaymentStatus")
		}

		if payment.DocumentStatus.PaymentStatusDate == (SafdateTimeType{}) {
			return fmt.Errorf("saft: missing Payment.DocumentStatus.PaymentStatusDate")
		}

		if time.Time(payment.DocumentStatus.PaymentStatusDate).After(time.Now()) {
			return fmt.Errorf("saft: invalid Payment.DocumentStatus.PaymentStatusDate")
		}

		if payment.DocumentStatus.SourceId == "" {
			return fmt.Errorf("saft: missing Payment.DocumentStatus.SourceId")
		}

		if payment.DocumentStatus.SourcePayment != SaftptsourcePaymentP &&
			payment.DocumentStatus.SourcePayment != SaftptsourcePaymentI &&
			payment.DocumentStatus.SourcePayment != SaftptsourcePaymentM {
			return fmt.Errorf("saft: invalid Payment.DocumentStatus.SourcePayment")
		}

		if len(payment.PaymentMethod) == 0 {
			return fmt.Errorf("saft: missing Payment.PaymentMethod")
		}

		taxPayable := decimal.NewFromInt(0)
		for _, method := range payment.PaymentMethod {
			if method.PaymentMechanism != nil {
				switch *method.PaymentMechanism {
				case PaymentMechanismCC, PaymentMechanismCD, PaymentMechanismCH, PaymentMechanismCI, PaymentMechanismCO, PaymentMechanismCS, PaymentMechanismDE, PaymentMechanismLC, PaymentMechanismMB, PaymentMechanismNU, PaymentMechanismOU, PaymentMechanismPR, PaymentMechanismTB, PaymentMechanismTR:
					// Ignore
				default:
					return fmt.Errorf("saft: invalid PaymentMethod.PaymentMechanism: %s", *method.PaymentMechanism)
				}
			}

			if decimal.Decimal(method.PaymentAmount).LessThan(decimal.NewFromInt(0)) {
				return fmt.Errorf("saft: invalid PaymentMethod.PaymentAmount: %s", method.PaymentAmount)
			}

			if method.PaymentDate == (SafdateType{}) {
				return fmt.Errorf("saft: missing PaymentMethod.PaymentDate")
			}

			if time.Time(method.PaymentDate).After(time.Now()) {
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
					if decimal.Decimal(*line.DebitAmount).LessThan(decimal.NewFromInt(0)) {
						return fmt.Errorf("saft: invalid PaymentLine.DebitAmount: %s", line.DebitAmount)
					}
				}

				if line.CreditAmount != nil {
					if decimal.Decimal(*line.CreditAmount).LessThan(decimal.NewFromInt(0)) {
						return fmt.Errorf("saft: invalid PaymentLine.CreditAmount: %s", line.CreditAmount)
					}
				}

				// If PaymentType is RC, then Tax must be present
				if payment.PaymentType == SaftptpaymentTypeRC && line.Tax == nil {
					return fmt.Errorf("saft: missing PaymentLine.Tax")
				}

				// If PaymentType is RG, then Tax must not be present
				if payment.PaymentType == SaftptpaymentTypeRG && line.Tax != nil {
					return fmt.Errorf("saft: invalid PaymentLine.Tax")
				}

				if line.Tax != nil {
					if line.Tax.TaxType != TaxTypeIVA && line.Tax.TaxType != TaxTypeIS && line.Tax.TaxType != TaxTypeNS {
						return fmt.Errorf("saft: invalid PaymentLine.Tax.TaxType: %s", line.Tax.TaxType)
					}

					// Check if its a ISO 3166-1 alpha-2 country code + PT regions
					if slices.Contains(common.CountryCodesPTRegions, line.Tax.TaxCountryRegion) {
						return fmt.Errorf("saft: invalid PaymentLine.Tax.TaxCountryRegion: %s", line.Tax.TaxCountryRegion)
					}

					// Cant have both TaxPercentage and TaxAmount
					if line.Tax.TaxPercentage != nil && line.Tax.TaxAmount != nil {
						return fmt.Errorf("saft: invalid PaymentLine.Tax: both TaxPercentage and TaxAmount present")
					}

					// TaxPercentage can only be zero if TaxCode is Ise or Na
					if line.Tax.TaxPercentage != nil &&
						decimal.Decimal(*line.Tax.TaxPercentage) == decimal.NewFromInt(0) &&
						(line.Tax.TaxCode != PaymentTaxCode(TaxCodeIse) && line.Tax.TaxCode != PaymentTaxCode(TaxCodeNa)) {
						return fmt.Errorf("saft: invalid PaymentLine.Tax: TaxPercentage is zero but TaxCode is not Ise or Na")
					}

					// TaxPercentage must be 0 or greater
					if line.Tax.TaxPercentage != nil && decimal.Decimal(*line.Tax.TaxPercentage).LessThan(decimal.NewFromInt(0)) {
						return fmt.Errorf("saft: invalid PaymentLine.Tax.TaxPercentage: %s", line.Tax.TaxPercentage)
					}

					// Calculate the TaxPayable using the TaxPercentage
					if line.Tax.TaxPercentage != nil {
						taxPercentage := decimal.Decimal(*line.Tax.TaxPercentage)
						debitAmount := decimal.Decimal(*line.DebitAmount)
						taxMultiplier := taxPercentage.Div(decimal.NewFromInt(100)).Add(decimal.NewFromInt(1))
						taxPayable = taxPayable.Add((debitAmount.Mul(taxMultiplier)).Sub(debitAmount))
					}

					// TaxAmount can only appear if TaxType is IS
					if line.Tax.TaxAmount != nil && line.Tax.TaxType != TaxTypeIS {
						return fmt.Errorf("saft: invalid PaymentLine.Tax: TaxAmount present but TaxType is not IS")
					}

					if decimal.Decimal(*line.Tax.TaxAmount).LessThan(decimal.NewFromInt(0)) {
						return fmt.Errorf("saft: invalid PaymentLine.Tax.TaxAmount: %s", line.Tax.TaxAmount)
					}

					// Add the TaxAmount to the TaxPayable
					if line.Tax.TaxAmount != nil {
						taxPayable = taxPayable.Add(decimal.Decimal(*line.Tax.TaxAmount))
					}

				}

				//// Cant have Tax and TaxExemptionReason and TaxExemptionCode
				//if line.Tax != nil && line.TaxExemptionReason != nil && line.TaxExemptionCode != nil {
				//	return errcodes.ErrPaymentLineTaxTaxExemption
				//}

				// TaxExemptionReason and TaxExemptionCode must exist when TaxPercentage/TaxAmount is 0
				if line.Tax != nil && line.Tax.TaxPercentage != nil && decimal.Decimal(*line.Tax.TaxPercentage) == decimal.NewFromInt(0) {
					if line.TaxExemptionReason == nil || line.TaxExemptionCode == nil {
						return fmt.Errorf("saft: missing PaymentLine.TaxExemptionReason or PaymentLine.TaxExemptionCode")
					}
				}

				// Check if TaxExemption is valid
				if line.TaxExemptionCode != nil && line.TaxExemptionReason != nil {
					for _, exemption := range common.VatExemptionCodes {
						if SafptportugueseTaxExemptionCode(exemption.Code) == *line.TaxExemptionCode && SafptportugueseTaxExemptionReason(exemption.Description) == *line.TaxExemptionReason {
							break
						}
					}
					return fmt.Errorf("saft: invalid PaymentLine.TaxExemptionCode or PaymentLine.TaxExemptionReason")
				}
			}

		}

		if payment.DocumentTotals == (PaymentDocumentTotals{}) {
			return fmt.Errorf("saft: missing Payment.DocumentTotals")
		}

		if decimal.Decimal(payment.DocumentTotals.TaxPayable).Cmp(taxPayable) != 0 {
			return fmt.Errorf("saft: invalid Payment.DocumentTotals.TaxPayable: %s != calculated %s", payment.DocumentTotals.GrossTotal, totalDebit)
		}

		// Calculate NetTotal and GrossTotal
		netTotal := decimal.NewFromInt(0)
		grossTotal := decimal.NewFromInt(0)

		for _, line := range payment.Line {
			if line.DebitAmount != nil {
				netTotal = netTotal.Add(decimal.Decimal(*line.DebitAmount))
				//grossTotal = grossTotal.Add(decimal.Decimal(*line.DebitAmount))
			}

			if line.CreditAmount != nil {
				netTotal = netTotal.Sub(decimal.Decimal(*line.CreditAmount))
				//grossTotal = grossTotal.Sub(decimal.Decimal(*line.CreditAmount))
			}

			//  get the tax amount
			if line.Tax != nil && line.Tax.TaxPercentage != nil {
				taxMultiplier := decimal.Decimal(*line.Tax.TaxPercentage).Div(decimal.NewFromInt(100)).Add(decimal.NewFromInt(1))
				grossTotal = grossTotal.Add(netTotal.Div(taxMultiplier))
			} else if line.Tax != nil && line.Tax.TaxAmount != nil {
				grossTotal = grossTotal.Add(netTotal.Add(decimal.Decimal(*line.Tax.TaxAmount)))
			}
		}

	}
	return nil
}

func (a *AuditFile) checkConstraints() error {
	// Master Files Constraints
	// CustomerIDConstraint
	customers := make(map[SafpttextTypeMandatoryMax30Car]bool)
	for _, customer := range a.MasterFiles.Customer {
		if _, ok := customers[customer.CustomerId]; ok {
			//return errcodes.ErrUQCustomerId
			return fmt.Errorf("saft: unique constraint violated on Customer.CustomerId: %s", customer.CustomerId)
		}
		customers[customer.CustomerId] = true
	}

	// SupplierIDConstraint
	suppliers := make(map[SafpttextTypeMandatoryMax30Car]bool)
	for _, supplier := range a.MasterFiles.Supplier {
		if _, ok := suppliers[supplier.SupplierId]; ok {
			//return errcodes.ErrUQSupplierId
			return fmt.Errorf("saft: unique constraint violated on Supplier.SupplierId: %s", supplier.SupplierId)
		}
		suppliers[supplier.SupplierId] = true
	}

	// ProductCodeConstraint
	products := make(map[SafpttextTypeMandatoryMax60Car]bool)
	for _, product := range a.MasterFiles.Product {
		if _, ok := products[product.ProductCode]; ok {
			//return errcodes.ErrUQProductCode
			return fmt.Errorf("saft: unique constraint violated on Product.ProductCode: %s", product.ProductCode)
		}
		products[product.ProductCode] = true
	}

	// General Ledger Constraints
	transactions := make(map[SafpttransactionId]bool)
	if a.MasterFiles.GeneralLedgerAccounts != nil {
		accounts := make(map[SafptglaccountId]bool)
		for _, account := range a.MasterFiles.GeneralLedgerAccounts.Account {
			// AccountIDConstraint
			if _, ok := accounts[account.AccountId]; ok {
				//return errcodes.ErrUQAccountId
				return fmt.Errorf("saft: unique constraint violated on GeneralLedgerAccounts.Account.AccountId: %s", account.AccountId)
			}
			accounts[account.AccountId] = true

			// GroupingCodeConstraint
			if account.GroupingCode != nil && *account.GroupingCode != "" {
				if _, ok := accounts[*account.GroupingCode]; !ok {
					//return errcodes.ErrKRGenLedgerEntriesAccountID
					// key reference GeneralLedgerAccounts.Account.GroupingCode
					return fmt.Errorf("saft: key reference violated on GeneralLedgerAccounts.Account.GroupingCode: %s", *account.GroupingCode)
				}
			}
		}

		journals := make(map[SafptjournalId]bool)
		for _, entry := range a.GeneralLedgerEntries.Journal {
			// GeneralLedgerEntriesJournalIdConstraint
			if _, ok := journals[entry.JournalId]; ok {
				//return errcodes.ErrUQJournalId
				return fmt.Errorf("saft: unique constraint violated on GeneralLedgerEntries.Journal.JournalId: %s", entry.JournalId)
			}
			journals[entry.JournalId] = true

			for _, line := range entry.Transaction {
				// GeneralLedgerEntriesDebitLineAccountIDConstraint
				for _, debit := range line.Lines.DebitLine {
					if _, ok := accounts[debit.AccountId]; !ok {
						//return errcodes.ErrKRGenLedgerEntriesAccountID
						return fmt.Errorf("saft: key reference violated on GeneralLedgerEntries.Transaction.Lines.DebitLine.AccountId: %s", debit.AccountId)
					}
				}

				// GeneralLedgerEntriesCreditLineAccountIDConstraint
				for _, credit := range line.Lines.CreditLine {
					if _, ok := accounts[credit.AccountId]; !ok {
						//return errcodes.ErrKRGenLedgerEntriesAccountID
						return fmt.Errorf("saft: key reference violated on GeneralLedgerEntries.Transaction.Lines.CreditLine.AccountId: %s", credit.AccountId)
					}
				}

				// GeneralLedgerEntriesCustomerIDConstraint
				if line.CustomerId != nil && *line.CustomerId != "" {
					if _, ok := customers[*line.CustomerId]; !ok {
						//return errcodes.ErrKRGenLedgerEntriesCustomerID
						return fmt.Errorf("saft: key reference violated on GeneralLedgerEntries.Transaction.CustomerId: %s", *line.CustomerId)
					}
				}

				// GeneralLedgerEntriesSupplierIDConstraint
				if line.SupplierId != nil && *line.SupplierId != "" {
					if _, ok := suppliers[*line.SupplierId]; !ok {
						//return errcodes.ErrKRGenLedgerEntriesSupplierID
						return fmt.Errorf("saft: key reference violated on GeneralLedgerEntries.Transaction.SupplierId: %s", *line.SupplierId)
					}
				}

				// GeneralLedgerEntriesTransactionIdConstraint
				if _, ok := transactions[line.TransactionId]; ok {
					//return errcodes.ErrUQTransactionId
					return fmt.Errorf("saft: unique constraint violated on GeneralLedgerEntries.Transaction.TransactionId: %s", line.TransactionId)
				}
				transactions[line.TransactionId] = true
			}
		}
	}

	// Sales Invoices constraints
	if a.SourceDocuments != nil && a.SourceDocuments.SalesInvoices != nil {
		invoices := make(map[string]bool)
		for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
			// InvoiceNoConstraint
			if _, ok := invoices[invoice.InvoiceNo]; ok {
				//return errcodes.ErrUQInvoiceNo
				return fmt.Errorf("saft: unique constraint violated on SalesInvoices.Invoice.InvoiceNo: %s", invoice.InvoiceNo)
			}
			invoices[invoice.InvoiceNo] = true

			// InvoiceCustomerIDConstraint
			if _, ok := customers[invoice.CustomerId]; !ok {
				//return errcodes.ErrKRInvoiceCustomerID
				return fmt.Errorf("saft: key reference violated on SalesInvoices.Invoice.CustomerId: %s", invoice.CustomerId)
			}

			// InvoiceProductCodeConstraint
			for _, line := range invoice.Line {
				if _, ok := products[line.ProductCode]; !ok {
					//return errcodes.ErrKRInvoiceProductCode
					return fmt.Errorf("saft: key reference violated on SalesInvoices.Invoice.Line.ProductCode: %s", line.ProductCode)
				}
			}
		}
	}

	if a.SourceDocuments != nil && a.SourceDocuments.MovementOfGoods != nil {
		documents := make(map[string]bool)
		for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
			// DocumentNumberConstraint
			if _, ok := documents[stock.DocumentNumber]; !ok {
				//return errcodes.ErrUQDocumentNo
				return fmt.Errorf("saft: unique constraint violated on MovementOfGoods.StockMovement.DocumentNumber: %s", stock.DocumentNumber)
			}
			documents[stock.DocumentNumber] = true

			// StockMovementCustomerIDConstraint
			if _, ok := customers[*stock.CustomerId]; !ok {
				//return errcodes.ErrKRStockMovementCustomerID
				return fmt.Errorf("saft: key reference violated on MovementOfGoods.StockMovement.CustomerId: %s", *stock.CustomerId)
			}

			// StockMovementSupplierIDConstraint
			if _, ok := suppliers[*stock.SupplierId]; !ok {
				//return errcodes.ErrKRStockMovementSupplierID
				return fmt.Errorf("saft: key reference violated on MovementOfGoods.StockMovement.SupplierId: %s", *stock.SupplierId)
			}

			// StockMovementProductCodeConstraint
			for _, line := range stock.Line {
				if _, ok := products[line.ProductCode]; !ok {
					//return errcodes.ErrKRStockMovementProductCode
					return fmt.Errorf("saft: key reference violated on MovementOfGoods.StockMovement.Line.ProductCode: %s", line.ProductCode)
				}
			}
		}
	}

	if a.SourceDocuments != nil && a.SourceDocuments.WorkingDocuments != nil {
		workDocs := make(map[string]bool)
		for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
			// WorkDocumentDocumentNumberConstraint
			if _, ok := workDocs[workDoc.DocumentNumber]; !ok {
				//return errcodes.ErrUQWorkDocNo
				return fmt.Errorf("saft: unique constraint violated on WorkingDocuments.WorkDocument.DocumentNumber: %s", workDoc.DocumentNumber)
			}
			workDocs[workDoc.DocumentNumber] = true

			// WorkDocumentDocumentCustomerIDConstraint
			if _, ok := customers[workDoc.CustomerId]; !ok {
				//return errcodes.ErrKRWorkDocumentCustomerID
				return fmt.Errorf("saft: key reference violated on WorkingDocuments.WorkDocument.CustomerId: %s", workDoc.CustomerId)
			}

			// WorkDocumentDocumentProductCodeConstraint
			for _, line := range workDoc.Line {
				if _, ok := products[line.ProductCode]; !ok {
					//return errcodes.ErrKRWorkDocumentProductCode
					return fmt.Errorf("saft: key reference violated on WorkingDocuments.WorkDocument.Line.ProductCode: %s", line.ProductCode)
				}
			}
		}
	}

	if a.SourceDocuments != nil && a.SourceDocuments.Payments != nil {
		payments := make(map[string]bool)
		for _, payment := range a.SourceDocuments.Payments.Payment {
			// PaymentPaymentRefNoConstraint

			if _, ok := payments[payment.PaymentRefNo]; ok {
				//return errcodes.ErrUQPaymentRefNo
				return fmt.Errorf("saft: unique constraint violated on Payments.Payment.PaymentRefNo: %s", payment.PaymentRefNo)
			}
			payments[payment.PaymentRefNo] = true

			// PaymentPaymentRefNoCustomerIDConstraint
			if _, ok := customers[payment.CustomerId]; !ok {
				//return errcodes.ErrKRStockMovementCustomerID
				return fmt.Errorf("saft: key reference violated on Payments.Payment.CustomerId: %s", payment.CustomerId)
			}

			// Check TransactionId against GeneralLedgerEntries
			// Not an official constraint
			if payment.TransactionId != nil && *payment.TransactionId != "" {
				if len(transactions) != 0 {
					if _, ok := transactions[*payment.TransactionId]; !ok {
						//return errcodes.ErrKRPaymentTransactionId
						return fmt.Errorf("saft: key reference violated on Payments.Payment.TransactionId: %s", *payment.TransactionId)
					}
				}

			}
		}
	}

	return nil
}
