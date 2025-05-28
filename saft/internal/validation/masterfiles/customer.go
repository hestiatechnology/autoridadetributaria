package masterfiles

import (
	"fmt"

	"github.com/hestiatechnology/autoridadetributaria/common"
	"github.com/hestiatechnology/autoridadetributaria/saft"
)

func ValidateCustomers(a *saft.AuditFile) error {
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

		if customer.BillingAddress == (saft.CustomerAddressStructure{}) {
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

		if customer.SelfBillingIndicator != saft.SelfBillingIndicatorNo && customer.SelfBillingIndicator != saft.SelfBillingIndicatorYes {
			return fmt.Errorf("saft: invalid Customer.SelfBillingIndicator: %d", customer.SelfBillingIndicator)
		}
	}

	return nil
}
