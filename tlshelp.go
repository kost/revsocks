package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"time"
)
import "crypto/tls"

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandString generates random string of n size
// It returns the generated random string.and any write error encountered.
func RandString(n int) string {
	r := make([]byte, n)
	_, err := rand.Read(r)
	if err != nil {
		return ""
	}

	b := make([]byte, n)
	l := len(letters)
	for i := range b {
		b[i] = letters[int(r[i])%l]
	}
	return string(b)
}

// RandBytes generates random bytes of n size
// It returns the generated random bytes
func RandBytes(n int) []byte {
	r := make([]byte, n)
	_, err := rand.Read(r)
	if err != nil {
	}

	return r
}

// RandBigInt generates random big integer with max number
// It returns the generated random big integer
func RandBigInt(max *big.Int) *big.Int {
	r, _ := rand.Int(rand.Reader, max)
	return r
}

func genPair(keysize int) (cacert []byte, cakey []byte, cert []byte, certkey []byte) {

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	ca := &x509.Certificate{
		SerialNumber: RandBigInt(serialNumberLimit),
		Subject: pkix.Name{
			Country:            []string{RandString(16)},
			Organization:       []string{RandString(16)},
			OrganizationalUnit: []string{RandString(16)},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          RandBytes(5),
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, _ := rsa.GenerateKey(rand.Reader, keysize)
	pub := &priv.PublicKey
	caBin, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		log.Println("create ca failed", err)
		return
	}

	cert2 := &x509.Certificate{
		SerialNumber: RandBigInt(serialNumberLimit),
		Subject: pkix.Name{
			Country:            []string{RandString(16)},
			Organization:       []string{RandString(16)},
			OrganizationalUnit: []string{RandString(16)},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: RandBytes(6),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	priv2, _ := rsa.GenerateKey(rand.Reader, keysize)
	pub2 := &priv2.PublicKey
	cert2Bin, err2 := x509.CreateCertificate(rand.Reader, cert2, ca, pub2, priv)
	if err2 != nil {
		log.Println("create cert2 failed", err2)
		return
	}

	privBin := x509.MarshalPKCS1PrivateKey(priv)
	priv2Bin := x509.MarshalPKCS1PrivateKey(priv2)

	return caBin, privBin, cert2Bin, priv2Bin

}

func verifyCert(cacert []byte, cert []byte) bool {
	caBin, _ := x509.ParseCertificate(cacert)
	cert2Bin, _ := x509.ParseCertificate(cert)
	err3 := cert2Bin.CheckSignatureFrom(caBin)
	if err3 != nil {
		return false
	}
	return true
}

func getPEMs(cert []byte, key []byte) (pemcert []byte, pemkey []byte) {
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: key,
	})

	return certPem, keyPem
}

func getTLSPair(certPem []byte, keyPem []byte) (tls.Certificate, error) {
	tlspair, errt := tls.X509KeyPair(certPem, keyPem)
	if errt != nil {
		return tlspair, errt
	}
	return tlspair, nil
}

func getRandomTLS(keysize int) (tls.Certificate, error) {
	_, _, cert, certkey := genPair(keysize)
	certPem, keyPem := getPEMs(cert, certkey)
	tlspair, err := getTLSPair(certPem, keyPem)
	return tlspair, err
}
