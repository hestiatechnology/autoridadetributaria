package main

import (
	"encoding/xml"
	"log"
	"regexp"
	"time"
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

/* Restricao: Tipos de Documento fatura (FT - Fatura,
NC-Nota de Credito,
ND - Nota de Debito,
FS - Fatura Simplificada,
FR - Fatura-recibo).
Para o setor Segurador (a), ainda pode ser
preenchido com:
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
	r := regexp.MustCompile(`[A-Z]{2}`)
	if !r.MatchString(value) || value != "Desconhecido" {
		return ""
	}
	return TaxIDCountry(value)
}

type OriginatingON SAFPTtextTypeMandatoryMax60Car
type InvoiceDate time.Time

type SourceDocumentID struct {
	XMLName       xml.Name      `xml:"SourceDocumentID"`
	OriginatingON OriginatingON `xml:"OriginatingON"`
	InvoiceDate   InvoiceDate   `xml:"InvoiceDate"`
}

func main() {
	nif := NewSAFPTPortugueseVatNumber(251487547)
	log.Println(nif)
}
