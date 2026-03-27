package workDocuments

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
)

// Client is a SOAP client for the AT Documentos de Transporte webservice.
type Client struct {
	httpClient *http.Client
	url        string
	username   string
	password   string
	atPubKey   *rsa.PublicKey
}

// NewClient creates a new WS client configured for MTLS and WS-Security.
func NewClient(
	url string,
	username, password string,
	clientCert tls.Certificate,
	atPubKey *rsa.PublicKey,
) *Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{clientCert},
			},
		},
	}

	return &Client{
		httpClient: httpClient,
		url:        url,
		username:   username,
		password:   password,
		atPubKey:   atPubKey,
	}
}

type soapHeader struct {
	XMLName  xml.Name    `xml:"soapenv:Header"`
	Security interface{} `xml:"wsse:Security"`
}

type soapBody struct {
	XMLName xml.Name    `xml:"soapenv:Body"`
	Content interface{} `xml:",any"`
}

type soapEnvelope struct {
	XMLName xml.Name   `xml:"soapenv:Envelope"`
	Soapenv string     `xml:"xmlns:soapenv,attr"`
	Wsse    string     `xml:"xmlns:wsse,attr"`
	Header  soapHeader `xml:"soapenv:Header"`
	Body    soapBody   `xml:"soapenv:Body"`
}

type soapResponseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Fault   *soapFault `xml:"Fault"`
		Content []byte     `xml:",innerxml"`
	} `xml:"Body"`
}

type soapFault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

// call sends the request using the same WS-Security logic as the invoice client
func (c *Client) call(requestBody interface{}) ([]byte, error) {
	secHeader, err := security.Build(c.username, c.password, c.atPubKey)
	if err != nil {
		return nil, fmt.Errorf("build security header: %w", err)
	}

	env := soapEnvelope{
		Soapenv: "http://schemas.xmlsoap.org/soap/envelope/",
		Wsse:    "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd",
		Header:  soapHeader{Security: secHeader},
		Body:    soapBody{Content: requestBody},
	}

	envBytes, err := xml.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshal soap envelope: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(envBytes))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
	httpReq.Header.Set("SOAPAction", "https://servicos.portaldasfinancas.gov.pt/sgdtws/documentosTransporte/")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read http response: %w", err)
	}

	var envelope soapResponseEnvelope
	if err := xml.Unmarshal(respBytes, &envelope); err != nil {
		fmt.Printf("\n=== AT TRANSPORT SOAP FAULT ===\nRequest XML:\n%s\n\nResponse XML:\n%s\n===============================\n\n", string(envBytes), string(respBytes))
		return nil, fmt.Errorf("unmarshal soap response: %w", err)
	}
	if envelope.Body.Fault != nil {
		fmt.Printf("\n=== AT TRANSPORT SOAP FAULT ===\nRequest XML:\n%s\n\nResponse XML:\n%s\n===============================\n\n", string(envBytes), string(respBytes))
		return nil, errors.New(envelope.Body.Fault.FaultString)
	}

	return envelope.Body.Content, nil
}

func (c *Client) EnvioDocumentoTransporte(username, password string, request *StockMovement) (*StockMovementResponse, error) {
	// Update credentials for this call
	c.username = username
	c.password = password

	// Wrap with a prefixed namespace (ns1:) so child elements are NOT in any namespace.
	// If we let StockMovement's own XMLName set xmlns="" (default namespace), AT's XSD
	// parser rejects every child element because the schema has elementFormDefault=unqualified
	// (children must be namespace-free). Using a prefix keeps the default namespace empty.
	type wrapper struct {
		XMLName xml.Name `xml:"ns1:envioDocumentoTransporteRequestElem"`
		Ns1     string   `xml:"xmlns:ns1,attr"`
		*StockMovement
	}
	req := wrapper{
		Ns1:           "https://servicos.portaldasfinancas.gov.pt/sgdtws/documentosTransporte/",
		StockMovement: request,
	}

	respBytes, err := c.call(req)
	if err != nil {
		return nil, err
	}

	var resp StockMovementResponse
	if err := xml.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}
