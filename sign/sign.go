package sign

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

func CheckPrivateKeyDetails(pemData []byte) (bool, error) {
	// 1. Decodificar o bloco PEM. Se isso funcionar, a chave está em Base64
	// dentro de um container de texto (geralmente ASCII/UTF-8).
	block, _ := pem.Decode(pemData)
	if block == nil {
		return false, errors.New("failure decoding PEM block")
	}

	// 2. Tentar parsear a chave. O pacote x509 lida com o formato DER (ASN.1)
	// que está codificado em Base64 no PEM.
	// Primeiro, tentamos o formato moderno PKCS#8.
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Se falhar, tentamos o formato legado PKCS#1, específico para RSA.
		parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return false, fmt.Errorf("failure parsing private key (neither PKCS#8 nor PKCS#1): %v", err)
		}
	}

	// 3. Verificar o tipo e o tamanho da chave.
	// O requisito de "1024 bits" geralmente se aplica a chaves RSA.
	switch key := parsedKey.(type) {
	case *rsa.PrivateKey:
		// Para uma chave RSA, o tamanho é o comprimento em bits do módulo (N).
		keySize := key.N.BitLen()
		//fmt.Printf("Chave RSA detectada com tamanho de %d bits.\n", keySize)
		if keySize == 1024 {
			return true, nil // A chave é RSA e tem 1024 bits.
		}
		// Se o tamanho for diferente, não atende ao requisito.
		return false, nil
	default:
		// Se a chave for de outro tipo (ex: ECDSA, Ed25519), não atende ao requisito de tamanho.
		return false, fmt.Errorf("a chave não é do tipo RSA, mas sim %T", parsedKey)
	}
}

func LoadPrivateKey(privateKeyStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return nil, errors.New("failure decoding PEM block")
	}

	// 2. Tentar parsear a chave. O pacote x509 lida com o formato DER (ASN.1)
	// que está codificado em Base64 no PEM.
	// Primeiro, tentamos o formato moderno PKCS#8.
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Se falhar, tentamos o formato legado PKCS#1, específico para RSA.
		parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failure parsing private key (neither PKCS#8 nor PKCS#1): %v", err)
		}
	}

	return parsedKey.(*rsa.PrivateKey), nil
}

type Document struct {
	Date            time.Time
	SystemEntryDate time.Time
	DocumentNo      string
	GrossTotal      decimal.Decimal
	Hash            string
}

func SignDocument(key *rsa.PrivateKey, document Document) ([]byte, error) {
	strToEncode := fmt.Sprintf("%s;%s;%s;%s;%s", document.Date.Format("2006-01-02"), document.SystemEntryDate.Format("2006-01-02T15:04:05"), document.DocumentNo, document.GrossTotal.StringFixed(2), document.Hash)

	fmt.Printf("String to sign: %s\n", strToEncode)

	// Passo 1: Calcular o hash SHA-1 da string
	hasher := sha1.New()
	hasher.Write([]byte(strToEncode))
	hashed := hasher.Sum(nil)

	// Passo 2: Assinar o hash com a chave privada (RSA-PKCS1v15)
	//rsa.Sign
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA1, hashed)
	if err != nil {
		return nil, fmt.Errorf("erro ao assinar o documento: %w", err)
	}

	// Passo 3: Codificar a assinatura em Base64
	encodedSignature := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
	base64.StdEncoding.Encode(encodedSignature, signature)

	return encodedSignature, nil
}
