//  Copyright (c) 2014, Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tao

import (
	"bytes"
	"crypto"
	"crypto/aes"
	// "crypto/sha256"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	// "io/ioutil"
	// "os"
)

func printKey(cryptoKey *CryptoKey) {
	if cryptoKey.KeyHeader.Version == nil || *cryptoKey.KeyHeader.Version != CryptoVersion_CRYPTO_VERSION_2 {
		fmt.Printf("Wrong version\n")
	}
	if cryptoKey.KeyHeader.KeyName == nil {
		fmt.Printf("No key name\n")
	} else {
		fmt.Printf("Key name: %s\n", *cryptoKey.KeyHeader.KeyName)
	}
	if cryptoKey.KeyHeader.KeyType == nil {
		fmt.Printf("No key type\n")
	} else {
		fmt.Printf("Key type: %s\n", *cryptoKey.KeyHeader.KeyType)
	}
	if cryptoKey.KeyHeader.KeyPurpose == nil {
		fmt.Printf("No Purpose\n")
	} else {
		fmt.Printf("Purpose: %s\n", *cryptoKey.KeyHeader.KeyPurpose)
	}
	if cryptoKey.KeyHeader.KeyStatus == nil {
		fmt.Printf("No key status\n")
	} else {
		fmt.Printf("Key status: %s\n", *cryptoKey.KeyHeader.KeyStatus)
	}
	n := len(cryptoKey.KeyComponents)
	for i := 0; i < n; i++ {
		fmt.Printf("Component %d: %x\n", i, cryptoKey.KeyComponents[i])
	}
}

func TestGenerateKeys(t *testing.T) {
	var keyName string
	var keyEpoch int32
	var keyPurpose string
	var keyStatus string

	// "aes-128-raw"
	keyName = "keyName1"
	keyEpoch = 1
	keyPurpose = "crypting"
	keyStatus = "active"
	cryptoKey1 := GenerateCryptoKey("aes-128-raw", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey1 == nil {
		t.Fatal("Can't generate aes-128-raw key\n")
	}
	fmt.Printf("Testing aes-128-raw generation\n")
	printKey(cryptoKey1)
	fmt.Printf("\n")
	m1 := MarshalCryptoKey(*cryptoKey1)
	if m1 == nil {
		t.Fatal("Can't MarshalCryptoKey aes-128-raw key\n")
	}
	cryptoKey1_d, err := UnmarshalCryptoKey(m1)
	if err != nil {
		t.Fatal("Can't UnmarshalCryptoKey aes-128-raw key\n")
	}
	printKey(cryptoKey1_d)
	fmt.Printf("\n")
	crypter, err := aes.NewCipher(cryptoKey1_d.KeyComponents[0])
	if err != nil {
		t.Fatal("Can't create aes-128 encrypter\n")
	}
	plain := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	fmt.Printf("aes key size %d, key: %x, plain size %d, BlockSize: %d\n", len(cryptoKey1_d.KeyComponents[0]),
		cryptoKey1_d.KeyComponents[0], len(plain), crypter.BlockSize())
	encrypted := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	decrypted := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	crypter.Encrypt(encrypted, plain)
	crypter.Decrypt(decrypted, encrypted)
	if !bytes.Equal(plain, decrypted) {
		t.Fatal("aes-128-raw plain and decrypted dont match\n")
	} else {
		fmt.Printf("Encryption works, encrypted: %x\n", encrypted)
	}
	fmt.Printf("\n")

	// "aes-256-raw"
	keyName = "keyName2"
	keyEpoch = 2
	keyPurpose = "crypting"
	keyStatus = "active"
	cryptoKey2 := GenerateCryptoKey("aes-256-raw", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey2 == nil {
		t.Fatal("Can't generate aes-256-raw key\n")
	}
	fmt.Printf("Testing aes-256-raw generation\n")
	printKey(cryptoKey2)
	fmt.Printf("\n")
	m2 := MarshalCryptoKey(*cryptoKey2)
	if m2 == nil {
		t.Fatal("Can't MarshalCryptoKey aes-256-raw key\n")
	}
	cryptoKey2_d, err := UnmarshalCryptoKey(m2)
	if err != nil {
		t.Fatal("Can't UnmarshalCryptoKey aes-256-raw key\n")
	}
	printKey(cryptoKey2_d)
	fmt.Printf("\n")
	crypter2, err := aes.NewCipher(cryptoKey2_d.KeyComponents[0])
	if err != nil {
		t.Fatal("Can't create aes-256 encrypter\n")
	}
	plain2 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	fmt.Printf("aes key size %d, key: %x, plain size %d, BlockSize: %d\n", len(cryptoKey2_d.KeyComponents[0]),
		cryptoKey2_d.KeyComponents[0], len(plain2), crypter2.BlockSize())
	encrypted2 := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	decrypted2 := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	crypter2.Encrypt(encrypted2, plain2)
	crypter2.Decrypt(decrypted2, encrypted2)
	if !bytes.Equal(plain2, decrypted2) {
		t.Fatal("aes-256-raw plain and decrypted dont match\n")
	} else {
		fmt.Printf("Encryption works, encrypted: %x\n", encrypted2)
	}
	fmt.Printf("\n")

	// "aes-128-ctr"
	keyName = "keyName3"
	keyEpoch = 3
	keyPurpose = "crypting"
	keyStatus = "active"
	cryptoKey3 := GenerateCryptoKey("aes-128-ctr", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey1 == nil {
		t.Fatal("Can't generate aes-128-ctr key\n")
	}
	fmt.Printf("Testing aes-128-ctr generation\n")
	printKey(cryptoKey3)
	fmt.Printf("\n")

	// "aes-256-ctr"
	keyName = "keyName4"
	keyEpoch = 4
	keyPurpose = "crypting"
	keyStatus = "active"
	cryptoKey4 := GenerateCryptoKey("aes-256-ctr", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey1 == nil {
		t.Fatal("Can't generate aes-256-ctrkey\n")
	}
	fmt.Printf("Testing aes-256-ctr generation\n")
	printKey(cryptoKey4)
	fmt.Printf("\n")

	// "aes-128-sha-256-cbc"
	keyName = "keyName5"
	keyEpoch = 2
	keyPurpose = "crypting"
	keyStatus = "active"
	cryptoKey5 := GenerateCryptoKey("aes-128-sha-256-cbc", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey5 == nil {
		t.Fatal("Can't generate aes-128-sha-256-cbc key\n")
	}
	fmt.Printf("Testing aes-128-sha-256-cbc generation\n")
	printKey(cryptoKey5)
	fmt.Printf("\n")

	// "aes-256-sha-384-cbc"
	keyName = "keyName6"
	keyEpoch = 2
	keyPurpose = "crypting"
	keyStatus = "active"
	cryptoKey6 := GenerateCryptoKey("aes-256-sha-384-cbc", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey6 == nil {
		t.Fatal("Can't generate key\n")
	}
	fmt.Printf("Testing aes-256-sha-384-cbc generation\n")
	printKey(cryptoKey6)
	fmt.Printf("\n")

	// "sha-256-hmac"
	keyName = "keyName7"
	keyEpoch = 2
	keyPurpose = "crypting"
	keyStatus = "active"
	cryptoKey7 := GenerateCryptoKey("sha-256-hmac", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey7 == nil {
		t.Fatal("Can't sha-256-hmac key\n")
	}
	fmt.Printf("Testing sha-256-hmac generation\n")
	printKey(cryptoKey7)
	fmt.Printf("\n")

	// "sha-384-hmac"
	keyName = "keyName8"
	keyEpoch = 2
	keyPurpose = "crypting"
	keyStatus = "active"
	cryptoKey8 := GenerateCryptoKey("sha-384-hmac", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey1 == nil {
		t.Fatal("Can't generate sha-384-hmac key\n")
	}
	fmt.Printf("Testing sha-384-hmac generation\n")
	printKey(cryptoKey8)
	fmt.Printf("\n")

	// "sha-512-hmac"
	keyName = "keyName9"
	keyEpoch = 2
	keyPurpose = "crypting"
	keyStatus = "active"
	cryptoKey9 := GenerateCryptoKey("sha-512-hmac", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey1 == nil {
		t.Fatal("Can't generate sha-512-hmac key\n")
	}
	fmt.Printf("Testing sha-512-hmac generation\n")
	printKey(cryptoKey9)
	fmt.Printf("\n")

	// "rsa-1024"
	keyName = "keyName10"
	keyEpoch = 2
	keyPurpose = "signing"
	keyStatus = "primary"
	cryptoKey10 := GenerateCryptoKey("rsa-1024", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey1 == nil {
		t.Fatal("Can't generate rsa-1024 key\n")
	}
	fmt.Printf("Testing rsa-1024 generation\n")
	printKey(cryptoKey10)
	fmt.Printf("\n")
	m10 := MarshalCryptoKey(*cryptoKey10)
	if m10 == nil {
		t.Fatal("Can't MarshalCryptoKey rsa-1024 key\n")
	}
	cryptoKey10_d, err := UnmarshalCryptoKey(m10)
	if err != nil {
		t.Fatal("Can't UnmarshalCryptoKey rsa-1024 key\n")
	}
	printKey(cryptoKey10_d)
	fmt.Printf("\n")

	// "rsa-2048"
	keyName = "keyName11"
	keyEpoch = 2
	keyPurpose = "signing"
	keyStatus = "primary"
	cryptoKey11 := GenerateCryptoKey("rsa-2048", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey11 == nil {
		t.Fatal("Can't generate rsa-2048 key\n")
	}
	fmt.Printf("Testing rsa-2048 generation\n")
	printKey(cryptoKey11)
	fmt.Printf("\n")

	// "rsa-3072"
	keyName = "keyName12"
	keyEpoch = 2
	keyPurpose = "signing"
	keyStatus = "primary"
	cryptoKey12 := GenerateCryptoKey("rsa-3072", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey12 == nil {
		t.Fatal("Can't generate rsa-3072 key\n")
	}
	fmt.Printf("Testing rsa-3072 generation\n")
	printKey(cryptoKey12)
	fmt.Printf("\n")

	// "ecdsa-P256"
	keyName = "keyName13"
	keyEpoch = 2
	keyPurpose = "signing"
	keyStatus = "primary"
	cryptoKey13 := GenerateCryptoKey("ecdsa-P256", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey13 == nil {
		t.Fatal("Can't generate ecdsa-P256 key\n")
	}
	fmt.Printf("Testing ecdsa-P256 generation\n")
	printKey(cryptoKey13)
	fmt.Printf("\n")

	// "ecdsa-P384"
	keyName = "keyName14"
	keyEpoch = 2
	keyPurpose = "signing"
	keyStatus = "primary"
	cryptoKey14 := GenerateCryptoKey("ecdsa-P384", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey14 == nil {
		t.Fatal("Can't generate ecdsa-P384 key\n")
	}
	fmt.Printf("Testing ecdsa-P384 generation\n")
	printKey(cryptoKey14)
	fmt.Printf("\n")
}

func TestVerifierSerialization(t *testing.T) {
	// func (v *Verifier) Verify(data []byte, context string, sig []byte) (bool, error) {
}

func TestSignerToPublic(t *testing.T) {
	// (s *Signer) GetVerifierFromSigner() *Verifier
}

func TestCanonicalizeKeyBytes(t *testing.T) {
	// func (v *Verifier) CanonicalKeyBytesFromVerifier() []byte
	// func (s *Signer) CanonicalKeyBytesFromSigner() []byte {
}

func TestSignerVerifierDERSerialization(t *testing.T) {
	// func (v *Verifier) MarshalKey() []byte {
	// func UnmarshalKey(material []byte) (*Verifier, error) {
	// func FromX509(cert *x509.Certificate) (*Verifier, error) {
	// func (v *Verifier) KeyEqual(cert *x509.Certificate) bool {
	// func UnmarshalVerifierProto(ck *CryptoKey) (*Verifier, error) {
}

func TestSelfSignedX509(t *testing.T) {
	// (s *Signer) CreateSelfSignedDER(pkAlg int, sigAlg int, sn int64, name *pkix.Name) ([]byte, error)
	// (s *Signer) CreateSelfSignedX509(pkAlg int, sigAlg int, sn int64,name *pkix.Name) (*x509.Certificate, error)
	// (s *Signer) CreateCRL(cert *x509.Certificate, revokedCerts []pkix.RevokedCertificate, now, expiry time.Time) ([]byte, error)
	// (s *Signer) CreateSignedX509(caCert *x509.Certificate, certSerial int, subjectKey *Verifier,
}

func TestSignAndVerify(t *testing.T) {

	var keyName string
	var keyEpoch int32
	var keyPurpose string
	var keyStatus string

	fmt.Printf("\n")
	keyName = "TestRsa2048SignandVerify"
	keyEpoch = 2
	keyPurpose = "signing"
	keyStatus = "primary"
	cryptoKey1 := GenerateCryptoKey("rsa-2048", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptoKey1 == nil {
		t.Fatal("Can't generate rsa-2048 key\n")
	}
	printKey(cryptoKey1)

	// save
	privateKey, err := PrivateKeyFromCryptoKey(*cryptoKey1)
	if err != nil {
		t.Fatal("PrivateKeyFromCryptoKey fails, ", err, "\n")
	}

	mesg := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	hash := crypto.SHA256
	blk := hash.New()
	blk.Write(mesg)
	hashed := blk.Sum(nil)
	fmt.Printf("Hashed: %x\n", hashed)

	// sig, err := privateKey.(*rsa.PrivateKey).Sign(rand.Reader, hashed, opts)
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
	if err != nil {
		t.Fatal("privateKey signing fails\n")
	}
	fmt.Printf("Signature: %x\n", sig)

	publicKey := privateKey.(*rsa.PrivateKey).PublicKey
	if err != nil {
		t.Fatal("PrivateKeyFromCryptoKey fails, ", err, "\n")
	}
	err = rsa.VerifyPKCS1v15(&publicKey, crypto.SHA256, hashed, sig)
	if err != nil {
		t.Fatal("Verify fails, ", err, "\n")
	} else {
		fmt.Printf("Rsa Verify succeeds\n")
	}

	/*
		s := SignerFromCryptoKey(*cryptoKey1)
		if s== nil {
			t.Fatal("SignerFromCryptoKey fails, ", err, "\n")
		}
		sig1, err := s.Sign(hashed, "Signing context")
		if err != nil {
			t.Fatal("Signer Sign failed, ", err, "\n")
		}
		fmt.Printf("Signer sign: %x\n", sig1)

		// verified, err := v.Verify(hashed, "Signing context", sig1)
	*/
}

func TestCerts(t *testing.T) {
	keyName := "keyName1"
	keyEpoch := int32(1)
	keyPurpose := "signing"
	keyStatus := "active"
	signingKey := GenerateCryptoKey("ecdsa-P256", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if signingKey == nil {
		t.Fatal("Can't generate signing key\n")
	}
	printKey(signingKey)
	fmt.Printf("\n")

	privateKey, err := PrivateKeyFromCryptoKey(*signingKey)
	if err != nil {
	}

	s := &Signer{
		header:     signingKey.KeyHeader,
		privateKey: privateKey,
	}

	details := &X509Details{
		CommonName:   proto.String("test"),
		Country:      proto.String("US"),
		State:        proto.String("WA"),
		Organization: proto.String("Google"),
	}
	der, err := s.CreateSelfSignedDER(int(x509.ECDSA), int(x509.ECDSAWithSHA256),
		int64(10), NewX509Name(details))
	if err != nil {
		t.Fatal("CreateSelfSignedDER failed, ", err, "\n")
	}
	fmt.Printf("Der: %x\n", der)
	cert, err := s.CreateSelfSignedX509(int(x509.ECDSA), int(x509.ECDSAWithSHA256),
		int64(10), NewX509Name(details))
	if err != nil {
		t.Fatal("CreateSelfSignedX509 failed, ", err, "\n")
	}
	fmt.Printf("Cert: %x\n", cert)
}

func TestCanonicalBytes(t *testing.T) {
	keyName := "keyName1"
	keyEpoch := int32(1)
	keyPurpose := "signing"
	keyStatus := "active"
	signingKey := GenerateCryptoKey("ecdsa-P256", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if signingKey == nil {
		t.Fatal("Can't generate signing key\n")
	}
	printKey(signingKey)
	fmt.Printf("\n")

	privateKey, err := PrivateKeyFromCryptoKey(*signingKey)
	if err != nil {
		t.Fatal("PrivateKeyFromCryptoKey fails\n")
	}

	s := &Signer{
		header:     signingKey.KeyHeader,
		privateKey: privateKey,
	}
	cb, err := s.CanonicalKeyBytesFromSigner()
	if err != nil {
		t.Fatal("CanonicalKeyBytesFromSigner fails\n")
	}
	fmt.Printf("Canonical bytes: %x\n", cb)
}

func TestNewCrypter(t *testing.T) {
	keyName := "keyName1"
	keyEpoch := int32(1)
	keyPurpose := "crypting"
	keyStatus := "active"
	cryptingKey := GenerateCryptoKey("aes-128-sha-256-cbc", &keyName, &keyEpoch, &keyPurpose, &keyStatus)
	if cryptingKey == nil {
		t.Fatal("Can't generate signing key\n")
	}
	printKey(cryptingKey)
	fmt.Printf("\n")
	c := CrypterFromCryptoKey(*cryptingKey)
	if c == nil {
		t.Fatal("CrypterFromCryptoKey fails\n")
	}
	plain := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	crypted, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal("Crypter failed to encrypt\n")
	}
	fmt.Printf("Encrypted: %x\n", crypted)
	decrypted, err := c.Decrypt(crypted)
	if err != nil {
		t.Fatal("Crypter failed to decrypt\n")
	}
	fmt.Printf("Decrypted: %x\n", decrypted)
	if !bytes.Equal(plain, decrypted) {
		t.Fatal("plain an decrypted bytes don't match\n")
	}
}

func TestEncryptAndDecrypt(t *testing.T) {
	// func (c *Crypter) Encrypt(data []byte) ([]byte, error) {
	// func (c *Crypter) Decrypt(ciphertext []byte) ([]byte, error) {
}

func TestDeriveSecret(t *testing.T) {
	// func (d *Deriver) Derive(salt, context, material []byte) error {

}
