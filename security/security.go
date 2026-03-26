// Package security implements the WS-Security UsernameToken Profile required
// by all AT (Autoridade Tributária) webservices.
//
// The AT authentication scheme works as follows for every SOAP request:
//
//  1. Generate a random 128-bit AES key (Ks) — used only for this request.
//  2. Password  = Base64( AES_ECB_PKCS5(Ks, plainPassword) )
//  3. Nonce     = Base64( RSA_PKCS1v15(KpubAT, Ks) )
//  4. Created   = Base64( AES_ECB_PKCS5(Ks, timestampUTC_ISO8601) )
//
// The AT public key (KpubAT) is obtained from AT by email request (see the
// integration manual, section 2.1.1).
package security

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/pkcs12"
)

// Header is the WS-Security SOAP header element.
// Embed it in a soap:Header struct for SOAP requests.
type Header struct {
	XMLName       xml.Name      `xml:"wss:Security"`
	XmlNSWss      string        `xml:"xmlns:wss,attr"`
	UsernameToken UsernameToken `xml:"wss:UsernameToken"`
}

// UsernameToken holds the encrypted credentials for a single SOAP call.
type UsernameToken struct {
	Username string `xml:"wss:Username"`
	// AES-ECB-PKCS5 encrypted password, Base64-encoded.
	Password string `xml:"wss:Password"`
	// RSA-PKCS1v15 encrypted symmetric key (Ks), Base64-encoded.
	Nonce string `xml:"wss:Nonce"`
	// AES-ECB-PKCS5 encrypted UTC timestamp, Base64-encoded.
	Created string `xml:"wss:Created"`
}

// Build constructs a WS-Security header for a single AT webservice request.
//
//   - username: NIF/subuser identifier as configured in Portal das Finanças,
//     e.g. "555555555/37" (NIF of the taxpayer / sub-user number).
//   - password: plain-text Portal das Finanças password for the sub-user.
//   - atPubKey: AT's authentication system RSA public key. Load it with
//     [LoadATPublicKey] from the .cer file provided by AT.
//
// A new random symmetric key is generated on every call, so each invocation
// produces a different nonce — as required by the AT to prevent replay attacks.
func Build(username, password string, atPubKey *rsa.PublicKey) (Header, error) {
	// Step 1 — random 128-bit AES symmetric key for this request only.
	ks := make([]byte, 16)
	if _, err := rand.Read(ks); err != nil {
		return Header{}, fmt.Errorf("generate symmetric key: %w", err)
	}

	// Step 2 — encrypt password.
	encPassword, err := aesECBEncrypt(ks, []byte(password))
	if err != nil {
		return Header{}, fmt.Errorf("encrypt password: %w", err)
	}

	// Step 3 — encrypt Ks with AT's RSA public key.
	encKs, err := rsa.EncryptPKCS1v15(rand.Reader, atPubKey, ks)
	if err != nil {
		return Header{}, fmt.Errorf("encrypt symmetric key: %w", err)
	}

	// Step 4 — encrypt current UTC timestamp (ISO 8601).
	// The AT server uses this to enforce request freshness.
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	encTimestamp, err := aesECBEncrypt(ks, []byte(timestamp))
	if err != nil {
		return Header{}, fmt.Errorf("encrypt timestamp: %w", err)
	}

	return Header{
		XmlNSWss: "http://schemas.xmlsoap.org/ws/2002/12/secext",
		UsernameToken: UsernameToken{
			Username: username,
			Password: base64.StdEncoding.EncodeToString(encPassword),
			Nonce:    base64.StdEncoding.EncodeToString(encKs),
			Created:  base64.StdEncoding.EncodeToString(encTimestamp),
		},
	}, nil
}

// -----------------------------------------------------------------------
// AT public key loading
// -----------------------------------------------------------------------

// LoadATPublicKey loads AT's authentication system RSA public key from a file.
// The file may be:
//   - A PEM-encoded X.509 certificate (.cer, .crt, .pem)
//   - A DER-encoded X.509 certificate (.cer)
//   - A PEM or DER PKIX public key
//
// AT provides this key on request (see integration manual, section 2.1.1).
// The test and production keys are different; request both from AT.
func LoadATPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read AT public key file: %w", err)
	}
	return ParseATPublicKey(data)
}

// ATPublicKeyFromString parses AT's authentication system RSA public key from
// a PEM string (the contents of the .cer file provided by AT).
func ATPublicKeyFromString(pemOrDer string) (*rsa.PublicKey, error) {
	return ParseATPublicKey([]byte(pemOrDer))
}

// ParseATPublicKey parses an RSA public key from raw PEM or DER bytes.
func ParseATPublicKey(data []byte) (*rsa.PublicKey, error) {
	// Strip PEM wrapper if present.
	der := data
	if block, _ := pem.Decode(data); block != nil {
		der = block.Bytes
	}

	// Try X.509 certificate (most common — AT ships a .cer file).
	if cert, err := x509.ParseCertificate(der); err == nil {
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("AT certificate does not contain an RSA public key")
		}
		return pub, nil
	}

	// Fallback: bare PKIX public key.
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("cannot parse AT public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("AT public key is not RSA")
	}
	return rsaPub, nil
}

// -----------------------------------------------------------------------
// Client certificate loading
// -----------------------------------------------------------------------

// LoadClientCert loads the software producer's AT client certificate from
// separate PEM-encoded certificate and private key files (e.g. the .crt and
// .key files produced during the CSR process — see manual section 2.3).
func LoadClientCert(certPEMFile, keyPEMFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certPEMFile, keyPEMFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("load client cert: %w", err)
	}
	return cert, nil
}

// ClientCertFromStrings parses a TLS client certificate from PEM strings.
// certPEM is the certificate (contents of the .crt file) and keyPEM is the
// private key (contents of the .key file).
func ClientCertFromStrings(certPEM, keyPEM string) (tls.Certificate, error) {
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("parse client cert: %w", err)
	}
	return cert, nil
}

// LoadClientCertPFX loads the software producer's AT client certificate from
// a PKCS12 (.pfx / .p12) file. This is the format produced by:
//
//	openssl pkcs12 -export -in 555555555.crt -inkey 555555555.key -out 555555555.pfx
//
// password is the PFX export password chosen when running openssl; use an
// empty string if the file was created without a password.
func LoadClientCertPFX(pfxFile, password string) (tls.Certificate, error) {
	data, err := os.ReadFile(pfxFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("read pfx file: %w", err)
	}
	return clientCertFromPFXBytes(data, password)
}

// ClientCertFromPFXBase64 parses a TLS client certificate from a base64-encoded
// PFX string and its export password. This is useful when the certificate is
// stored in an environment variable or a configuration file.
//
// To get the base64 string from a .pfx file:
//
//	base64 -w 0 555555555.pfx
func ClientCertFromPFXBase64(pfxBase64, password string) (tls.Certificate, error) {
	data, err := base64.StdEncoding.DecodeString(pfxBase64)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("decode base64 pfx: %w", err)
	}
	return clientCertFromPFXBytes(data, password)
}

func clientCertFromPFXBytes(data []byte, password string) (tls.Certificate, error) {
	pemBlocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("decode pfx: %w", err)
	}

	var certPEM, keyPEM []byte
	for _, b := range pemBlocks {
		switch b.Type {
		case "CERTIFICATE":
			certPEM = append(certPEM, pem.EncodeToMemory(b)...)
		case "PRIVATE KEY":
			keyPEM = append(keyPEM, pem.EncodeToMemory(b)...)
		}
	}
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return tls.Certificate{}, fmt.Errorf("pfx must contain both a certificate and a private key")
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("build tls certificate: %w", err)
	}
	return cert, nil
}

// -----------------------------------------------------------------------
// Internal crypto helpers
// -----------------------------------------------------------------------

// aesECBEncrypt encrypts plaintext with AES in ECB mode (PKCS5/PKCS7 padding).
// Go's crypto/cipher does not expose ECB directly; we implement it by encrypting
// each block independently, which is the definition of ECB mode.
func aesECBEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := pkcs5Pad(plaintext, block.BlockSize())
	out := make([]byte, len(padded))
	for i := 0; i < len(padded); i += block.BlockSize() {
		block.Encrypt(out[i:i+block.BlockSize()], padded[i:i+block.BlockSize()])
	}
	return out, nil
}

// pkcs5Pad applies PKCS5/PKCS7 padding so len(data) is a multiple of blockSize.
func pkcs5Pad(data []byte, blockSize int) []byte {
	n := blockSize - (len(data) % blockSize)
	return append(data, bytes.Repeat([]byte{byte(n)}, n)...)
}
