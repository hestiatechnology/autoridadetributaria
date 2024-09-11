package errorcodes

import "errors"

// UQ = Unique Constraint
// KR = Key Reference
var (
	ErrorUQAccountId                  = errors.New("saft: unique constraint violation - account id")
	ErrorUQCustomerId                 = errors.New("saft: unique constraint violation - customer id")
	ErrorUQSupplierId                 = errors.New("saft: unique constraint violation - supplier id")
	ErrorUQProductCode                = errors.New("saft: unique constraint violation - product code")
	ErrorUQJournalId                  = errors.New("saft: unique constraint violation - journal id")
	ErrorUQTransactionId              = errors.New("saft: unique constraint violation - transaction id")
	ErrorUQInvoiceNo                  = errors.New("saft: unique constraint violation - invoice no")
	ErrorUQDocumentNo                 = errors.New("saft: unique constraint violation - document no")
	ErrorUQWorkDocNo                  = errors.New("saft: unique constraint violation - work document no")
	ErrorUQPaymentRefNo               = errors.New("saft: unique constraint violation - payment ref no")
	ErrorKRGenLedgerAccountsAccountID = errors.New("saft: key reference violation - general ledger entries account id")
	ErrorKRGenLedgerEntriesSupplierID = errors.New("saft: key reference violation - general ledger entries supplier id")
	ErrorKRGenLedgerEntriesAccountID  = errors.New("saft: key reference violation - general ledger entries account id")
	ErrorKRGenLedgerEntriesCustomerID = errors.New("saft: key reference violation - general ledger entries customer id")
	ErrorKRInvoiceCustomerID          = errors.New("saft: key reference violation - invoice customer id")
	ErrorKRInvoiceProductCode         = errors.New("saft: key reference violation - invoice product code")
	ErrorKRStockMovementCustomerID    = errors.New("saft: key reference violation - stock movement customer id")
	ErrorKRStockMovementSupplierID    = errors.New("saft: key reference violation - stock movement supplier id")
	ErrorKRStockMovementProductCode   = errors.New("saft: key reference violation - stock movement product code")
	ErrorKRWorkDocumentCustomerID     = errors.New("saft: key reference violation - work document customer id")
	ErrorKRWorkDocumentProductCode    = errors.New("saft: key reference violation - work document product code")
	ErrorKRPaymentCustomerID          = errors.New("saft: key reference violation - payment customer id")
	ErrorValidationAccountId          = errors.New("saft: account id validation error")
	ErrorCashVatschemeIndicator       = errors.New("saft: cash vat scheme indicator validation error")
)

// Other errors
var (
	ErrorUnsupportedSAFTVersion       = errors.New("saft: unsupported SAFT version")
	ErrorGroupingCategoryTaxonomyCode = errors.New("saft: grouping category GM and taxonomy code must be present together")
	ErrorGroupingCategoryGroupingCode = errors.New("saft: invalid grouping category and grouping code combination")
)
