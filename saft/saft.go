package saft

import (
	"encoding/xml"
	"os"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
)

//func (a *AuditFile) CheckTypes() (string, error) {
//	// Check the types of the AuditFile
//	return "", nil
//}

func (a *AuditFile) ExportInvoicing() (string, error) {
	a.MasterFiles.GeneralLedgerAccounts = nil
	a.GeneralLedgerEntries = nil
	a.MasterFiles.Supplier = nil

	// Export the AuditFile to a file
	if err := a.Validate(); err != nil {
		return "", err
	}

	return "", nil
}

func (a *AuditFile) ToXML() (string, error) {
	a.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"
	a.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	//a.XsiSchemaLocation = "urn:OECD:StandardAuditFile-Tax:PT_1.04_01 ../saftpt1.04_01.xsd"
	// Convert the AuditFile to XML

	// Validation must be done manually
	if err := a.Validate(); err != nil {
		return "", err
	}

	out, err := xml.MarshalIndent(a, "", "    ")
	if err != nil {
		return "", err
	}

	encoder := charmap.Windows1252.NewEncoder()
	outStr, err := encoder.String(string(out))
	if err != nil {
		return "", err
	}

	return `<?xml version="1.0" encoding="Windows-1252"?>` + "\n" + outStr, nil
}

func FromXML(xmlFile string) (*AuditFile, error) {
	// Read the XML file
	file, err := os.Open(xmlFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Unmarshal the XML file
	a := &AuditFile{}
	decoder := xml.NewDecoder(file)
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(a); err != nil {
		return nil, err
	}

	// Validate the AuditFile
	//f err := a.Validate(); err != nil {
	//	return nil, err
	//
	return a, nil
}
