package saft

import (
	"encoding/xml"
	"hestia/api/utils/logger"
	"hestia/api/utils/saft/errorcodes"
	"regexp"
)

func (a *AuditFile) Validate() error {
	if a.Header.AuditFileVersion != "1.04_01" {
		return errorcodes.ErrorUnsupportedSAFTVersion
	}

	// Check the tests of the AuditFile
	if a.MasterFiles.GeneralLedgerAccounts != nil {
		for _, account := range a.MasterFiles.GeneralLedgerAccounts.Account {
			// Test 1
			if (account.GroupingCategory == "GM" && account.TaxonomyCode == nil) || (account.GroupingCategory != "GM" && account.TaxonomyCode != nil) {
				logger.DebugLogger.Println("Error: Grouping category GM and taxonomy code must be present together")
				return errorcodes.ErrorGroupingCategoryTaxonomyCode
			}
			// Test 2
			if (account.GroupingCategory == "GR" && account.GroupingCode != nil) ||
				(account.GroupingCategory == "AR" && account.GroupingCode != nil) ||
				(account.GroupingCategory == "GA" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "AA" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "GM" && account.GroupingCode == nil) ||
				(account.GroupingCategory == "AM" && account.GroupingCode == nil) {
				logger.DebugLogger.Println("Error: Invalid GroupingCategory and GroupingCode combination: ", account.GroupingCategory, *account.GroupingCode)
				return errorcodes.ErrorGroupingCategoryGroupingCode
			}
		}
	}

	for _, customer := range a.MasterFiles.Customer {
		r := regexp.MustCompile(`([^^]*)`)
		// Check if it doesnt match Desconhecido or the regex
		if customer.AccountId != "Desconhecido" && !r.MatchString(customer.AccountId) {
			return errorcodes.ErrorValidationAccountId
		}
	}

	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		if invoice.SpecialRegimes.CashVatschemeIndicator != CashVatschemeIndicatorNo && invoice.SpecialRegimes.CashVatschemeIndicator != CashVatschemeIndicatorYes {
			return errorcodes.ErrorCashVatschemeIndicator
		}
	}

	if err := a.CheckConstraints(); err != nil {
		return err
	}
	return nil
}

func (a *AuditFile) CheckConstraints() error {
	// AccountIDConstraint
	accounts := make(map[SafptglaccountId]bool)
	for _, account := range a.MasterFiles.GeneralLedgerAccounts.Account {
		if _, ok := accounts[account.AccountId]; ok {
			return errorcodes.ErrorUQAccountId
		}
		accounts[account.AccountId] = true
	}

	// GroupingCodeConstraint
	for _, account := range a.MasterFiles.GeneralLedgerAccounts.Account {
		if account.GroupingCode != nil && *account.GroupingCode != "" {
			if _, ok := accounts[SafptglaccountId(*account.GroupingCode)]; !ok {
				return errorcodes.ErrorKRGenLedgerEntriesAccountID
			}
		}
	}

	// CustomerIDConstraint
	customers := make(map[SafpttextTypeMandatoryMax30Car]bool)
	for _, customer := range a.MasterFiles.Customer {
		if _, ok := customers[customer.CustomerId]; ok {
			return errorcodes.ErrorUQCustomerId
		}
		customers[customer.CustomerId] = true
	}

	// SupplierIDConstraint
	suppliers := make(map[SafpttextTypeMandatoryMax30Car]bool)
	for _, supplier := range a.MasterFiles.Supplier {
		if _, ok := suppliers[supplier.SupplierId]; ok {
			return errorcodes.ErrorUQSupplierId
		}
		suppliers[supplier.SupplierId] = true
	}

	// ProductCodeConstraint
	products := make(map[SafpttextTypeMandatoryMax60Car]bool)
	for _, product := range a.MasterFiles.Product {
		if _, ok := products[product.ProductCode]; ok {
			return errorcodes.ErrorUQProductCode
		}
		products[product.ProductCode] = true
	}

	// GeneralLedgerEntriesDebitLineAccountIDConstraint
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			for _, debit := range line.Lines.DebitLine {
				if _, ok := accounts[debit.AccountId]; !ok {
					return errorcodes.ErrorKRGenLedgerEntriesAccountID
				}
			}
		}
	}

	// GeneralLedgerEntriesCreditLineAccountIDConstraint
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			for _, credit := range line.Lines.CreditLine {
				if _, ok := accounts[credit.AccountId]; !ok {
					return errorcodes.ErrorKRGenLedgerEntriesAccountID
				}
			}
		}
	}

	// GeneralLedgerEntriesCustomerIDConstraint
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			if line.CustomerId != nil && *line.CustomerId != "" {
				if _, ok := customers[*line.CustomerId]; !ok {
					return errorcodes.ErrorKRGenLedgerEntriesCustomerID
				}
			}
		}
	}

	// GeneralLedgerEntriesJournalIdConstraint
	journals := make(map[SafptjournalId]bool)
	for _, entry := range a.GeneralLedgerEntries.Journal {
		if _, ok := journals[entry.JournalId]; ok {
			return errorcodes.ErrorUQJournalId
		}
		journals[entry.JournalId] = true
	}

	// GeneralLedgerEntriesSupplierIDConstraint
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			if line.SupplierId != nil && *line.SupplierId != "" {
				if _, ok := suppliers[*line.SupplierId]; !ok {
					return errorcodes.ErrorKRGenLedgerEntriesSupplierID
				}
			}
		}
	}

	// GeneralLedgerEntriesTransactionIdConstraint
	transactions := make(map[SafpttransactionId]bool)
	for _, entry := range a.GeneralLedgerEntries.Journal {
		for _, line := range entry.Transaction {
			if _, ok := transactions[line.TransactionId]; ok {
				return errorcodes.ErrorUQTransactionId
			}
			transactions[line.TransactionId] = true
		}
	}

	// InvoiceNoConstraint
	invoices := make(map[string]bool)
	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		if _, ok := invoices[invoice.InvoiceNo]; ok {
			return errorcodes.ErrorUQInvoiceNo
		}
		invoices[invoice.InvoiceNo] = true
	}

	// InvoiceCustomerIDConstraint
	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		if _, ok := customers[invoice.CustomerId]; !ok {
			return errorcodes.ErrorKRInvoiceCustomerID
		}
	}

	// InvoiceProductCodeConstraint
	for _, invoice := range a.SourceDocuments.SalesInvoices.Invoice {
		for _, line := range invoice.Line {
			if _, ok := products[line.ProductCode]; !ok {
				return errorcodes.ErrorKRInvoiceProductCode
			}
		}
	}

	// DocumentNumberConstraint
	documents := make(map[string]bool)
	for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
		if _, ok := documents[stock.DocumentNumber]; !ok {
			return errorcodes.ErrorUQDocumentNo
		}
		documents[stock.DocumentNumber] = true
	}

	// StockMovementCustomerIDConstraint
	for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
		if _, ok := customers[*stock.CustomerId]; !ok {
			return errorcodes.ErrorKRStockMovementCustomerID
		}
	}

	// StockMovementSupplierIDConstraint
	for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
		if _, ok := suppliers[*stock.SupplierId]; !ok {
			return errorcodes.ErrorKRStockMovementSupplierID
		}
	}

	// StockMovementProductCodeConstraint
	for _, stock := range a.SourceDocuments.MovementOfGoods.StockMovement {
		for _, line := range stock.Line {
			if _, ok := products[line.ProductCode]; !ok {
				return errorcodes.ErrorKRStockMovementProductCode
			}
		}
	}

	// WorkDocumentDocumentNumberConstraint
	workDocs := make(map[string]bool)
	for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
		if _, ok := workDocs[workDoc.DocumentNumber]; !ok {
			return errorcodes.ErrorUQWorkDocNo
		}
		workDocs[workDoc.DocumentNumber] = true
	}

	// WorkDocumentDocumentCustomerIDConstraint
	for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
		if _, ok := customers[workDoc.CustomerId]; !ok {
			return errorcodes.ErrorKRWorkDocumentCustomerID
		}
	}

	// WorkDocumentDocumentProductCodeConstraint
	for _, workDoc := range a.SourceDocuments.WorkingDocuments.WorkDocument {
		for _, line := range workDoc.Line {
			if _, ok := products[line.ProductCode]; !ok {
				return errorcodes.ErrorKRWorkDocumentProductCode
			}
		}
	}

	// PaymentPaymentRefNoConstraint
	payments := make(map[string]bool)
	for _, payment := range a.SourceDocuments.Payments.Payment {
		if _, ok := payments[payment.PaymentRefNo]; ok {
			return errorcodes.ErrorUQPaymentRefNo
		}
		payments[payment.PaymentRefNo] = true
	}

	// PaymentPaymentRefNoCustomerIDConstraint
	for _, payment := range a.SourceDocuments.Payments.Payment {
		if _, ok := customers[payment.CustomerId]; !ok {
			return errorcodes.ErrorKRStockMovementCustomerID
		}
	}

	return nil
}

func (a *AuditFile) CheckTypes() (string, error) {
	// Check the types of the AuditFile
	return "", nil
}

func (a *AuditFile) ExportInvoicing() (string, error) {
	a.MasterFiles.GeneralLedgerAccounts = nil
	a.GeneralLedgerEntries = nil
	a.MasterFiles.Supplier = nil

	// Export the AuditFile to a file
	if err := a.Validate(); err != nil {
		return "", err
	}

	return "", nil
}

func (a *AuditFile) ToXML() (string, error) {
	// Convert the AuditFile to XML
	if err := a.Validate(); err != nil {
		return "", err
	}

	out, err := xml.MarshalIndent(a, "", "    ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(out), nil
}
