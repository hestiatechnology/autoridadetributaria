package main

import (
	"fmt"

	"github.com/hestiatechnology/autoridadetributaria/saft"
)

func main() {
	a, err := saft.FromXML("saf-t.xml")
	if err != nil {
		fmt.Println(err)
	}

	// print the AuditFile
	fmt.Println(a)

	// Convert the AuditFile to XML
	xml, err := a.ToXML()
	if err != nil {
		fmt.Println(err)
	}

	// Print the XML
	fmt.Println("------------------------------------")
	fmt.Println(xml)
}
