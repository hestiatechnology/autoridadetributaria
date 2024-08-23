package main

import (
	"encoding/xml"
	"log"
	"regexp"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/common"
	"github.com/shopspring/decimal"
)

// Campo para indicacao da mais recente versao do SAF-T
// que consegue criar o software que processa as faturas
// e comunica por webservice.
type AuditFileVersion string

const (
	Version1_04_01     AuditFileVersion = "1.04_01"
	Version1_03_01     AuditFileVersion = "1.03_01"
	Version1_02_01     AuditFileVersion = "1.02_01"
	Version1_01_01     AuditFileVersion = "1.01_01"
	Version1_00_01     AuditFileVersion = "1.00_01"
	VersionInexistente AuditFileVersion = "inexistente"
)

func SetAuditFileVersion(value string) AuditFileVersion {
	switch value {
	case "1.04_01":
		return Version1_04_01
	case "1.03_01":
		return Version1_03_01
	case "1.02_01":
		return Version1_02_01
	case "1.01_01":
		return Version1_01_01
	case "1.00_01":
		return Version1_00_01
	default:
		return VersionInexistente
	}
}

//const eFaturaMDVersion = "0.0.1"

type SAFPTPortugueseVatNumber uint

func NewSAFPTPortugueseVatNumber(nif uint) SAFPTPortugueseVatNumber {
	if nif > 999999999 || nif < 100000000 {
		return 0
	}
	nifSplit := make([]uint, 9)
	nifCopy := nif
	for i := 8; i >= 0; i-- {
		nifSplit[i] = nifCopy % 10
		nifCopy /= 10
	}

	// Calculo do digito de controlo
	checkDigit := uint(0)
	for i := 0; i < 8; i++ {
		checkDigit += nifSplit[i] * (10 - uint(i) - 1)
	}

	checkDigit = 11 - (checkDigit % 11)
	// Se der 10 então o dígito de controlo tem de ser 0
	if checkDigit >= 10 {
		checkDigit = 0
	}
	log.Println(checkDigit)

	if int(checkDigit) == int(nifSplit[8]) {
		return SAFPTPortugueseVatNumber(nif)
	}
	return 0
}

type SAFPTtextTypeMandatoryMax20Car string

func NewSAFPTtextTypeMandatoryMax20Car(value string) SAFPTtextTypeMandatoryMax20Car {
	if len(value) > 20 {
		return ""
	}
	return SAFPTtextTypeMandatoryMax20Car(value)
}

type SAFPTtextTypeMandatoryMax30Car string

func NewSAFPTtextTypeMandatoryMax30Car(value string) SAFPTtextTypeMandatoryMax30Car {
	if len(value) > 30 {
		return ""
	}
	return SAFPTtextTypeMandatoryMax30Car(value)
}

type SAFPTtextTypeMandatoryMax60Car string

func NewSAFPTtextTypeMandatoryMax60Car(value string) SAFPTtextTypeMandatoryMax60Car {
	if len(value) > 60 {
		return ""
	}
	return SAFPTtextTypeMandatoryMax60Car(value)
}

type SAFPTtextTypeMandatoryMax100Car string

func NewSAFPTtextTypeMandatoryMax100Car(value string) SAFPTtextTypeMandatoryMax100Car {
	if len(value) > 100 {
		return ""
	}
	return SAFPTtextTypeMandatoryMax100Car(value)
}

func validateDocNo(value string) bool {
	const regex = `^[^ ]+ [^\/^ ]+\/[0-9]+`
	r := regexp.MustCompile(regex)
	return r.MatchString(value)
}

type InvoiceNo string

func NewInvoiceNo(value string) InvoiceNo {
	if len(value) > 60 {
		return ""
	}
	if !validateDocNo(value) {
		return ""
	}
	return InvoiceNo(value)
}

type DocumentNumber string

func NewDocumentNumber(value string) DocumentNumber {
	if len(value) > 60 {
		return ""
	}
	if !validateDocNo(value) {
		return ""
	}
	return DocumentNumber(value)
}

type PaymentRefNo string

func NewPaymentRefNo(value string) PaymentRefNo {
	if len(value) > 60 {
		return ""
	}
	if !validateDocNo(value) {
		return ""
	}
	return PaymentRefNo(value)
}

type InvoiceStatusLetter string

// N para Normal, A anulado, F faturado, S Autofaturado
const (
	InvoiceNormal    InvoiceStatusLetter = "N"
	InvoiceCancelled InvoiceStatusLetter = "A"
	InvoiceFaturado  InvoiceStatusLetter = "F"
	InvoiceSelf      InvoiceStatusLetter = "S"
)

func NewInvoiceStatusLetter(value string) InvoiceStatusLetter {
	switch value {
	case "N":
		return InvoiceNormal
	case "A":
		return InvoiceCancelled
	case "F":
		return InvoiceFaturado
	case "S":
		return InvoiceSelf
	default:
		return ""
	}
}

type InvoiceStatusDate time.Time

type InvoiceStatus struct {
	XMLName           xml.Name            `xml:"InvoiceStatus"` // XMLName is used to set the name of the XML element
	InvoiceStatus     InvoiceStatusLetter `xml:"InvoiceStatus"`
	InvoiceStatusDate InvoiceStatusDate   `xml:"InvoiceStatusDate"`
}

type StatusTypeNAF string

// N para Normal, A anulado, F faturado

const (
	WorkNormal    StatusTypeNAF = "N"
	WorkCancelled StatusTypeNAF = "A"
	WorkFaturado  StatusTypeNAF = "F"
)

func NewStatusTypeNAF(value string) StatusTypeNAF {
	switch value {
	case "N":
		return WorkNormal
	case "A":
		return WorkCancelled
	case "F":
		return WorkFaturado
	default:
		return ""
	}
}

type WorkStatusDate time.Time

type WorkStatus struct {
	XMLName        xml.Name       `xml:"WorkStatus"` // XMLName is used to set the name of the XML element
	WorkStatus     StatusTypeNAF  `xml:"WorkStatus"`
	WorkStatusDate WorkStatusDate `xml:"WorkStatusDate"`
}

type PaymentStatusLetter string

// N para Normal, A anulado
const (
	PaymentNormal    PaymentStatusLetter = "N"
	PaymentCancelled PaymentStatusLetter = "A"
)

func NewPaymentStatusLetter(value string) PaymentStatusLetter {
	switch value {
	case "N":
		return PaymentNormal
	case "A":
		return PaymentCancelled
	default:
		return ""
	}
}

type PaymentStatusDate time.Time

type PaymentStatus struct {
	XMLName           xml.Name            `xml:"PaymentStatus"`
	PaymentStatus     PaymentStatusLetter `xml:"PaymentStatus"`
	PaymentStatusDate PaymentStatusDate   `xml:"PaymentStatusDate"`
}

type NewInvoiceStatusType struct {
	XMLName           xml.Name      `xml:"NewInvoiceStatusType"`
	InvoiceStatus     StatusTypeNAF `xml:"InvoiceStatus"`
	InvoiceStatusDate time.Time     `xml:"InvoiceStatusDate"`
}

type NewWorkStatusType struct {
	XMLName        xml.Name      `xml:"NewWorkStatusType"`
	WorkStatus     StatusTypeNAF `xml:"WorkStatus"`
	WorkStatusDate time.Time     `xml:"WorkStatusDate"`
}

type NewPaymentStatusType struct {
	XMLName           xml.Name            `xml:"NewPaymentStatusType"`
	PaymentStatus     PaymentStatusLetter `xml:"PaymentStatus"`
	PaymentStatusDate time.Time           `xml:"PaymentStatusDate"`
}

type HashCharaters string

func NewHashCharaters(value string) HashCharaters {
	r := regexp.MustCompile(`[0]|[^^]{4}`)
	if !r.MatchString(value) {
		return ""
	}
	return HashCharaters(value)
}

type InvoiceTypeType string

/* Restricao: Tipos de Documento fatura (
FT - Fatura,
NC - Nota de Credito,
ND - Nota de Debito,
FS - Fatura Simplificada,
FR - Fatura-recibo).
Para o setor Segurador (a), ainda pode ser preenchido com:
RP para Premio ou recibo de premio,
RE para Estorno ou recibo de estorno,
CS para Imputacao a co-seguradoras,
LD para Imputacao a co-seguradora lider,
RA para Resseguro aceite.
(a) Para os dados ate 2019-12-31. */

const (
	InvoiceFT InvoiceTypeType = "FT"
	InvoiceNC InvoiceTypeType = "NC"
	InvoiceND InvoiceTypeType = "ND"
	InvoiceFS InvoiceTypeType = "FS"
	InvoiceFR InvoiceTypeType = "FR"
	InvoiceRP InvoiceTypeType = "RP"
	InvoiceRE InvoiceTypeType = "RE"
	InvoiceCS InvoiceTypeType = "CS"
	InvoiceLD InvoiceTypeType = "LD"
	InvoiceRA InvoiceTypeType = "RA"
)

func NewInvoiceTypeType(value string) InvoiceTypeType {
	switch value {
	case "FT":
		return InvoiceFT
	case "NC":
		return InvoiceNC
	case "ND":
		return InvoiceND
	case "FS":
		return InvoiceFS
	case "FR":
		return InvoiceFR
	case "RP":
		return InvoiceRP
	case "RE":
		return InvoiceRE
	case "CS":
		return InvoiceCS
	case "LD":
		return InvoiceLD
	case "RA":
		return InvoiceRA
	default:
		return ""
	}
}

type WorkTypeType string

/* Deve ser preenchido com:
CM Consultas de mesa;
CC Credito de consignacao;
FC Fatura de consignacao nos termos do art 38 do codigo do IVA;
FO Folhas de obra;
NE Nota de Encomenda;
OU Outros;
OR Orcamentos;
PF Proforma.
Para o setor Segurador, ainda pode ser preenchido com:
RP para Premio ou recibo de premio,
RE para Estorno ou recibo de estorno,
CS para Imputacao a co-seguradoras,
LD para Imputacao a co-seguradora lider,
RA para Resseguro aceite. */

const (
	WorkCM WorkTypeType = "CM"
	WorkCC WorkTypeType = "CC"
	WorkFC WorkTypeType = "FC"
	WorkFO WorkTypeType = "FO"
	WorkNE WorkTypeType = "NE"
	WorkOU WorkTypeType = "OU"
	WorkOR WorkTypeType = "OR"
	WorkPF WorkTypeType = "PF"
	WorkRP WorkTypeType = "RP"
	WorkRE WorkTypeType = "RE"
	WorkCS WorkTypeType = "CS"
	WorkLD WorkTypeType = "LD"
	WorkRA WorkTypeType = "RA"
)

func NewWorkTypeType(value string) WorkTypeType {
	switch value {
	case "CM":
		return WorkCM
	case "CC":
		return WorkCC
	case "FC":
		return WorkFC
	case "FO":
		return WorkFO
	case "NE":
		return WorkNE
	case "OU":
		return WorkOU
	case "OR":
		return WorkOR
	case "PF":
		return WorkPF
	case "RP":
		return WorkRP
	case "RE":
		return WorkRE
	case "CS":
		return WorkCS
	case "LD":
		return WorkLD
	case "RA":
		return WorkRA
	default:
		return ""
	}
}

type PaymentTypeType string

/*
	Deve ser preenchido com:

RC Recibo emitido no ambito do regime de IVA de Caixa
(incluindo os relativos a adiantamentos desse regime)
*/
const (
	PaymentRC PaymentTypeType = "RC"
)

// Only exists one type of payment so always return PaymentRC
func NewPaymentTypeType() PaymentTypeType {
	return PaymentRC
}

type DebitCreditIndicator string

// D para Débito, C para Crédito
const (
	Debit  DebitCreditIndicator = "D"
	Credit DebitCreditIndicator = "C"
)

func NewDebitCreditIndicator(value string) DebitCreditIndicator {
	switch value {
	case "D":
		return Debit
	case "C":
		return Credit
	default:
		return ""
	}
}

// CAE code (Código de Atividade Económica)
type EACCode string

func NewEACCode(value string) EACCode {
	if len(value) != 5 {
		return ""
	}
	r := regexp.MustCompile(`(([0-9]*))`)
	if !r.MatchString(value) {
		return ""
	}
	return EACCode(value)
}

// Self-biling regime (regime de autofaturação)
type SelfBillingIndicator uint

/*
Deve ser preenchido com “1” se respeitar a
autofacturação e com “0” (zero) no caso contrário.
*/
const (
	SelfBillingYes SelfBillingIndicator = 1
	SelfBillingNo  SelfBillingIndicator = 0
)

func NewSelfBillingIndicator(value uint) SelfBillingIndicator {
	if value == 1 {
		return SelfBillingYes
	}
	return SelfBillingNo
}

/*
	Indicador da existência de adesão ao regime de IVA de

Caixa.
Deve ser preenchido com “1” se houver adesão e com “0”
(zero) no caso contrário.
*/
type CashVATSchemeIndicator uint

const (
	CashVATSchemeYes CashVATSchemeIndicator = 1
	CashVATSchemeNo  CashVATSchemeIndicator = 0
)

func NewCashVATSchemeIndicator(value uint) CashVATSchemeIndicator {
	if value == 1 {
		return CashVATSchemeYes
	}
	return CashVATSchemeNo
}

/*
Indicador da emissão da fatura sem papel.
Deve ser preenchido com “1” se a fatura for emitida sem
papel e com “0” (zero) no caso contrário.
*/
type PaperLessIndicator uint

const (
	PaperLessYes PaperLessIndicator = 1
	PaperLessNo  PaperLessIndicator = 0
)

func NewPaperLessIndicator(value uint) PaperLessIndicator {
	if value == 1 {
		return PaperLessYes
	}
	return PaperLessNo
}

type TaxIDCountry string

func NewTaxIDCountry(value string) TaxIDCountry {
	if value == "Desconhecido" {
		return TaxIDCountry(value)
	}
	for _, country := range common.CountryCodes {
		if value == country {
			return TaxIDCountry(value)
		}
	}
	return ""
}

type OriginatingON SAFPTtextTypeMandatoryMax60Car
type InvoiceDate time.Time

type SourceDocumentID struct {
	XMLName       xml.Name      `xml:"SourceDocumentID"`
	OriginatingON OriginatingON `xml:"OriginatingON"`
	InvoiceDate   InvoiceDate   `xml:"InvoiceDate"`
}

type OrderReferences struct {
	XMLName       xml.Name      `xml:"OrderReferences"`
	OriginatingON OriginatingON `xml:"OriginatingON"`
	OrderDate     time.Time     `xml:"OrderDate"`
}

type MonetaryType decimal.Decimal

// If value is negative or greater than 9999999999999.99 returns -1.
// AT: Elemento do SAF-T alterado, com a inclusão do valor
// máximo permitido.
func NewMonetaryType(value decimal.Decimal) MonetaryType {
	if value.IsNegative() || value.GreaterThan(decimal.NewFromFloat(9999999999999.99)) {
		return MonetaryType(decimal.NewFromFloat(float64(-1)))
	}
	return MonetaryType(value)
}

// Either "IVA" or "IS" or "NS"
type TaxType string

const (
	TaxIVA TaxType = "IVA"
	TaxIS  TaxType = "IS"
	TaxNS  TaxType = "NS"
)

func NewTaxType(value string) TaxType {
	switch value {
	case "IVA":
		return TaxIVA
	case "IS":
		return TaxIS
	case "NS":
		return TaxNS
	default:
		return ""
	}
}

// ISO 3166-1 alpha-2 country code
type TaxCountryRegion string

func NewTaxCountryRegion(value string) TaxCountryRegion {
	if value == "PT-MA" || value == "PT-AC" {
		return TaxCountryRegion(value)
	}

	// Avoid using the regex provided by AT
	// Regex is 600% slower than this way
	for _, country := range common.CountryCodes {
		if value == country {
			return TaxCountryRegion(value)
		}
	}

	return ""
}

type TaxCode string

const (
	TaxCodeReduced    TaxCode = "RED"
	TaxCodeInterm     TaxCode = "INT"
	TaxCodeNormal     TaxCode = "NOR"
	TaxCodeExempt     TaxCode = "ISE"
	TaxCodeOther      TaxCode = "OUT"
	TaxCodeNotSubject TaxCode = "NS"
	// ??? Unsure about this one, it's one the WSDL but on the
	// pdf documentation it's not present
	TaxCodeNA TaxCode = "NA"
)

// Returns the TaxCode based on the value
func NewTaxCode(value string) TaxCode {
	if len(value) > 10 || len(value) < 1 {
		return ""
	}
	if value == "RED" {
		return TaxCodeReduced
	} else if value == "INT" {
		return TaxCodeInterm
	} else if value == "NOR" {
		return TaxCodeNormal
	} else if value == "ISE" {
		return TaxCodeExempt
	} else if value == "OUT" {
		return TaxCodeOther
	} else if value == "NS" {
		return TaxCodeNotSubject
	} else if value == "NA" {
		return TaxCodeNA
	} else {
		r := regexp.MustCompile(`([a-zA-Z0-9.])*`)
		if r.MatchString(value) {
			return TaxCode(value)
		}
	}
	return ""
}

type PercentageType decimal.Decimal

// If value is negative or greater than 100 returns -1.
func NewPercentageType(value decimal.Decimal) PercentageType {
	if value.IsNegative() || value.GreaterThan(decimal.NewFromInt(100)) {
		return PercentageType(decimal.NewFromFloat(float64(-1)))
	}
	return PercentageType(value)
}

// Code is of type M[0-9]{2}
type TaxExemptionCode string

func NewTaxExemptionCode(value string) TaxExemptionCode {
	for _, code := range common.VatExemptionCodes {
		if value == code {
			return TaxExemptionCode(value)
		}
	}
	return ""
}

type DocumentTotals struct {
	XMLName    xml.Name     `xml:"DocumentTotals"`
	TaxPayable MonetaryType `xml:"TaxPayable"`
	NetTotal   MonetaryType `xml:"NetTotal"`
	GrossTotal MonetaryType `xml:"GrossTotal"`
}

type WithholdingTaxType string

const (
	WithholdingTaxIRS WithholdingTaxType = "IRS"
	WithholdingTaxIRC WithholdingTaxType = "IRC"
	WithholdingTaxIS  WithholdingTaxType = "IS"
)

func NewWithholdingTaxType(value string) WithholdingTaxType {
	switch value {
	case "IRS":
		return WithholdingTaxIRS
	case "IRC":
		return WithholdingTaxIRC
	case "IS":
		return WithholdingTaxIS
	default:
		return ""
	}
}

type WithholdingTax struct {
	XMLName xml.Name `xml:"WithholdingTax"`
	// optional
	WithholdingTaxType   WithholdingTaxType `xml:"WithholdingTax"`
	WithholdingTaxAmount MonetaryType       `xml:"WithholdingTaxAmount"`
}

type DateRangeType struct {
	XMLName   xml.Name  `xml:"DateRangeType"`
	StartDate time.Time `xml:"StartDate"`
	EndDate   time.Time `xml:"EndDate"`
}

type InvoiceType InvoiceTypeType
type ATCUD SAFPTtextTypeMandatoryMax100Car
type CustomerTaxID SAFPTtextTypeMandatoryMax30Car
type CustomerTaxIDCountry TaxIDCountry
type WorkDate time.Time
type WorkType WorkTypeType

type InvoiceHeaderType struct {
	XMLName              xml.Name             `xml:"InvoiceHeaderType"`
	InvoiceNo            InvoiceNo            `xml:"InvoiceNo"`
	InvoiceDate          InvoiceDate          `xml:"InvoiceDate"`
	InvoiceType          InvoiceType          `xml:"InvoiceType"`
	SelfBillingIndicator SelfBillingIndicator `xml:"SelfBillingIndicator"`
	CustomerTaxID        CustomerTaxID        `xml:"CustomerTaxID"`
	CustomerTaxIDCountry CustomerTaxIDCountry `xml:"CustomerTaxIDCountry"`
}

type ListInvoicesDocumentsType struct {
	XMLName xml.Name `xml:"ListInvoicesDocumentsType"`
	//mandatory at least one
	Invoice []InvoiceHeaderType `xml:"invoice"`
}

type WorkHeaderType struct {
	XMLName              xml.Name             `xml:"WorkHeaderType"`
	DocumentNumber       DocumentNumber       `xml:"DocumentNumber"`
	ATCUD                ATCUD                `xml:"ATCUD"`
	WorkDate             WorkDate             `xml:"WorkDate"`
	WorkType             WorkType             `xml:"WorkType"`
	CustomerTaxID        CustomerTaxID        `xml:"CustomerTaxID"`
	CustomerTaxIDCountry CustomerTaxIDCountry `xml:"CustomerTaxIDCountry"`
}

type ListWorkDocumentsType struct {
	XMLName xml.Name `xml:"ListWorkDocumentsType"`
	//mandatory at least one
	Work []WorkHeaderType `xml:"work"`
}

type TransactionDate time.Time
type PaymentType PaymentTypeType

type PaymentHeaderType struct {
	XMLName              xml.Name             `xml:"PaymentHeaderType"`
	PaymentRefNo         PaymentRefNo         `xml:"PaymentRefNo"`
	ATCUD                ATCUD                `xml:"ATCUD"`
	TransactionDate      TransactionDate      `xml:"TransactionDate"`
	PaymentType          PaymentType          `xml:"PaymentType"`
	CustomerTaxID        CustomerTaxID        `xml:"CustomerTaxID"`
	CustomerTaxIDCountry CustomerTaxIDCountry `xml:"CustomerTaxIDCountry"`
}

type ListPaymentDocumentsType struct {
	XMLName xml.Name `xml:"ListPaymentDocumentsType"`
	//mandatory at least one
	Payment []PaymentHeaderType `xml:"payment"`
}

type CodigoResposta uint
type Mensagem string

func NewMensagem(value string) Mensagem {
	if len(value) > 256 {
		return ""
	}
	return Mensagem(value)
}

type DataOperacao time.Time

type ResponseType struct {
	XMLName        xml.Name       `xml:"ResponseType"`
	CodigoResposta CodigoResposta `xml:"CodigoResposta"`
	Mensagem       Mensagem       `xml:"Mensagem"`
	DataOperacao   DataOperacao   `xml:"DataOperacao"`
}

type ChannelType struct {
	XMLName xml.Name `xml:"ChannelType"`
	// mandatory
	Sistema string `xml:"Sistema"`
	// optional
	Versao string `xml:"Versao"`
}

type EFaturaMDVersion string

func NewEFaturaMDVersion() EFaturaMDVersion {
	return EFaturaMDVersion("0.0.1")
}

type TaxRegistrationNumber SAFPTPortugueseVatNumber

func NewTaxRegistrationNumber(value uint) TaxRegistrationNumber {
	return TaxRegistrationNumber(NewSAFPTPortugueseVatNumber(value))
}

type TaxEntity SAFPTtextTypeMandatoryMax20Car

func NewTaxEntity(value string) TaxEntity {
	return TaxEntity(NewSAFPTtextTypeMandatoryMax20Car(value))
}

type InvoiceDataType struct {
	InvoiceHeaderType
	DocumentStatus         InvoiceStatus          `xml:"DocumentStatus"`
	HashCharaters          HashCharaters          `xml:"HashCharaters"`
	CashVATSchemeIndicator CashVATSchemeIndicator `xml:"CashVATSchemeIndicator"`
}

// Requests

type RegisterInvoiceRequest struct {
	XMLName                   xml.Name              `xml:"RegisterInvoiceRequest"`
	EFaturaMDVersion          EFaturaMDVersion      `xml:"eFaturaMDVersion"`
	TaxRegistrationNumber     TaxRegistrationNumber `xml:"TaxRegistrationNumber"`
	TaxEntity                 TaxEntity             `xml:"TaxEntity"`
	SoftwareCertificateNumber uint                  `xml:"SoftwareCertificateNumber"`
	InvoiceData               InvoiceDataType       `xml:"InvoiceData"`
}
