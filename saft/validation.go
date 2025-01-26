package saft

import (
	"regexp"
	"strconv"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/common"
	"github.com/hestiatechnology/autoridadetributaria/saft/errcodes"
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

	if err := a.CheckConstraints(); err != nil {
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
		return common.ErrInvalidNIFPT
	}

	ok = common.ValidateNIFPT(string(a.Header.ProductCompanyTaxId))
	if !ok {
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

func (a *AuditFile) CheckConstraints() error {
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
	if a.MasterFiles.GeneralLedgerAccounts != nil {
		// AccountIDConstraint
		accounts := make(map[SafptglaccountId]bool)
		for _, account := range a.MasterFiles.GeneralLedgerAccounts.Account {
			if _, ok := accounts[account.AccountId]; ok {
				return errcodes.ErrUQAccountId
			}
			accounts[account.AccountId] = true
		}

		// GroupingCodeConstraint
		for _, account := range a.MasterFiles.GeneralLedgerAccounts.Account {
			if account.GroupingCode != nil && *account.GroupingCode != "" {
				if _, ok := accounts[SafptglaccountId(*account.GroupingCode)]; !ok {
					return errcodes.ErrKRGenLedgerEntriesAccountID
				}
			}
		}

		// GeneralLedgerEntriesDebitLineAccountIDConstraint
		for _, entry := range a.GeneralLedgerEntries.Journal {
			for _, line := range entry.Transaction {
				for _, debit := range line.Lines.DebitLine {
					if _, ok := accounts[debit.AccountId]; !ok {
						return errcodes.ErrKRGenLedgerEntriesAccountID
					}
				}
			}
		}

		// GeneralLedgerEntriesCreditLineAccountIDConstraint
		for _, entry := range a.GeneralLedgerEntries.Journal {
			for _, line := range entry.Transaction {
				for _, credit := range line.Lines.CreditLine {
					if _, ok := accounts[credit.AccountId]; !ok {
						return errcodes.ErrKRGenLedgerEntriesAccountID
					}
				}
			}
		}

		// GeneralLedgerEntriesCustomerIDConstraint
		for _, entry := range a.GeneralLedgerEntries.Journal {
			for _, line := range entry.Transaction {
				if line.CustomerId != nil && *line.CustomerId != "" {
					if _, ok := customers[*line.CustomerId]; !ok {
						return errcodes.ErrKRGenLedgerEntriesCustomerID
					}
				}
			}
		}

		// GeneralLedgerEntriesJournalIdConstraint
		journals := make(map[SafptjournalId]bool)
		for _, entry := range a.GeneralLedgerEntries.Journal {
			if _, ok := journals[entry.JournalId]; ok {
				return errcodes.ErrUQJournalId
			}
			journals[entry.JournalId] = true
		}

		// GeneralLedgerEntriesSupplierIDConstraint
		for _, entry := range a.GeneralLedgerEntries.Journal {
			for _, line := range entry.Transaction {
				if line.SupplierId != nil && *line.SupplierId != "" {
					if _, ok := suppliers[*line.SupplierId]; !ok {
						return errcodes.ErrKRGenLedgerEntriesSupplierID
					}
				}
			}
		}

		// GeneralLedgerEntriesTransactionIdConstraint
		transactions := make(map[SafpttransactionId]bool)
		for _, entry := range a.GeneralLedgerEntries.Journal {
			for _, line := range entry.Transaction {
				if _, ok := transactions[line.TransactionId]; ok {
					return errcodes.ErrUQTransactionId
				}
				transactions[line.TransactionId] = true
			}
		}

	}

	// Sales Invoices constraints
	if a.SourceDocuments != nil && a.SourceDocuments.SalesInvoices != nil {
		// InvoiceNoConstraint
		invoices := make(map[string]bool)
		for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
			if _, ok := invoices[invoice.InvoiceNo]; ok {
				return errcodes.ErrUQInvoiceNo
			}
			invoices[invoice.InvoiceNo] = true
		}

		// InvoiceCustomerIDConstraint
		for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
			if _, ok := customers[invoice.CustomerId]; !ok {
				return errcodes.ErrKRInvoiceCustomerID
			}
		}

		// InvoiceProductCodeConstraint
		for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
			for _, line := range invoice.Line {
				if _, ok := products[line.ProductCode]; !ok {
					return errcodes.ErrKRInvoiceProductCode
				}
			}
		}
	}

	if a.SourceDocuments != nil && a.SourceDocuments.MovementOfGoods != nil {
		// DocumentNumberConstraint
		documents := make(map[string]bool)
		for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
			if _, ok := documents[stock.DocumentNumber]; !ok {
				return errcodes.ErrUQDocumentNo
			}
			documents[stock.DocumentNumber] = true
		}

		// StockMovementCustomerIDConstraint
		for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
			if _, ok := customers[*stock.CustomerId]; !ok {
				return errcodes.ErrKRStockMovementCustomerID
			}
		}

		// StockMovementSupplierIDConstraint
		for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
			if _, ok := suppliers[*stock.SupplierId]; !ok {
				return errcodes.ErrKRStockMovementSupplierID
			}
		}

		// StockMovementProductCodeConstraint
		for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
			for _, line := range stock.Line {
				if _, ok := products[line.ProductCode]; !ok {
					return errcodes.ErrKRStockMovementProductCode
				}
			}
		}
	}

	if a.SourceDocuments != nil && a.SourceDocuments.WorkingDocuments != nil {
		// WorkDocumentDocumentNumberConstraint
		workDocs := make(map[string]bool)
		for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
			if _, ok := workDocs[workDoc.DocumentNumber]; !ok {
				return errcodes.ErrUQWorkDocNo
			}
			workDocs[workDoc.DocumentNumber] = true
		}

		// WorkDocumentDocumentCustomerIDConstraint
		for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
			if _, ok := customers[workDoc.CustomerId]; !ok {
				return errcodes.ErrKRWorkDocumentCustomerID
			}
		}

		// WorkDocumentDocumentProductCodeConstraint
		for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
			for _, line := range workDoc.Line {
				if _, ok := products[line.ProductCode]; !ok {
					return errcodes.ErrKRWorkDocumentProductCode
				}
			}
		}
	}

	if a.SourceDocuments != nil && a.SourceDocuments.Payments != nil {
		// PaymentPaymentRefNoConstraint
		payments := make(map[string]bool)
		for _, payment := range a.SourceDocuments.Payments.Payment {
			if _, ok := payments[payment.PaymentRefNo]; ok {
				return errcodes.ErrUQPaymentRefNo
			}
			payments[payment.PaymentRefNo] = true
		}

		// PaymentPaymentRefNoCustomerIDConstraint
		for _, payment := range a.SourceDocuments.Payments.Payment {
			if _, ok := customers[payment.CustomerId]; !ok {
				return errcodes.ErrKRStockMovementCustomerID
			}
		}
	}

	return nil
}
