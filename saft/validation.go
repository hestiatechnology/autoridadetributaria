package saft

import (
	"errors"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/common"
	"github.com/hestiatechnology/autoridadetributaria/saft/errcodes"
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
				return errcodes.ErrGroupingCategoryTaxonomyCode
			}
			// Test 2
			if (account.GroupingCategory == "GR" && account.GroupingCode != nil) ||
				(account.GroupingCategory == "AR" && account.GroupingCode != nil) ||
				(account.GroupingCategory == "GA" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "AA" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "GM" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "AM" && account.GroupingCode == nil) {
				// logger.DebugLogger.Println("Err: Invalid GroupingCategory and GroupingCode combination: ", account.GroupingCategory, *account.GroupingCode)
				return errcodes.ErrGroupingCategoryGroupingCode
			}
		}
	}

	for _, customer := range a.MasterFiles.Customer {
		r := regexp.MustCompile(`([^^]*)`)
		// Check if it doesnt match Desconhecido or the regex
		if customer.AccountId != "Desconhecido" && !r.MatchString(customer.AccountId) {
			return errcodes.ErrValidationAccountId
		}
	}

	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		if invoice.SpecialRegimes.CashVatschemeIndicator != CashVatschemeIndicatorNo && invoice.SpecialRegimes.CashVatschemeIndicator != CashVatschemeIndicatorYes {
			return errcodes.ErrCashVatschemeIndicator
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
		return errcodes.ErrUnsupportedSAFTVersion
	}

	// CompanyId must be either a NIF (9 numbers) or Conservatória (string) + NIF
	r := regexp.MustCompile(`([0-9]{9})+|([^^]+ [0-9/]+)`)
	if !r.MatchString(a.Header.CompanyId) {
		return errcodes.ErrCompanyId
	}

	// Check if TaxRegistrationNumber is a NIF (9 numbers)
	taxNumber := strconv.FormatUint(uint64(a.Header.TaxRegistrationNumber), 10)
	ok := common.ValidateNIFPT(taxNumber)
	if !ok {
		// TODO: ADD ERROR
		return common.ErrInvalidNIFPT
	}

	switch a.Header.TaxAccountingBasis {
	case SaftAccounting, SaftInvoicingThirdParties, SaftInvoicing, SaftIntegrated, SaftInvoicingParcial, SaftPayments, SaftSelfBilling, SaftTransportDocuments:
		break
	default:
		return errcodes.ErrInvalidSaftType
	}

	if a.Header.CompanyName == "" {
		return errcodes.ErrEmptyCompanyName
	}

	// Check if CompanyAddress is empty
	if a.Header.CompanyAddress == (AddressStructure{}) {
		return errcodes.ErrEmptyCompanyAddress
	}

	if a.Header.CompanyAddress.AddressDetail == "" {
		return errcodes.ErrEmptyAddressDetail
	}

	if a.Header.CompanyAddress.City == "" {
		return errcodes.ErrEmptyCity
	}

	if a.Header.CompanyAddress.PostalCode == "" {
		return errcodes.ErrEmptyPostalCode
	}

	if a.Header.CompanyAddress.Country != "PT" {
		return errcodes.ErrInvalidCountry
	}

	if a.Header.FiscalYear == "" {
		return errcodes.ErrInvalidFiscalYear
	}

	// Check if its a valid year
	year, err := strconv.Atoi(a.Header.FiscalYear)
	if err != nil {
		return errcodes.ErrInvalidFiscalYear
	}

	// check if its higher than current year
	currentYear := time.Now().Year()
	if year < 2000 && year > currentYear {
		return errcodes.ErrInvalidFiscalYear
	}

	if a.Header.StartDate == (SafptdateSpan{}) {
		return errcodes.ErrInvalidStartDate
	}

	if a.Header.EndDate == (SafptdateSpan{}) {
		return errcodes.ErrInvalidEndDate
	}

	if time.Time(a.Header.StartDate).After(time.Time(a.Header.EndDate)) {
		return errcodes.ErrInvalidDateSpan
	}

	if a.Header.CurrencyCode != "EUR" {
		return errcodes.ErrCurrencyNotEuro
	}

	if a.Header.DateCreated == (SafptdateSpan{}) {
		return errcodes.ErrMissingDateCreated
	}

	if a.Header.TaxEntity == "" {
		return errcodes.ErrMissingTaxEntity
	}

	if a.Header.ProductCompanyTaxId == "" {
		return errcodes.ErrMissingProductCompanyTaxId
	}

	if len(a.Header.ProductCompanyTaxId) != 9 {
		// TODO: ADD ERROR
		return common.ErrInvalidNIFPT
	}

	ok = common.ValidateNIFPT(string(a.Header.ProductCompanyTaxId))
	if !ok {
		// TODO: ADD ERROR
		return common.ErrInvalidNIFPT
	}

	//SoftwareCertificateNumber is by default 0, which is a valid value

	r = regexp.MustCompile(`[^/]+\/[^/]+`)
	if !r.MatchString(string(a.Header.ProductId)) {
		return errcodes.ErrInvalidProductId
	}

	if a.Header.ProductVersion == "" {
		return errcodes.ErrMissingProductVersion
	}

	return nil
}

func (a *AuditFile) checkCustomers() error {
	for _, customer := range a.MasterFiles.Customer {
		if customer.CustomerId == "" {
			return errcodes.ErrMissingCustomerID
		}

		if customer.AccountId == "" {
			return errcodes.ErrMissingAccountID
		}

		if customer.CustomerTaxId == "" {
			return errcodes.ErrMissingCustomerTaxID
		}

		if customer.CompanyName == "" {
			return errcodes.ErrMissingCompanyName
		}

		if customer.BillingAddress == (CustomerAddressStructure{}) {
			return errcodes.ErrMissingBillingAddress
		}

		if customer.BillingAddress.AddressDetail == "" {
			return errcodes.ErrMissingAddressDetail
		}

		if customer.BillingAddress.City == "" {
			return errcodes.ErrMissingCity
		}

		if customer.BillingAddress.PostalCode == "" {
			return errcodes.ErrMissingPostalCode
		}

		if customer.BillingAddress.Country == "" {
			return errcodes.ErrMissingCountry
		}

		if len(customer.CustomerTaxId) == 9 && customer.BillingAddress.Country == "PT" {
			ok := common.ValidateNIFPT(string(customer.CustomerTaxId))
			if !ok {
				return errors.Join(errcodes.ErrInvalidCustomerTaxID, common.ErrInvalidNIFPT)
			}
		}

		for _, shipTo := range customer.ShipToAddress {
			if shipTo.AddressDetail == "" {
				return errcodes.ErrMissingAddressDetail
			}

			if shipTo.City == "" {
				return errcodes.ErrMissingCity
			}

			if shipTo.PostalCode == "" {
				return errcodes.ErrMissingPostalCode
			}

			if shipTo.Country == "" {
				return errcodes.ErrMissingCountry
			}
		}

		if customer.SelfBillingIndicator != SelfBillingIndicatorNo && customer.SelfBillingIndicator != SelfBillingIndicatorYes {
			return errcodes.ErrInvalidSelfBillingIndicator
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
			return errcodes.ErrMissingPaymentLine
		}

		numPayments++
		for _, line := range payment.Line {
			// Can't have both debit and credit
			if line.DebitAmount != nil && line.CreditAmount != nil {
				return errcodes.ErrDebitCreditAmount
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
		return errcodes.ErrPaymentNumberOfEntries
	}

	if a.SourceDocuments.Payments.TotalCredit != SafmonetaryType(totalCredit) {
		return errcodes.ErrPaymentTotalCredit
	}

	if a.SourceDocuments.Payments.TotalDebit != SafmonetaryType(totalDebit) {
		return errcodes.ErrPaymentTotalDebit
	}

	for _, payment := range a.SourceDocuments.Payments.Payment {
		if payment.PaymentRefNo == "" {
			return errcodes.ErrMissingPaymentRefNo
		}

		if payment.Atcud == "" {
			return errcodes.ErrMissingPaymentATCUD
		}

		// TODO: Validate ATCUD?

		// Transaction Id validation is done in the constraints

		if payment.TransactionDate == (SafdateType{}) {
			return errcodes.ErrMissingPaymentTransactionDate
		}

		if time.Time(payment.TransactionDate).After(time.Now()) {
			return errcodes.ErrInvalidPaymentTransactionDate
		}

		if payment.PaymentType != SaftptpaymentTypeRC && payment.PaymentType != SaftptpaymentTypeRG {
			return errcodes.ErrInvalidPaymentType
		}

		if payment.DocumentStatus == (PaymentDocumentStatus{}) {
			return errcodes.ErrMissingPaymentDocumentStatus
		}

		if payment.DocumentStatus.PaymentStatus != PaymentStatusNormal && payment.DocumentStatus.PaymentStatus != PaymentStatusCancelled {
			return errcodes.ErrInvalidPaymentStatus
		}

		if payment.DocumentStatus.PaymentStatusDate == (SafdateTimeType{}) {
			return errcodes.ErrMissingPaymentDocumentStatusDate
		}

		if time.Time(payment.DocumentStatus.PaymentStatusDate).After(time.Now()) {
			return errcodes.ErrInvalidPaymentDocumentStatusDate
		}

		if payment.DocumentStatus.SourceId == "" {
			return errcodes.ErrMissingPaymentDocumentStatusSourceId
		}

		if payment.DocumentStatus.SourcePayment != SaftptsourcePaymentP &&
			payment.DocumentStatus.SourcePayment != SaftptsourcePaymentI &&
			payment.DocumentStatus.SourcePayment != SaftptsourcePaymentM {
			return errcodes.ErrInvalidPaymentDocumentStatusSourcePayment
		}

		if len(payment.PaymentMethod) == 0 {
			return errcodes.ErrMissingPaymentMethod
		}

		for _, method := range payment.PaymentMethod {
			if method.PaymentMechanism != nil {
				switch *method.PaymentMechanism {
				case PaymentMechanismCC, PaymentMechanismCD, PaymentMechanismCH, PaymentMechanismCI, PaymentMechanismCO, PaymentMechanismCS, PaymentMechanismDE, PaymentMechanismLC, PaymentMechanismMB, PaymentMechanismNU, PaymentMechanismOU, PaymentMechanismPR, PaymentMechanismTB, PaymentMechanismTR:
					// Ignore
				default:
					return errcodes.ErrInvalidPaymentMechanism
				}
			}

			if decimal.Decimal(method.PaymentAmount).LessThan(decimal.NewFromInt(0)) {
				return errcodes.ErrPaymentAmount
			}

			if method.PaymentDate == (SafdateType{}) {
				return errcodes.ErrInvalidPaymentDate
			}

			if time.Time(method.PaymentDate).After(time.Now()) {
				return errcodes.ErrInvalidPaymentDate
			}

			if payment.SourceId == "" {
				return errcodes.ErrMissingPaymentSourceId
			}

			if payment.CustomerId == "" {
				return errcodes.ErrMissingPaymentCustomerID
			}

			for _, line := range payment.Line {
				// LineNumber can be 0 I guess

				if len(line.SourceDocumentId) == 0 {
					return errcodes.ErrPaymentLineSourceDocumentId
				}
				for _, sourceDocument := range line.SourceDocumentId {
					if sourceDocument.OriginatingOn == "" {
						return errcodes.ErrPaymentLineSourceDocumentIdOriginatingOn
					}
				}

				if line.DebitAmount != nil {
					if decimal.Decimal(*line.DebitAmount).LessThan(decimal.NewFromInt(0)) {
						return errcodes.ErrPaymentAmount
					}
				}

				if line.CreditAmount != nil {
					if decimal.Decimal(*line.CreditAmount).LessThan(decimal.NewFromInt(0)) {
						return errcodes.ErrPaymentAmount
					}
				}

				// If PaymentType is RC, then Tax must be present
				if payment.PaymentType == SaftptpaymentTypeRC && line.Tax == nil {
					return errcodes.ErrPaymentLineTax
				}

				// If PaymentType is RG, then Tax must not be present
				if payment.PaymentType == SaftptpaymentTypeRG && line.Tax != nil {
					return errcodes.ErrPaymentLineTax
				}

				if line.Tax != nil {
					if line.Tax.TaxType != TaxTypeIVA && line.Tax.TaxType != TaxTypeIS && line.Tax.TaxType != TaxTypeNS {
						return errcodes.ErrPaymentLineTaxType
					}

					// Check if its a ISO 3166-1 alpha-2 country code + PT regions
					if slices.Contains(common.CountryCodesPTRegions, line.Tax.TaxCountryRegion) {
						return errcodes.ErrPaymentLineTaxCountryRegion
					}

					// Cant have both TaxPercentage and TaxAmount
					if line.Tax.TaxPercentage != nil && line.Tax.TaxAmount != nil {
						return errcodes.ErrPaymentLineTaxPercentageAmount
					}

					// TaxPercentage can only be zero if TaxCode is Ise or Na
					if line.Tax.TaxPercentage != nil &&
						decimal.Decimal(*line.Tax.TaxPercentage) == decimal.NewFromInt(0) &&
						(line.Tax.TaxCode != PaymentTaxCode(TaxCodeIse) && line.Tax.TaxCode != PaymentTaxCode(TaxCodeNa)) {
						return errcodes.ErrPaymentLineTaxPercentage
					}

					// TaxPercentage must be 0 or greater
					if line.Tax.TaxPercentage != nil && decimal.Decimal(*line.Tax.TaxPercentage).LessThan(decimal.NewFromInt(0)) {
						return errcodes.ErrPaymentLineTaxPercentage
					}

					// TaxAmount can only appear if TaxType is IS
					if line.Tax.TaxAmount != nil && line.Tax.TaxType != TaxTypeIS {
						return errcodes.ErrPaymentLineTaxAmount
					}

					if decimal.Decimal(*line.Tax.TaxAmount).LessThan(decimal.NewFromInt(0)) {
						return errcodes.ErrPaymentLineTaxAmount
					}
				}

				//// Cant have Tax and TaxExemptionReason and TaxExemptionCode
				//if line.Tax != nil && line.TaxExemptionReason != nil && line.TaxExemptionCode != nil {
				//	return errcodes.ErrPaymentLineTaxTaxExemption
				//}

				// TaxExemptionReason and TaxExemptionCode must exist when TaxPercentage/TaxAmount is 0
				if line.Tax != nil && line.Tax.TaxPercentage != nil && decimal.Decimal(*line.Tax.TaxPercentage) == decimal.NewFromInt(0) {
					if line.TaxExemptionReason == nil || line.TaxExemptionCode == nil {
						return errcodes.ErrPaymentLineTaxTaxExemption
					}
				}

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
			return errcodes.ErrUQCustomerId
		}
		customers[customer.CustomerId] = true
	}

	// SupplierIDConstraint
	suppliers := make(map[SafpttextTypeMandatoryMax30Car]bool)
	for _, supplier := range a.MasterFiles.Supplier {
		if _, ok := suppliers[supplier.SupplierId]; ok {
			return errcodes.ErrUQSupplierId
		}
		suppliers[supplier.SupplierId] = true
	}

	// ProductCodeConstraint
	products := make(map[SafpttextTypeMandatoryMax60Car]bool)
	for _, product := range a.MasterFiles.Product {
		if _, ok := products[product.ProductCode]; ok {
			return errcodes.ErrUQProductCode
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
				return errcodes.ErrUQAccountId
			}
			accounts[account.AccountId] = true

			// GroupingCodeConstraint
			if account.GroupingCode != nil && *account.GroupingCode != "" {
				if _, ok := accounts[*account.GroupingCode]; !ok {
					return errcodes.ErrKRGenLedgerEntriesAccountID
				}
			}
		}

		journals := make(map[SafptjournalId]bool)
		for _, entry := range a.GeneralLedgerEntries.Journal {
			// GeneralLedgerEntriesJournalIdConstraint
			if _, ok := journals[entry.JournalId]; ok {
				return errcodes.ErrUQJournalId
			}
			journals[entry.JournalId] = true

			for _, line := range entry.Transaction {
				// GeneralLedgerEntriesDebitLineAccountIDConstraint
				for _, debit := range line.Lines.DebitLine {
					if _, ok := accounts[debit.AccountId]; !ok {
						return errcodes.ErrKRGenLedgerEntriesAccountID
					}
				}

				// GeneralLedgerEntriesCreditLineAccountIDConstraint
				for _, credit := range line.Lines.CreditLine {
					if _, ok := accounts[credit.AccountId]; !ok {
						return errcodes.ErrKRGenLedgerEntriesAccountID
					}
				}

				// GeneralLedgerEntriesCustomerIDConstraint
				if line.CustomerId != nil && *line.CustomerId != "" {
					if _, ok := customers[*line.CustomerId]; !ok {
						return errcodes.ErrKRGenLedgerEntriesCustomerID
					}
				}

				// GeneralLedgerEntriesSupplierIDConstraint
				if line.SupplierId != nil && *line.SupplierId != "" {
					if _, ok := suppliers[*line.SupplierId]; !ok {
						return errcodes.ErrKRGenLedgerEntriesSupplierID
					}
				}

				// GeneralLedgerEntriesTransactionIdConstraint
				if _, ok := transactions[line.TransactionId]; ok {
					return errcodes.ErrUQTransactionId
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
				return errcodes.ErrUQInvoiceNo
			}
			invoices[invoice.InvoiceNo] = true

			// InvoiceCustomerIDConstraint
			if _, ok := customers[invoice.CustomerId]; !ok {
				return errcodes.ErrKRInvoiceCustomerID
			}

			// InvoiceProductCodeConstraint
			for _, line := range invoice.Line {
				if _, ok := products[line.ProductCode]; !ok {
					return errcodes.ErrKRInvoiceProductCode
				}
			}
		}
	}

	if a.SourceDocuments != nil && a.SourceDocuments.MovementOfGoods != nil {
		documents := make(map[string]bool)
		for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
			// DocumentNumberConstraint
			if _, ok := documents[stock.DocumentNumber]; !ok {
				return errcodes.ErrUQDocumentNo
			}
			documents[stock.DocumentNumber] = true

			// StockMovementCustomerIDConstraint
			if _, ok := customers[*stock.CustomerId]; !ok {
				return errcodes.ErrKRStockMovementCustomerID
			}

			// StockMovementSupplierIDConstraint
			if _, ok := suppliers[*stock.SupplierId]; !ok {
				return errcodes.ErrKRStockMovementSupplierID
			}

			// StockMovementProductCodeConstraint
			for _, line := range stock.Line {
				if _, ok := products[line.ProductCode]; !ok {
					return errcodes.ErrKRStockMovementProductCode
				}
			}
		}
	}

	if a.SourceDocuments != nil && a.SourceDocuments.WorkingDocuments != nil {
		workDocs := make(map[string]bool)
		for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
			// WorkDocumentDocumentNumberConstraint
			if _, ok := workDocs[workDoc.DocumentNumber]; !ok {
				return errcodes.ErrUQWorkDocNo
			}
			workDocs[workDoc.DocumentNumber] = true

			// WorkDocumentDocumentCustomerIDConstraint
			if _, ok := customers[workDoc.CustomerId]; !ok {
				return errcodes.ErrKRWorkDocumentCustomerID
			}

			// WorkDocumentDocumentProductCodeConstraint
			for _, line := range workDoc.Line {
				if _, ok := products[line.ProductCode]; !ok {
					return errcodes.ErrKRWorkDocumentProductCode
				}
			}
		}
	}

	if a.SourceDocuments != nil && a.SourceDocuments.Payments != nil {
		payments := make(map[string]bool)
		for _, payment := range a.SourceDocuments.Payments.Payment {
			// PaymentPaymentRefNoConstraint

			if _, ok := payments[payment.PaymentRefNo]; ok {
				return errcodes.ErrUQPaymentRefNo
			}
			payments[payment.PaymentRefNo] = true

			// PaymentPaymentRefNoCustomerIDConstraint
			if _, ok := customers[payment.CustomerId]; !ok {
				return errcodes.ErrKRStockMovementCustomerID
			}

			// Check TransactionId against GeneralLedgerEntries
			// Not an official constraint
			if payment.TransactionId != nil && *payment.TransactionId != "" {
				if len(transactions) != 0 {
					if _, ok := transactions[*payment.TransactionId]; !ok {
						return errcodes.ErrKRPaymentTransactionId
					}
				}

			}
		}
	}

	return nil
}
