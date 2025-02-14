package main

import (
	"fmt"
	"os"

	"github.com/hestiatechnology/autoridadetributaria/saft"
)

func main() {
	a, err := saft.FromXML("real_saft.xml")
	if err != nil {
		fmt.Println(err)
	}

	// print the AuditFile
	//fmt.Println(a.SourceDocuments.SalesInvoices.Invoice[0])

	// Convert the AuditFile to XML
	xml, err := a.ToXML()
	if err != nil {
		fmt.Println(err)
	}

	// Print the XML
	fmt.Println("------------------------------------")
	fmt.Println(xml)

	// export to a file
	os.WriteFile("saf-t-export.xml", []byte(xml), 0644)

	//dec := decimal.NewFromFloat(23.0)
	//safDec := saft.SafdecimalType(dec)
	//

	//b := saft.AuditFile{SourceDocuments: &saft.SourceDocuments{
	//	SalesInvoices: &saft.SourceDocumentsSalesInvoices{
	//		Invoice: []saft.SalesInvoicesInvoice{{WithholdingTax: []saft.WithholdingTax{}}},
	//	}}}
	//xml, err = b.ToXML()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//// Print the XML
	////fmt.Println("------------------------------------")
	////fmt.Println(xml)
	//
	//// export to a file
	//os.WriteFile("saf-t-export-2.xml", []byte(xml), 0644)

	//b, err := ns.FromXML("real_saft.xml")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//// print the AuditFile
	//fmt.Println(b)
	//
	//// Convert the AuditFile to XML
	//xml, err = b.ToXML()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//// Print the XML
	//fmt.Println("------------------------------------")
	////fmt.Println(xml)
	//
	//// export to a file
	//os.WriteFile("saf-t-export-2.xml", []byte(xml), 0644)

}
