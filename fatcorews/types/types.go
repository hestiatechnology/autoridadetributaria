package main

import (
	"log"
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

type InvoiceNo string

func NewInvoiceNo(value string) InvoiceNo {
	if len(value) > 60 {
		return ""
	}
	return InvoiceNo(value)
}

func main() {
	nif := NewSAFPTPortugueseVatNumber(251487547)
	log.Println(nif)
}
