package types

import (
	"time"

	"github.com/shopspring/decimal"
	"encoding/xml"
)

// Dates
func (t InvoiceStatusDate) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format(time.RFC3339)), nil
}

func (t WorkStatusDate) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format(time.RFC3339)), nil
}

func (t PaymentStatusDate) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format(time.RFC3339)), nil
}

func (t InvoiceDate) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format("2006-01-02")), nil
}

func (t WorkDate) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format("2006-01-02")), nil
}

func (t TransactionDate) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format("2006-01-02")), nil
}

func (t DataOperacao) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format(time.RFC3339)), nil
}

func (t TaxPointDate) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format("2006-01-02")), nil
}


// Decimals
func (d MonetaryType) MarshalText() ([]byte, error) {
	return []byte(decimal.Decimal(d).StringFixed(2)), nil
}

func (d PercentageType) MarshalText() ([]byte, error) {
	return []byte(decimal.Decimal(d).StringFixed(2)), nil
}

func (d *MonetaryType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}
	return e.EncodeElement(decimal.Decimal(*d).StringFixed(2), start)
}

func (d *PercentageType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}
	return e.EncodeElement(decimal.Decimal(*d).StringFixed(2), start)
}

// Wrapper for complex structs
type shadowTax struct {
	TaxType          TaxType          `xml:"TaxType"`
	TaxCountryRegion TaxCountryRegion `xml:"TaxCountryRegion"`
	TaxCode          TaxCode          `xml:"TaxCode"`
	TaxPercentage    string           `xml:"TaxPercentage,omitempty"`
	TotalTaxAmount   string           `xml:"TaxAmount,omitempty"`
}

func (t Tax) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	s := shadowTax{
		TaxType:          t.TaxType,
		TaxCountryRegion: t.TaxCountryRegion,
		TaxCode:          t.TaxCode,
	}
	if t.TaxPercentage != nil {
		s.TaxPercentage = decimal.Decimal(*t.TaxPercentage).StringFixed(2)
	}
	if t.TotalTaxAmount != nil && t.TaxPercentage == nil {
		s.TotalTaxAmount = decimal.Decimal(*t.TotalTaxAmount).StringFixed(2)
	}
	return e.EncodeElement(s, start)
}

type shadowLineSummary struct {
	OrderReferences      *OrderReferences     `xml:"OrderReferences,omitempty"`
	TaxPointDate         string               `xml:"TaxPointDate"`
	Reference            *Reference           `xml:"Reference,omitempty"`
	DebitCreditIndicator DebitCreditIndicator `xml:"DebitCreditIndicator"`
	TotalTaxBase         *string              `xml:"TotalTaxBase,omitempty"`
	Amount               string               `xml:"Amount"`
	Tax                  Tax                  `xml:"Tax"`
	TaxExemptionCode     *TaxExemptionCode    `xml:"TaxExemptionCode,omitempty"`
}

func (l LineSummary) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	s := shadowLineSummary{
		OrderReferences:      l.OrderReferences,
		Reference:            l.Reference,
		DebitCreditIndicator: l.DebitCreditIndicator,
		Tax:                  l.Tax,
		TaxExemptionCode:     l.TaxExemptionCode,
	}
	
	if !time.Time(l.TaxPointDate).IsZero() {
		s.TaxPointDate = time.Time(l.TaxPointDate).Format("2006-01-02")
	}

	if l.Amount != nil {
		s.Amount = decimal.Decimal(*l.Amount).StringFixed(2)
	}
	if l.TotalTaxBase != nil {
		v := decimal.Decimal(*l.TotalTaxBase).StringFixed(2)
		s.TotalTaxBase = &v
	}
	
	return e.EncodeElement(s, start)
}

type shadowDocumentTotals struct {
	TaxPayable string `xml:"TaxPayable"`
	NetTotal   string `xml:"NetTotal"`
	GrossTotal string `xml:"GrossTotal"`
}

func (d DocumentTotals) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	s := shadowDocumentTotals{
		TaxPayable: decimal.Decimal(d.TaxPayable).StringFixed(2),
		NetTotal:   decimal.Decimal(d.NetTotal).StringFixed(2),
		GrossTotal: decimal.Decimal(d.GrossTotal).StringFixed(2),
	}
	return e.EncodeElement(s, start)
}

type shadowInvoiceStatus struct {
	InvoiceStatus     InvoiceStatusLetter `xml:"InvoiceStatus"`
	InvoiceStatusDate string              `xml:"InvoiceStatusDate"`
}

func (is InvoiceStatus) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	s := shadowInvoiceStatus{
		InvoiceStatus: is.InvoiceStatus,
	}
	if !time.Time(is.InvoiceStatusDate).IsZero() {
		s.InvoiceStatusDate = time.Time(is.InvoiceStatusDate).Format(time.RFC3339)
	}
	return e.EncodeElement(s, start)
}
