// Package seriesws implements a SOAP client for the AT Series webservice
// (Comunicação de Séries — SeriesWSService).
//
// Usage:
//
//	// 1. Load certificates
//	clientCert, _ := security.LoadClientCertPFX("555555555.pfx", "password")
//	atPubKey,   _ := security.LoadATPublicKey("SA.cer")
//
//	// 2. Create client
//	c := seriesws.NewClient(seriesws.TestURL, "555555555/37", "password", clientCert, atPubKey)
//
//	// 3. Call an operation
//	resp, err := c.RegistarSerie(types.RegistarSerieRequest{
//	    Serie:                "A",
//	    TipoSerie:            types.TipoSerieNormal,
//	    ClasseDoc:            types.ClasseDocSI,
//	    TipoDoc:              types.TipoDocFT,
//	    NumInicialSeq:        1,
//	    DataInicioPrevUtiliz: types.NewXSDDate(time.Now()),
//	    NumCertSWFatur:       0,
//	    MeioProcessamento:    types.MeioProcessamentoPI,
//	})
package seriesws

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hestiatechnology/autoridadetributaria/security"
	"github.com/hestiatechnology/autoridadetributaria/seriesws/types"
)

const (
	// TestURL is the AT staging endpoint for the Series webservice.
	TestURL = "https://servicos.portaldasfinancas.gov.pt:722/SeriesWSService"
	// ProdURL is the AT production endpoint for the Series webservice.
	ProdURL = "https://servicos.portaldasfinancas.gov.pt:422/SeriesWSService"
)

// Client is a SOAP client for the AT Series webservice.
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
// Operations
// -----------------------------------------------------------------------

// RegistarSerie communicates a new document series to AT and returns the
// AT-assigned series validation code (codValidacaoSerie) in the response.
func (c *Client) RegistarSerie(req types.RegistarSerieRequest) (*types.RegistarSerieResponse, error) {
	var resp types.RegistarSerieResponse
	if err := c.call(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FinalizarSerie marks a series as finalized — no more documents will be
// issued under it beyond the last emitted sequence number.
func (c *Client) FinalizarSerie(req types.FinalizarSerieRequest) (*types.FinalizarSerieResponse, error) {
	var resp types.FinalizarSerieResponse
	if err := c.call(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ConsultarSeries queries registered series. All filter fields in the request
// are optional; a zero-value request returns all series.
func (c *Client) ConsultarSeries(req types.ConsultarSeriesRequest) (*types.ConsultarSeriesResponse, error) {
	var resp types.ConsultarSeriesResponse
	if err := c.call(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AnularSerie cancels a previously registered series (correction of error).
// req.DeclaracaoNaoEmissao must be true, confirming that no documents have
// been issued with the series being cancelled.
func (c *Client) AnularSerie(req types.AnularSerieRequest) (*types.AnularSerieResponse, error) {
	var resp types.AnularSerieResponse
	if err := c.call(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// -----------------------------------------------------------------------
// SOAP transport
// -----------------------------------------------------------------------

// call builds a full SOAP 1.1 envelope with WS-Security header, posts it to
// the AT endpoint, and unmarshals the response body into resp.
func (c *Client) call(requestBody, resp interface{}) error {
	// Build security header (new random Ks per call).
	secHeader, err := security.Build(c.username, c.password, c.atPubKey)
	if err != nil {
		return fmt.Errorf("build security header: %w", err)
	}

	// Marshal request body.
	bodyBytes, err := xml.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	// Assemble SOAP envelope.
	env := soapEnvelope{
		Header: soapHeader{Security: secHeader},
		Body:   soapBody{Content: bodyBytes},
	}
	envBytes, err := xml.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal soap envelope: %w", err)
	}

	// HTTP POST.
	httpReq, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(envBytes))
	if err != nil {
		return fmt.Errorf("create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
	httpReq.Header.Set("SOAPAction", "")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Parse SOAP response envelope.
	var envelope soapResponseEnvelope
	if err := xml.Unmarshal(respBytes, &envelope); err != nil {
		return fmt.Errorf("unmarshal soap response: %w", err)
	}
	if envelope.Body.Fault != nil {
		return errors.New(envelope.Body.Fault.FaultString)
	}

	if err := xml.Unmarshal(envelope.Body.Content, resp); err != nil {
		return fmt.Errorf("unmarshal response body: %w", err)
	}
	return nil
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
