package workDocuments

import (
	"encoding/xml"
	"github.com/shopspring/decimal"
    "github.com/hooklift/gowsdl/soap"
)

// Dates
func (t *SAFdateType) MarshalText() ([]byte, error) {
	return []byte((*soap.XSDDate)(t).ToGoTime().Format("2006-01-02")), nil
}

func (t *SAFdateTimeType) MarshalText() ([]byte, error) {
	return []byte((*soap.XSDDateTime)(t).ToGoTime().Format("2006-01-02T15:04:05")), nil
}

// Decimals
func (d SAFdecimalType) MarshalText() ([]byte, error) {
	return []byte(decimal.NewFromFloat(float64(d)).StringFixed(2)), nil
}

func (d SAFmonetaryType) MarshalText() ([]byte, error) {
	return []byte(decimal.NewFromFloat(float64(d)).StringFixed(2)), nil
}

func (d *SAFdecimalType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}
	return e.EncodeElement(decimal.NewFromFloat(float64(*d)).StringFixed(2), start)
}

func (d *SAFmonetaryType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if d == nil {
		return nil
	}
	return e.EncodeElement(decimal.NewFromFloat(float64(*d)).StringFixed(2), start)
}

// Special case for ATCUD - ensure it is never empty if it contains a value
func (t SAFPTtextTypeMandatoryMax100Car) MarshalText() ([]byte, error) {
	if string(t) == "" {
		return nil, nil
	}
	return []byte(string(t)), nil
}
