package types

import "encoding/xml"

// Add xmlns to the RegisterInvoiceRequest so the AT servers can actually read the payload
func (r RegisterInvoiceRequest) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Space = "http://factemi.at.min_financas.pt/documents"
	// We MUST define 'doc' prefix manually to map to the correct AT namespace 
	start.Attr = append(start.Attr, xml.Attr{
		Name:  xml.Name{Local: "xmlns:doc"},
		Value: "http://factemi.at.min_financas.pt/documents",
	})
	
	type alias RegisterInvoiceRequest
	return e.EncodeElement(alias(r), start)
}
