package masterfiles

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/lusis/go-saft/saft"
)

// ValidateGeneralLedgerAccounts validates the GeneralLedgerAccounts section of the SAFT file
func ValidateGeneralLedgerAccounts(gla *saft.GeneralLedgerAccounts, accountIDs map[string]bool) []error {
	var errs []error

	if gla == nil {
		errs = append(errs, errors.New("GeneralLedgerAccounts is nil"))
		return errs
	}

	if gla.TaxonomyReference == "" {
		errs = append(errs, errors.New("TaxonomyReference is required"))
	}

	for _, account := range gla.Account {
		if account.AccountID == "" {
			errs = append(errs, errors.New("AccountID is required"))
		}
		if account.AccountDescription == "" {
			errs = append(errs, errors.New("AccountDescription is required"))
		}

		validGroupingCategories := map[string]bool{
			"GR": true,
			"GA": true,
			"GM": true,
			"GI": true,
			"AR": true,
		}
		if !validGroupingCategories[account.GroupingCategory] {
			errs = append(errs, fmt.Errorf("invalid GroupingCategory: %s for AccountID: %s", account.GroupingCategory, account.AccountID))
		}

		if account.GroupingCode != "" {
			if _, ok := accountIDs[account.GroupingCode]; !ok {
				errs = append(errs, fmt.Errorf("invalid GroupingCode: %s for AccountID: %s. It does not match any existing AccountID", account.GroupingCode, account.AccountID))
			}
		}

		if account.TaxonomyCode != "" {
			if _, err := strconv.Atoi(account.TaxonomyCode); err != nil {
				errs = append(errs, fmt.Errorf("invalid TaxonomyCode: %s for AccountID: %s. It must be an integer", account.TaxonomyCode, account.AccountID))
			}
		}
	}

	return errs
}
