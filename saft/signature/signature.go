package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrPrivateKeyNotSet = errors.New("private key not set")
	ErrInvoiceNoNotSet  = errors.New("invoice number not set")
)

func SignFiscalDocument(r *rsa.PrivateKey, date time.Time, systemDate time.Time, invoiceNo string, grossTotal decimal.Decimal, lastHash string) ([]byte, error) {
	if r == nil {
		return nil, ErrPrivateKeyNotSet
	}

	if date.IsZero() {
		date = time.Now()
	}

	if systemDate.IsZero() {
		systemDate = time.Now()
	}

	if invoiceNo == "" {
		return nil, ErrInvoiceNoNotSet
	}

	// Sign in this format 2010-05-18;2010-05-18T11:22:19;FAC 001/14;3.12;lastHash
	message := date.Format("2006-01-02") + ";" + systemDate.Format("2006-01-02T15:04:05") + ";" + invoiceNo + ";" + grossTotal.StringFixed(2) + ";" + lastHash
	hashed := crypto.SHA1.New().Sum([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, r, crypto.SHA1, hashed)
	if err != nil {
		return nil, err
	}

	// Encode the signature in base64
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	return []byte(signatureBase64), nil
}
