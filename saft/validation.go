package saft

import (
	"fmt"
	"regexp"
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
	/* 	if err := validation.checkHeader(a); err != nil {
		return err
	} */

	/*if err := a.checkCustomers(); err != nil {
		return err
	}*/

	// TODO: Check TaxTable

	/*if err := a.checkPayments(); err != nil {
		return err
	}*/
	return nil
}

/* // Check Header mandatory fields
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
} */

//func (a *AuditFile) checkCustomers() error {}

//func (a *AuditFile) checkPayments() error {}

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
