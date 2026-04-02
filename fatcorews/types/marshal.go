package types

import (
	"encoding/xml"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Date-only types (xsd:date) — format as "2006-01-02".

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

func (t TaxPointDate) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format("2006-01-02")), nil
}

// DateTime types (xsd:dateTime) — format as RFC3339.

func (t SystemEntryDate) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return []byte(time.Time(t).Format(time.RFC3339)), nil
}

// LineSummary is a named type for the inline LineSummary element in
// InvoiceDataType and WorkDataType, allowing external packages to construct values.
type LineSummary struct {
	OrderReferences      []*OrderReferences    `xml:"OrderReferences,omitempty" json:"OrderReferences,omitempty"`
	TaxPointDate         *TaxPointDate         `xml:"TaxPointDate,omitempty" json:"TaxPointDate,omitempty"`
	Reference            []*Reference          `xml:"Reference,omitempty" json:"Reference,omitempty"`
	DebitCreditIndicator *DebitCreditIndicator `xml:"DebitCreditIndicator,omitempty" json:"DebitCreditIndicator,omitempty"`
	Tax                  *Tax                  `xml:"Tax,omitempty" json:"Tax,omitempty"`
	TaxExemptionCode     *TaxExemptionCode     `xml:"TaxExemptionCode,omitempty" json:"TaxExemptionCode,omitempty"`
	TotalTaxBase         *TotalTaxBase         `xml:"TotalTaxBase,omitempty" json:"TotalTaxBase,omitempty"`
	Amount               *Amount               `xml:"Amount,omitempty" json:"Amount,omitempty"`
}

func formatDecimal(d decimal.Decimal) string {
	s := d.String()
	parts := strings.Split(s, ".")
	if len(parts) > 1 && len(parts[1]) > 2 {
		return d.String()
	}
	return d.StringFixed(2)
}

func (d MonetaryType) MarshalText() ([]byte, error) {
	return []byte(formatDecimal(decimal.Decimal(d))), nil
}

func (d *MonetaryType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}
	return e.EncodeElement(formatDecimal(decimal.Decimal(*d)), start)
}

func (d PercentageType) MarshalText() ([]byte, error) {
	return []byte(formatDecimal(decimal.Decimal(d))), nil
}

func (d *PercentageType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}
	return e.EncodeElement(formatDecimal(decimal.Decimal(*d)), start)
}
