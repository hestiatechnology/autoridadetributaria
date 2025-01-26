package saft

import (
	"regexp"
	"strconv"

	"github.com/hestiatechnology/autoridadetributaria/common"
	err "github.com/hestiatechnology/autoridadetributaria/saft/err"
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
				return err.ErrGroupingCategoryTaxonomyCode
			}
			// Test 2
			if (account.GroupingCategory == "GR" && account.GroupingCode != nil) ||
				(account.GroupingCategory == "AR" && account.GroupingCode != nil) ||
				(account.GroupingCategory == "GA" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "AA" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "GM" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "AM" && account.GroupingCode == nil) {
				// logger.DebugLogger.Println("Err: Invalid GroupingCategory and GroupingCode combination: ", account.GroupingCategory, *account.GroupingCode)
				return err.ErrGroupingCategoryGroupingCode
			}
		}
	}

	for _, customer := range a.MasterFiles.Customer {
		r := regexp.MustCompile(`([^^]*)`)
		// Check if it doesnt match Desconhecido or the regex
		if customer.AccountId != "Desconhecido" && !r.MatchString(customer.AccountId) {
			return err.ErrValidationAccountId
		}
	}

	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		if invoice.SpecialRegimes.CashVatschemeIndicator != CashVatschemeIndicatorNo && invoice.SpecialRegimes.CashVatschemeIndicator != CashVatschemeIndicatorYes {
			return err.ErrCashVatschemeIndicator
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
		return err.ErrUnsupportedSAFTVersion
	}

	// CompanyId must be either a NIF (9 numbers) or Conservatória (string) + NIF
	r := regexp.MustCompile(`([0-9]{9})+|([^^]+ [0-9/]+)`)
	if !r.MatchString(a.Header.CompanyId) {
		return err.ErrCompanyId
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
		return err.ErrInvalidSaftType
	}

	if a.Header.CompanyName == "" {
		return err.ErrEmptyCompanyName
	}

	// Check if CompanyAddress is empty
	if a.Header.CompanyAddress == (AddressStructure{}) {
		return err.ErrEmptyCompanyAddress
	}

	if a.Header.CompanyAddress.AddressDetail == "" {
		return err.ErrEmptyAddressDetail
	}

	if a.Header.CompanyAddress.City == "" {
		return err.ErrEmptyCity
	}

	if a.Header.CompanyAddress.PostalCode == "" {
		return err.ErrEmptyPostalCode
	}

	if a.Header.CompanyAddress.Country != "PT" {
		return err.ErrInvalidCountry
	}

	if a.Header.FiscalYear == "" {
	}

	return nil
}

func (a *AuditFile) CheckConstraints() error {
	// AccountIDConstraint
	accounts := make(map[SafptglaccountId]bool)
	for _, account := range a.MasterFiles.GeneralLedgerAccounts.Account {
		if _, ok := accounts[account.AccountId]; ok {
			return err.ErrUQAccountId
		}
		accounts[account.AccountId] = true
	}

	// GroupingCodeConstraint
	for _, account := range a.MasterFiles.GeneralLedgerAccounts.Account {
		if account.GroupingCode != nil && *account.GroupingCode != "" {
			if _, ok := accounts[SafptglaccountId(*account.GroupingCode)]; !ok {
				return err.ErrKRGenLedgerEntriesAccountID
			}
		}
	}

	// CustomerIDConstraint
	customers := make(map[SafpttextTypeMandatoryMax30Car]bool)
	for _, customer := range a.MasterFiles.Customer {
		if _, ok := customers[customer.CustomerId]; ok {
			return err.ErrUQCustomerId
		}
		customers[customer.CustomerId] = true
	}

	// SupplierIDConstraint
	suppliers := make(map[SafpttextTypeMandatoryMax30Car]bool)
	for _, supplier := range a.MasterFiles.Supplier {
		if _, ok := suppliers[supplier.SupplierId]; ok {
			return err.ErrUQSupplierId
		}
		suppliers[supplier.SupplierId] = true
	}

	// ProductCodeConstraint
	products := make(map[SafpttextTypeMandatoryMax60Car]bool)
	for _, product := range a.MasterFiles.Product {
		if _, ok := products[product.ProductCode]; ok {
			return err.ErrUQProductCode
		}
		products[product.ProductCode] = true
	}

	// GeneralLedgerEntriesDebitLineAccountIDConstraint
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			for _, debit := range line.Lines.DebitLine {
				if _, ok := accounts[debit.AccountId]; !ok {
					return err.ErrKRGenLedgerEntriesAccountID
				}
			}
		}
	}

	// GeneralLedgerEntriesCreditLineAccountIDConstraint
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			for _, credit := range line.Lines.CreditLine {
				if _, ok := accounts[credit.AccountId]; !ok {
					return err.ErrKRGenLedgerEntriesAccountID
				}
			}
		}
	}

	// GeneralLedgerEntriesCustomerIDConstraint
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			if line.CustomerId != nil && *line.CustomerId != "" {
				if _, ok := customers[*line.CustomerId]; !ok {
					return err.ErrKRGenLedgerEntriesCustomerID
				}
			}
		}
	}

	// GeneralLedgerEntriesJournalIdConstraint
	journals := make(map[SafptjournalId]bool)
	for _, entry := range a.GeneralLedgerEntries.Journal {
		if _, ok := journals[entry.JournalId]; ok {
			return err.ErrUQJournalId
		}
		journals[entry.JournalId] = true
	}

	// GeneralLedgerEntriesSupplierIDConstraint
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			if line.SupplierId != nil && *line.SupplierId != "" {
				if _, ok := suppliers[*line.SupplierId]; !ok {
					return err.ErrKRGenLedgerEntriesSupplierID
				}
			}
		}
	}

	// GeneralLedgerEntriesTransactionIdConstraint
	transactions := make(map[SafpttransactionId]bool)
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			if _, ok := transactions[line.TransactionId]; ok {
				return err.ErrUQTransactionId
			}
			transactions[line.TransactionId] = true
		}
	}

	// InvoiceNoConstraint
	invoices := make(map[string]bool)
	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		if _, ok := invoices[invoice.InvoiceNo]; ok {
			return err.ErrUQInvoiceNo
		}
		invoices[invoice.InvoiceNo] = true
	}

	// InvoiceCustomerIDConstraint
	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		if _, ok := customers[invoice.CustomerId]; !ok {
			return err.ErrKRInvoiceCustomerID
		}
	}

	// InvoiceProductCodeConstraint
	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		for _, line := range invoice.Line {
			if _, ok := products[line.ProductCode]; !ok {
				return err.ErrKRInvoiceProductCode
			}
		}
	}

	// DocumentNumberConstraint
	documents := make(map[string]bool)
	for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
		if _, ok := documents[stock.DocumentNumber]; !ok {
			return err.ErrUQDocumentNo
		}
		documents[stock.DocumentNumber] = true
	}

	// StockMovementCustomerIDConstraint
	for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
		if _, ok := customers[*stock.CustomerId]; !ok {
			return err.ErrKRStockMovementCustomerID
		}
	}

	// StockMovementSupplierIDConstraint
	for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
		if _, ok := suppliers[*stock.SupplierId]; !ok {
			return err.ErrKRStockMovementSupplierID
		}
	}

	// StockMovementProductCodeConstraint
	for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
		for _, line := range stock.Line {
			if _, ok := products[line.ProductCode]; !ok {
				return err.ErrKRStockMovementProductCode
			}
		}
	}

	// WorkDocumentDocumentNumberConstraint
	workDocs := make(map[string]bool)
	for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
		if _, ok := workDocs[workDoc.DocumentNumber]; !ok {
			return err.ErrUQWorkDocNo
		}
		workDocs[workDoc.DocumentNumber] = true
	}

	// WorkDocumentDocumentCustomerIDConstraint
	for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
		if _, ok := customers[workDoc.CustomerId]; !ok {
			return err.ErrKRWorkDocumentCustomerID
		}
	}

	// WorkDocumentDocumentProductCodeConstraint
	for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
		for _, line := range workDoc.Line {
			if _, ok := products[line.ProductCode]; !ok {
				return err.ErrKRWorkDocumentProductCode
			}
		}
	}

	// PaymentPaymentRefNoConstraint
	payments := make(map[string]bool)
	for _, payment := range a.SourceDocuments.Payments.Payment {
		if _, ok := payments[payment.PaymentRefNo]; ok {
			return err.ErrUQPaymentRefNo
		}
		payments[payment.PaymentRefNo] = true
	}

	// PaymentPaymentRefNoCustomerIDConstraint
	for _, payment := range a.SourceDocuments.Payments.Payment {
		if _, ok := customers[payment.CustomerId]; !ok {
			return err.ErrKRStockMovementCustomerID
		}
	}

	return nil
}
