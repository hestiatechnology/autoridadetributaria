package seriesws_test

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/security"
	"github.com/hestiatechnology/autoridadetributaria/seriesws"
	"github.com/hooklift/gowsdl/soap"
)

func Example() {
	// 1. Load the AT Public Key (used to encrypt the WS-Security password and nonce)
	// You must request this from AT (e.g., SA.cer)
	atPubKey, err := security.LoadATPublicKey("path/to/SA.cer")
	if err != nil {
		log.Fatalf("Failed to load AT public key: %v", err)
	}

	// 2. Load your software producer client certificate (mTLS)
	// This was generated during the CSR process and signed by AT
	clientCert, err := security.LoadClientCertPFX("path/to/555555555.pfx", "your_pfx_password")
	if err != nil {
		log.Fatalf("Failed to load client certificate: %v", err)
	}

	// 3. Configure TLS for the SOAP client to use the mTLS certificate
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
	}

	// Create an underlying HTTP client with the TLS config
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Wrap the HTTP client with the LoggingHTTPClient to see the raw SOAP XML
	loggingClient := &security.LoggingHTTPClient{
		Client: httpClient,
		Logger: log.Printf,
	}

	// 4. Initialize the SOAP client pointing to the AT webservice endpoint
	// Test env: https://servicos.portaldasfinancas.gov.pt:701/SeriesWSService/SeriesWS
	// Prod env: https://servicos.portaldasfinancas.gov.pt:401/SeriesWSService/SeriesWS
	client := soap.NewClient(
		"https://servicos.portaldasfinancas.gov.pt:701/SeriesWSService/SeriesWS",
		soap.WithHTTPClient(loggingClient),
	)

	// 5. Build and add the WS-Security header
	// The username is typically the NIF and subuser ID (e.g., "555555555/37")
	// The password is the subuser password configured in Portal das Finanças
	secHeader, err := security.Build("555555555/37", "subuser_password", atPubKey)
	if err != nil {
		log.Fatalf("Failed to build security header: %v", err)
	}
	client.AddHeader(secHeader)

	// 6. Create the SeriesWS service client using the configured SOAP client
	service := seriesws.NewSeriesWS(client)

	// Prepare variable pointers needed for the request.
	// `gowsdl` uses pointers so you can selectively omit fields.
	serie := seriesws.SerieType("2024")
	tipoSerie := seriesws.TipoSerieType("N")  // N - Normal
	classeDoc := seriesws.ClasseDocType("FT") // FT - Fatura
	tipoDoc := seriesws.TipoDocType("FT")     // FT - Fatura
	numInicialSeq := seriesws.NumSeqType(1)
	numCertSWFatur := seriesws.NumCertSWFaturType(0)          // 0 se não aplicável (e.g., documentos emitidos no portal)
	meioProcessamento := seriesws.MeioProcessamentoType("PI") // PI - Processado por programa Informático

	// Build the request payload
	req := &seriesws.RegistarSerie{
		Serie:                &serie,
		TipoSerie:            &tipoSerie,
		ClasseDoc:            &classeDoc,
		TipoDoc:              &tipoDoc,
		NumInicialSeq:        &numInicialSeq,
		DataInicioPrevUtiliz: soap.CreateXsdDate(time.Now(), false), // Required type for XSD date
		NumCertSWFatur:       &numCertSWFatur,
		MeioProcessamento:    &meioProcessamento,
	}

	// Execute the SOAP call
	resp, err := service.RegistarSerie(req)
	if err != nil {
		log.Fatalf("Failed to register series: %v", err)
	}

	// Process and output the response
	if resp != nil && resp.RegistarSerieResp != nil && resp.RegistarSerieResp.InfoResultOper != nil {
		fmt.Printf("Result Code: %d\n", *resp.RegistarSerieResp.InfoResultOper.CodResultOper)
		fmt.Printf("Result Message: %s\n", *resp.RegistarSerieResp.InfoResultOper.MsgResultOper)

		// Code 0 usually means success
		if *resp.RegistarSerieResp.InfoResultOper.CodResultOper == 0 && resp.RegistarSerieResp.InfoSerie != nil {
			fmt.Printf("Validation Code (CodValidacaoSerie): %s\n", *resp.RegistarSerieResp.InfoSerie.CodValidacaoSerie)
		}
	}
}
