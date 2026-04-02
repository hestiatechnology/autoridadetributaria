// Package fatcorews implements a client for the AT FatCore webservice
// (e-Fatura — comunicação de documentos em tempo real).
//
// Use [types.NewFatcorewsPort] with a [soap.Client] configured with WS-Security
// (see [security.Build]) and mutual TLS to call the service.
package fatcorews

const (
	// TestURL is the AT staging endpoint for invoice communication.
	TestURL = "https://servicos.portaldasfinancas.gov.pt:723/fatcorews/ws/"
	// ProdURL is the AT production endpoint for invoice communication.
	ProdURL = "https://servicos.portaldasfinancas.gov.pt:423/fatcorews/ws/"
)
