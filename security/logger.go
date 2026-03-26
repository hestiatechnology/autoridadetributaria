package security

import (
	"log"
	"net/http"
	"net/http/httputil"
)

// HTTPDoer is an interface matching the Do method of *http.Client.
// It is used to allow wrapping of any custom http client.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// LoggingHTTPClient intercepts and logs HTTP requests and responses.
// It implements the HTTPClient interface required by the gowsdl soap package.
//
// Usage:
//
//	httpClient := &http.Client{
//		Transport: &http.Transport{
//			TLSClientConfig: &tls.Config{Certificates: []tls.Certificate{clientCert}},
//		},
//	}
//
//	loggingClient := &security.LoggingHTTPClient{
//		Client: httpClient,
//		// Optional custom logger, defaults to log.Printf
//		Logger: log.Printf,
//	}
//
//	client := soap.NewClient(
//		"https://servicos.portaldasfinancas.gov.pt:701/SeriesWSService/SeriesWS",
//		soap.WithHTTPClient(loggingClient),
//	)
type LoggingHTTPClient struct {
	// Client is the underlying HTTP client (defaults to http.DefaultClient if nil).
	Client HTTPDoer

	// Logger is the function used to print the dump (defaults to log.Printf if nil).
	Logger func(format string, args ...interface{})
}

// Do executes the HTTP request and logs both the raw request and response over the wire.
func (l *LoggingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	logFn := l.Logger
	if logFn == nil {
		logFn = log.Printf
	}

	// Dump the request (including the body)
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logFn("Error dumping HTTP request: %v\n", err)
	} else {
		logFn("=== SOAP HTTP REQUEST ===\n%s\n=========================\n", string(reqDump))
	}

	// Determine which client to use
	client := l.Client
	if client == nil {
		client = http.DefaultClient
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		logFn("HTTP request execution failed: %v\n", err)
		return resp, err
	}

	// Dump the response (including the body)
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		logFn("Error dumping HTTP response: %v\n", err)
	} else {
		logFn("=== SOAP HTTP RESPONSE ===\n%s\n==========================\n", string(respDump))
	}

	return resp, nil
}
