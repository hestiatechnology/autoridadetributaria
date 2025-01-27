package errcodes

import "errors"

// UQ = Unique Constraint
// KR = Key Reference
var (
	ErrUQAccountId                  = errors.New("saft: unique constraint violation - account id")
	ErrUQCustomerId                 = errors.New("saft: unique constraint violation - customer id")
	ErrUQSupplierId                 = errors.New("saft: unique constraint violation - supplier id")
	ErrUQProductCode                = errors.New("saft: unique constraint violation - product code")
	ErrUQJournalId                  = errors.New("saft: unique constraint violation - journal id")
	ErrUQTransactionId              = errors.New("saft: unique constraint violation - transaction id")
	ErrUQInvoiceNo                  = errors.New("saft: unique constraint violation - invoice no")
	ErrUQDocumentNo                 = errors.New("saft: unique constraint violation - document no")
	ErrUQWorkDocNo                  = errors.New("saft: unique constraint violation - work document no")
	ErrUQPaymentRefNo               = errors.New("saft: unique constraint violation - payment ref no")
	ErrKRGenLedgerAccountsAccountID = errors.New("saft: key reference violation - general ledger entries account id")
	ErrKRGenLedgerEntriesSupplierID = errors.New("saft: key reference violation - general ledger entries supplier id")
	ErrKRGenLedgerEntriesAccountID  = errors.New("saft: key reference violation - general ledger entries account id")
	ErrKRGenLedgerEntriesCustomerID = errors.New("saft: key reference violation - general ledger entries customer id")
	ErrKRInvoiceCustomerID          = errors.New("saft: key reference violation - invoice customer id")
	ErrKRInvoiceProductCode         = errors.New("saft: key reference violation - invoice product code")
	ErrKRStockMovementCustomerID    = errors.New("saft: key reference violation - stock movement customer id")
	ErrKRStockMovementSupplierID    = errors.New("saft: key reference violation - stock movement supplier id")
	ErrKRStockMovementProductCode   = errors.New("saft: key reference violation - stock movement product code")
	ErrKRWorkDocumentCustomerID     = errors.New("saft: key reference violation - work document customer id")
	ErrKRWorkDocumentProductCode    = errors.New("saft: key reference violation - work document product code")
	ErrKRPaymentCustomerID          = errors.New("saft: key reference violation - payment customer id")
	ErrValidationAccountId          = errors.New("saft: account id validation error")
	ErrCashVatschemeIndicator       = errors.New("saft: cash vat scheme indicator validation error")
)

// Other errors
var (
	ErrGroupingCategoryTaxonomyCode = errors.New("saft: grouping category GM and taxonomy code must be present together")
	ErrGroupingCategoryGroupingCode = errors.New("saft: invalid grouping category and grouping code combination")
)

// Header Errors
var (
	ErrUnsupportedSAFTVersion     = errors.New("saft: unsupported SAFT version")
	ErrCompanyId                  = errors.New("saft: company id missing or invalid")
	ErrInvalidSaftType            = errors.New("saft: invalid saft type")
	ErrEmptyCompanyName           = errors.New("saft: company name is empty")
	ErrEmptyCompanyAddress        = errors.New("saft: company address is empty")
	ErrEmptyAddressDetail         = errors.New("saft: address detail is empty")
	ErrEmptyCity                  = errors.New("saft: city is empty")
	ErrEmptyPostalCode            = errors.New("saft: postal code is empty")
	ErrEmptyCountry               = errors.New("saft: country is empty")
	ErrInvalidCountry             = errors.New("saft: country is invalid")
	ErrInvalidFiscalYear          = errors.New("saft: fiscal year missing or invalid")
	ErrInvalidStartDate           = errors.New("saft: start date missing or invalid")
	ErrInvalidEndDate             = errors.New("saft: end date missing or invalid")
	ErrInvalidDateSpan            = errors.New("saft: span date invalid")
	ErrCurrencyNotEuro            = errors.New("saft: currency is not EUR")
	ErrMissingDateCreated         = errors.New("saft: date created is missing")
	ErrMissingTaxEntity           = errors.New("saft: tax entity is missing, muste be Global, Sede or the establishment name")
	ErrMissingProductCompanyTaxId = errors.New("saft: product company tax id is missing")
	ErrInvalidProductId           = errors.New("saft: product id is invalid")
	ErrMissingProductVersion      = errors.New("saft: product version is missing")
)

// Customer Errors
var (
	ErrMissingCustomerID           = errors.New("saft: customer id is missing, muste be Desconhecido or a string")
	ErrMissingAccountID            = errors.New("missing account ID")
	ErrMissingCustomerTaxID        = errors.New("missing customer tax ID")
	ErrInvalidCustomerTaxID        = errors.New("invalid customer tax ID")
	ErrMissingCompanyName          = errors.New("missing company name")
	ErrMissingBillingAddress       = errors.New("missing billing address")
	ErrMissingAddressDetail        = errors.New("missing address detail")
	ErrMissingCity                 = errors.New("missing city")
	ErrMissingPostalCode           = errors.New("missing postal code")
	ErrMissingCountry              = errors.New("missing country")
	ErrInvalidSelfBillingIndicator = errors.New("invalid self-billing indicator")
)
