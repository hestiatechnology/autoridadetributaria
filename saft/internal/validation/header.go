package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/common"
	"github.com/hestiatechnology/autoridadetributaria/saft"
)

func ValidateHeader(a *saft.AuditFile) error {

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
	case saft.SaftAccounting, saft.SaftInvoicingThirdParties, saft.SaftInvoicing, saft.SaftIntegrated, saft.SaftInvoicingParcial, saft.SaftPayments, saft.SaftSelfBilling, saft.SaftTransportDocuments:
		break
	default:
		return fmt.Errorf("saft: invalid Header.TaxAccountingBasis: %s", a.Header.TaxAccountingBasis)
	}

	if a.Header.CompanyName == "" {
		return fmt.Errorf("saft: missing Header.CompanyName")
	}

	// Check if CompanyAddress is empty
	if a.Header.CompanyAddress == (saft.AddressStructure{}) {
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

	if a.Header.StartDate == (saft.SafptdateSpan{}) {
		return fmt.Errorf("saft: missing Header.StartDate")
	}

	if a.Header.EndDate == (saft.SafptdateSpan{}) {
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

	if a.Header.DateCreated == (saft.SafptdateSpan{}) {
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
