// Package fatcorews implements a SOAP client for the AT FatCore webservice
// (e-Fatura — comunicação de documentos em tempo real).
//
// All nine operations share the same authentication mechanism: WS-Security
// UsernameToken with AES-ECB-PKCS5 password encryption (see [security] package).
//
// Usage:
//
//	// 1. Load certificates
//	clientCert, _ := security.LoadClientCertPFX("555555555.pfx", "password")
//	atPubKey,   _ := security.LoadATPublicKey("SA.cer")
//
//	// 2. Create client
//	c := fatcorews.NewClient(fatcorews.TestURL, "555555555/37", "password", clientCert, atPubKey)
//
//	// 3. Register an invoice
//	resp, err := c.RegisterInvoice(myRegisterInvoiceRequestStruct)
package fatcorews

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/security"
)

const (
	// TestURL is the AT staging endpoint for invoice communication.
	TestURL = "https://servicos.portaldasfinancas.gov.pt:723/fatcorews/ws/"
	// ProdURL is the AT production endpoint for invoice communication.
	ProdURL = "https://servicos.portaldasfinancas.gov.pt:423/fatcorews/ws/"

	fatcoreNS = "http://factemi.at.min_financas.pt/documents"
)

// Client is a SOAP client for the AT FatCore webservice.
// Create one with [NewClient]; it is safe for concurrent use.
type Client struct {
	httpClient *http.Client
	url        string
	username   string
	password   string
	atPubKey   *rsa.PublicKey
}

// NewClient creates a Client.
//
//   - endpoint: use [TestURL] or [ProdURL].
//   - username: Portal das Finanças sub-user in the form "NIF/subuser" (e.g. "555555555/37").
//   - password: plain-text Portal das Finanças password for the sub-user.
//   - clientCert: AT-signed TLS client certificate; load with [security.LoadClientCertPFX]
//     or [security.LoadClientCert].
//   - atPubKey: AT's authentication system RSA public key; load with [security.LoadATPublicKey].
func NewClient(endpoint, username, password string, clientCert tls.Certificate, atPubKey *rsa.PublicKey) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{clientCert},
				},
			},
		},
		url:      endpoint,
		username: username,
		password: password,
		atPubKey: atPubKey,
	}
}

// -----------------------------------------------------------------------
// Response type (shared by all nine operations)
// -----------------------------------------------------------------------

// Response is the result returned by every FatCore operation.
type Response struct {
	// CodigoResposta is the AT result code. Successful submissions use codes
	// in the 2xxx range (e.g. 2011 = created, 2012 = updated).
	CodigoResposta int       `xml:"CodigoResposta"`
	Mensagem       string    `xml:"Mensagem"`
	DataOperacao   time.Time `xml:"DataOperacao"`
}

// OperationResponse wraps a Response, matching the XML element name returned
// by each FatCore operation (e.g. <RegisterInvoiceResponse><Response>…).
type OperationResponse struct {
	Response Response `xml:"Response"`
}

// -----------------------------------------------------------------------
// Invoice operations
// -----------------------------------------------------------------------

// RegisterInvoice sends a new invoice (fatura) to AT.
// req must be an xml.Marshaler-compatible struct whose root element is
// <RegisterInvoiceRequest xmlns="http://factemi.at.min_financas.pt/documents">.
func (c *Client) RegisterInvoice(req interface{}) (*Response, error) {
	return c.call(req)
}

// ChangeInvoiceStatus updates the status of a previously registered invoice.
// req must produce a <ChangeInvoiceStatusRequest> element in the fatcore namespace.
func (c *Client) ChangeInvoiceStatus(req interface{}) (*Response, error) {
	return c.call(req)
}

// DeleteInvoice removes a previously communicated invoice from AT.
// req must produce a <DeleteInvoiceRequest> element in the fatcore namespace.
func (c *Client) DeleteInvoice(req interface{}) (*Response, error) {
	return c.call(req)
}

// -----------------------------------------------------------------------
// Work document operations
// -----------------------------------------------------------------------

// RegisterWork sends a work document (documento de conferência) to AT.
func (c *Client) RegisterWork(req interface{}) (*Response, error) {
	return c.call(req)
}

// ChangeWorkStatus updates the status of a previously registered work document.
func (c *Client) ChangeWorkStatus(req interface{}) (*Response, error) {
	return c.call(req)
}

// DeleteWork removes a previously communicated work document from AT.
func (c *Client) DeleteWork(req interface{}) (*Response, error) {
	return c.call(req)
}

// -----------------------------------------------------------------------
// Payment operations
// -----------------------------------------------------------------------

// RegisterPayment sends a payment receipt (recibo IVA de caixa) to AT.
func (c *Client) RegisterPayment(req interface{}) (*Response, error) {
	return c.call(req)
}

// ChangePaymentStatus updates the status of a previously registered payment.
func (c *Client) ChangePaymentStatus(req interface{}) (*Response, error) {
	return c.call(req)
}

// DeletePayment removes a previously communicated payment from AT.
func (c *Client) DeletePayment(req interface{}) (*Response, error) {
	return c.call(req)
}

// -----------------------------------------------------------------------
// SOAP transport
// -----------------------------------------------------------------------

// call builds a full SOAP 1.1 envelope with WS-Security header, posts it to
// the AT endpoint, and returns the parsed Response.
func (c *Client) call(requestBody interface{}) (*Response, error) {
	// Build security header (new random Ks per call — required by AT).
	secHeader, err := security.Build(c.username, c.password, c.atPubKey)
	if err != nil {
		return nil, fmt.Errorf("build security header: %w", err)
	}

	// Marshal request body.
	bodyBytes, err := xml.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	// Assemble SOAP envelope.
	env := soapEnvelope{
		Header: soapHeader{Security: secHeader},
		Body:   soapBody{Content: bodyBytes},
	}
	envBytes, err := xml.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshal soap envelope: %w", err)
	}

	// HTTP POST with mutual TLS.
	httpReq, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(envBytes))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
	httpReq.Header.Set("SOAPAction", "")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read http response: %w", err)
	}

	// Parse SOAP response envelope.
	var envelope soapResponseEnvelope
	if err := xml.Unmarshal(respBytes, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshal soap response: %w", err)
	}
	if envelope.Body.Fault != nil {
		return nil, errors.New(envelope.Body.Fault.FaultString)
	}

	var opResp OperationResponse
	if err := xml.Unmarshal(envelope.Body.Content, &opResp); err != nil {
		return nil, fmt.Errorf("unmarshal operation response: %w", err)
	}
	return &opResp.Response, nil
}

// -----------------------------------------------------------------------
// SOAP envelope structs
// -----------------------------------------------------------------------

type soapEnvelope struct {
	XMLName xml.Name   `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Header  soapHeader `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header"`
	Body    soapBody   `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type soapHeader struct {
	Security security.Header
}

type soapBody struct {
	Content []byte `xml:",innerxml"`
}

type soapResponseEnvelope struct {
	XMLName xml.Name         `xml:"Envelope"`
	Body    soapResponseBody `xml:"Body"`
}

type soapResponseBody struct {
	Fault   *soapFault `xml:"Fault"`
	Content []byte     `xml:",innerxml"`
}

type soapFault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}
