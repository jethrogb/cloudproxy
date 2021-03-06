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
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/golang/protobuf/proto"
	"github.com/jlmucb/cloudproxy/go/tao/auth"
	"github.com/jlmucb/cloudproxy/go/util"
	"golang.org/x/crypto/pbkdf2"
)

// A KeyType represent the type(s) of keys held by a Keys struct.
type KeyType int

// These are the types of supported keys.
const (
	Signing KeyType = 1 << iota
	Crypting
	Deriving
)

// The Keys structure manages a set of signing, verifying, encrypting,
// and key-deriving // keys for many uses.  To some extent, the field
// meanings will differ between uses.  The comments below are focused
// on the use of the Keys structure for domains, including the policy
// domain, and Tao's (Root and Stacked).
type Keys struct {

	// This is the directory the Keys structure is saved to and
	// restored from.
	dir string

	// This is the policy for the keys, for example, governing
	// what the sealing/unsealing policy is.
	policy string

	// This represents the key types for the keys generated for this structure.
	// Note: the VerifyingKey is not generated for this structure so is
	// not included.
	keyTypes KeyType

	// This represents the private key used to sign statements.
	SigningKey *Signer

	// This represents the keys for the symmetric suite used to encrypt and
	// integrity protect data.
	CryptingKey *Crypter

	// This is the deriving key used to obtain keys from a master secret
	// like passwords in the case of domain keys.
	DerivingKey *Deriver

	// This represents the public key of the SigningKey.
	VerifyingKey *Verifier

	// This is an attestation by my host appointing the public key of
	// the Signing key.  This can be nil.
	Delegation *Attestation

	// This is the certificate for the signing key.
	// For a Root Tao, this cert is signed by the policy key or
	// other authority.  It can be nil.
	Cert *x509.Certificate

	// This is the certificate chain from the signer of Cert to the
	// policy key (or other authority).
	CertChain []*x509.Certificate
}

// The paths to the filename used by the Keys type.
const (
	X509VerifierPath    = "cert"
	PBEKeysetPath       = "keys"
	PBESignerPath       = "signer"
	SealedKeysetPath    = "sealed_keyset"
	PlaintextKeysetPath = "plaintext_keyset"
)

// X509VerifierPath returns the path to the verifier key, stored as an X.509
// certificate.
func (k *Keys) X509VerifierPath() string {
	if k.dir == "" {
		return ""
	}
	return path.Join(k.dir, X509VerifierPath)
}

// PBEKeysetPath returns the path for stored keys.
func (k *Keys) PBEKeysetPath() string {
	if k.dir == "" {
		return ""
	}
	return path.Join(k.dir, PBEKeysetPath)
}

// PBESignerPath returns the path for a stored signing key.
func (k *Keys) PBESignerPath() string {
	if k.dir == "" {
		return ""
	}
	return path.Join(k.dir, PBESignerPath)
}

// Generate or restore a signer.
// InitializeSigner uses marshaledCryptoKey to restore a signer from
// a serialized CryptoKey if it's not nil; otherwise it generates one.
// If generated, the remainder of the arguments are used as parameters;
// otherwise they are ignored.
func InitializeSigner(marshaledCryptoKey []byte, keyType string, keyName *string, keyEpoch *int32, keyPurpose *string, keyStatus *string) (*Signer, error) {
	if marshaledCryptoKey != nil {
		k, err := UnmarshalCryptoKey(marshaledCryptoKey)
		if err != nil {
			return nil, errors.New("Can't UnmarshalCryptoKey")
		}
		s := SignerFromCryptoKey(*k)
		if s == nil {
			k.Clear()
			return nil, errors.New("Can't SignerFromCryptoKey")
		}
		if s.Header.KeyPurpose == nil || *s.Header.KeyPurpose != "signing" {
			k.Clear()
			s.Clear()
			return nil, errors.New("Recovered key not a signer")
		}
	}
	k := GenerateCryptoKey(keyType, keyName, keyEpoch, keyPurpose, keyStatus)
	if k == nil {
		return nil, errors.New("Can't GenerateCryptoKey")
	}
	s := SignerFromCryptoKey(*k)
	if s == nil {
		k.Clear()
		return nil, errors.New("Can't SignerFromCryptoKey")
	}
	if s.Header.KeyPurpose == nil || *s.Header.KeyPurpose != "signing" {
		k.Clear()
		s.Clear()
		return nil, errors.New("Recovered key not a signer")
	}
	return s, nil
}

// Generate or restore a crypter.
// InitializeCrypter uses marshaledCryptoKey to restore a signer from
// a serialized CryptoKey if it's not nil; otherwise it generates one.
// If generated, the remainder of the arguments are used as parameters;
// otherwise they are ignored.
func InitializeCrypter(marshaledCryptoKey []byte, keyType string, keyName *string, keyEpoch *int32, keyPurpose *string, keyStatus *string) (*Crypter, error) {
	if marshaledCryptoKey != nil {
		k, err := UnmarshalCryptoKey(marshaledCryptoKey)
		if err != nil {
			return nil, errors.New("Can't UnmarshalCryptoKey")
		}
		c := CrypterFromCryptoKey(*k)
		if c == nil {
			k.Clear()
			return nil, errors.New("Can't CrypterFromCryptoKey")
		}
		if c.Header.KeyPurpose == nil || *c.Header.KeyPurpose != "crypting" {
			k.Clear()
			c.Clear()
			return nil, errors.New("Recovered key not a crypter")
		}
	}
	k := GenerateCryptoKey(keyType, keyName, keyEpoch, keyPurpose, keyStatus)
	if k == nil {
		return nil, errors.New("Can't GenerateCryptoKey")
	}
	c := CrypterFromCryptoKey(*k)
	if c == nil {
		k.Clear()
		return nil, errors.New("Can't CrypterFromCryptoKey")
	}
	if c.Header.KeyPurpose == nil || *c.Header.KeyPurpose != "crypting" {
		k.Clear()
		c.Clear()
		return nil, errors.New("Recovered key not a crypter")
	}
	return c, nil
}

// Generate or restore a deriver.
// InitializeDeriver uses marshaledCryptoKey to restore a signer from
// a serialized CryptoKey if it's not nil; otherwise it generates one.
// If generated, the remainder of the arguments are used as parameters;
// otherwise they are ignored.
func InitializeDeriver(marshaledCryptoKey []byte, keyType string, keyName *string, keyEpoch *int32, keyPurpose *string, keyStatus *string) (*Deriver, error) {
	if marshaledCryptoKey != nil {
		k, err := UnmarshalCryptoKey(marshaledCryptoKey)
		if err != nil {
			return nil, errors.New("Can't UnmarshalCryptoKey")
		}
		d := DeriverFromCryptoKey(*k)
		if d == nil {
			k.Clear()
			return nil, errors.New("Can't DeriverFromCryptoKey")
		}
		if d.Header.KeyPurpose == nil || *d.Header.KeyPurpose != "deriving" {
			k.Clear()
			return nil, errors.New("Recovered key not a deriver")
		}
	}
	k := GenerateCryptoKey(keyType, keyName, keyEpoch, keyPurpose, keyStatus)
	if k == nil {
		return nil, errors.New("Can't GenerateCryptoKey")
	}
	d := DeriverFromCryptoKey(*k)
	if d == nil {
		k.Clear()
		return nil, errors.New("Can't DeriverFromCryptoKey")
	}
	if d.Header.KeyPurpose == nil || *d.Header.KeyPurpose != "deriving" {
		k.Clear()
		d.Clear()
		return nil, errors.New("Recovered key not a deriver")
	}
	return d, nil
}

func PrintKeys(keys *Keys) {
	fmt.Printf("dir: %s\n", keys.dir)
	fmt.Printf("policy: %s\n", keys.policy)
	fmt.Printf("Key types: ")
	if keys.keyTypes&Signing != 0 {
		fmt.Printf("Signing ")
	}
	if keys.keyTypes&Crypting != 0 {
		fmt.Printf("Crypting ")
	}
	if keys.keyTypes&Deriving != 0 {
		fmt.Printf("Deriving ")
	}
	fmt.Printf("\n")
	if keys.SigningKey != nil {
		PrintCryptoKeyHeader(*keys.SigningKey.Header)
	}
	if keys.VerifyingKey != nil {
		PrintCryptoKeyHeader(*keys.VerifyingKey.Header)
	}
	if keys.CryptingKey != nil {
		PrintCryptoKeyHeader(*keys.CryptingKey.Header)
	}
	if keys.DerivingKey != nil {
		PrintCryptoKeyHeader(*keys.DerivingKey.Header)
	}
	if keys.Delegation != nil {
		fmt.Printf("Delegation present\n")
	} else {
		fmt.Printf("Delegation empty\n")
	}
	if keys.Cert != nil {
		fmt.Printf("Cert present\n")
	} else {
		fmt.Printf("Cert empty\n")
	}
}

// Encodes Keys to Cryptokeyset
func CryptoKeysetFromKeys(k *Keys) (*CryptoKeyset, error) {
	// fill in keys, cert, attestation
	var keyList [][]byte
	if k.keyTypes&Signing == Signing {
		ck := &CryptoKey{
			KeyHeader: k.SigningKey.Header,
		}
		keyComponents, err := KeyComponentsFromSigner(k.SigningKey)
		if err != nil {
			return nil, errors.New("Can't get key components from signing key")
		}
		ck.KeyComponents = keyComponents
		serializedCryptoKey, err := proto.Marshal(ck)
		if err != nil {
			return nil, errors.New("Can't serialize signing key")
		}
		keyList = append(keyList, serializedCryptoKey)
	}

	if k.keyTypes&Crypting == Crypting {
		ck := &CryptoKey{
			KeyHeader: k.CryptingKey.Header,
		}
		keyComponents, err := KeyComponentsFromCrypter(k.CryptingKey)
		if err != nil {
			return nil, errors.New("Can't get key components from crypting key")
		}
		ck.KeyComponents = keyComponents
		serializedCryptoKey, err := proto.Marshal(ck)
		if err != nil {
			return nil, errors.New("Can't serialize crypting key")
		}
		keyList = append(keyList, serializedCryptoKey)
	}

	if k.keyTypes&Deriving == Deriving {
		ck := &CryptoKey{
			KeyHeader: k.DerivingKey.Header,
		}
		keyComponents, err := KeyComponentsFromDeriver(k.DerivingKey)
		if err != nil {
			return nil, errors.New("Can't get key components from deriving key")
		}
		ck.KeyComponents = keyComponents
		serializedCryptoKey, err := proto.Marshal(ck)
		if err != nil {
			return nil, errors.New("Can't serialize deriving key")
		}
		keyList = append(keyList, serializedCryptoKey)
	}

	cks := &CryptoKeyset{
		Keys: keyList,
	}
	if k.Cert != nil {
		cks.Cert = k.Cert.Raw
	}
	cks.Delegation = k.Delegation
	for i := 0; i < len(k.CertChain); i++ {
		cks.CertChain = append(cks.CertChain, k.CertChain[i].Raw)
	}
	return cks, nil
}

// KeysFromCryptoKeyset decodes a CryptoKeyset into a Keys structure.
func KeysFromCryptoKeyset(cks *CryptoKeyset) (*Keys, error) {
	k := new(Keys)

	for i := 0; i < len(cks.Keys); i++ {
		var ck CryptoKey
		err := proto.Unmarshal(cks.Keys[i], &ck)
		if err != nil {
			return nil, errors.New("Can't unmarshal cryptokey")
		}
		if ck.KeyHeader.KeyType == nil {
			return nil, errors.New("Missing KeyType in CryptoHeader")
		}
		switch *ck.KeyHeader.KeyPurpose {
		default:
			return nil, errors.New("Unknown purpose")
		case "signing":
			k.SigningKey = SignerFromCryptoKey(ck)
			if k.SigningKey == nil {
				return nil, errors.New("Can't recover signing key from cryptokey")
			}
			k.keyTypes |= Signing
		case "crypting":
			k.CryptingKey = CrypterFromCryptoKey(ck)
			if k.CryptingKey == nil {
				return nil, errors.New("Can't recover crypting key from cryptokey")
			}
			k.keyTypes |= Crypting
		case "deriving":
			k.DerivingKey = DeriverFromCryptoKey(ck)
			if k.DerivingKey == nil {
				return nil, errors.New("Can't recover deriving key from cryptokey")
			}
			k.keyTypes |= Deriving
		}
	}
	k.Cert = nil
	if cks.Cert != nil {
		cert, err := x509.ParseCertificate(cks.Cert)
		if err != nil {
			return nil, err
		}
		k.Cert = cert
	}
	// CertChain
	for i := 0; i < len(cks.CertChain); i++ {
		cert, err := x509.ParseCertificate(cks.CertChain[i])
		if err != nil {
			return nil, err
		}
		k.CertChain = append(k.CertChain, cert)
	}
	k.Delegation = cks.Delegation
	return k, nil
}

func MarshalKeyset(cks *CryptoKeyset) ([]byte, error) {
	return proto.Marshal(cks)
}

func UnmarshalKeyset(buf []byte, cks *CryptoKeyset) error {
	return proto.Unmarshal(buf, cks)
}

func MarshalKeys(k *Keys) ([]byte, error) {
	cks, err := CryptoKeysetFromKeys(k)
	if err != nil {
		return nil, err
	}
	return MarshalKeyset(cks)
}

func UnmarshalKeys(b []byte) (*Keys, error) {
	var cks CryptoKeyset
	err := UnmarshalKeyset(b, &cks)
	if err != nil {
		return nil, errors.New("Can't unmarshal key set")
	}
	return KeysFromCryptoKeyset(&cks)
}

// SealedKeysetPath returns the path for a stored signing key.
func (k *Keys) SealedKeysetPath() string {
	if k.dir == "" {
		return ""
	}

	return path.Join(k.dir, SealedKeysetPath)
}

// PlaintextKeysetPath returns the path for a key stored in plaintext (this is
// not normally the case).
func (k *Keys) PlaintextKeysetPath() string {
	if k.dir == "" {
		return ""
	}

	return path.Join(k.dir, PlaintextKeysetPath)
}

// NewTemporaryKeys creates a new Keys structure with the specified keys.
func NewTemporaryKeys(keyTypes KeyType) (*Keys, error) {
	k := &Keys{
		keyTypes: keyTypes,
	}
	if k.keyTypes == 0 || (k.keyTypes & ^Signing & ^Crypting & ^Deriving != 0) {
		return nil, errors.New("bad key type")
	}

	var err error
	if k.keyTypes&Signing == Signing {
		k.SigningKey = GenerateAnonymousSigner()
		if k.SigningKey == nil {
			return nil, errors.New("Can't generate signer")
		}
		k.VerifyingKey = k.SigningKey.GetVerifierFromSigner()
		if k.VerifyingKey == nil {
			return nil, errors.New("Can't get verifier from signer")
		}
	}

	if k.keyTypes&Crypting == Crypting {
		k.CryptingKey = GenerateAnonymousCrypter()
		if k.CryptingKey == nil {
			return nil, errors.New("Can't generate crypter")
		}
	}

	if k.keyTypes&Deriving == Deriving {
		k.DerivingKey = GenerateAnonymousDeriver()
		if k.DerivingKey == nil {
			return nil, err
		}
	}
	return k, nil
}

// NewSignedOnDiskPBEKeys creates the same type of keys as NewOnDiskPBEKeys but
// signs a certificate for the signer with the provided Keys, which must have
// both a SigningKey and a Certificate.
func NewSignedOnDiskPBEKeys(keyTypes KeyType, password []byte, path string, name *pkix.Name, serial int, signer *Keys) (*Keys, error) {
	if signer == nil || name == nil {
		return nil, errors.New("must supply a signer and a name")
	}

	if signer.Cert == nil || signer.SigningKey == nil {
		return nil, newError("the signing key must have a SigningKey and a Cert")
	}

	if keyTypes&Signing == 0 {
		return nil, errors.New("can't sign a key that has no signer")
	}

	k, err := NewOnDiskPBEKeys(keyTypes, password, path, nil)
	if err != nil {
		return nil, err
	}

	// If there's already a cert, then this means that there was already a
	// keyset on disk, so don't create a new signed certificate.
	if k.Cert == nil {
		pkInt := PublicKeyAlgFromSignerAlg(*signer.SigningKey.Header.KeyType)
		sigInt := SignatureAlgFromSignerAlg(*signer.SigningKey.Header.KeyType)
		k.Cert, err = signer.SigningKey.CreateSignedX509(signer.Cert, serial, k.VerifyingKey, pkInt, sigInt, name)
		if err != nil {
			return nil, err
		}

		if err = util.WritePath(k.X509VerifierPath(), k.Cert.Raw, 0777, 0666); err != nil {
			return nil, err
		}
	}

	return k, nil
}

// NewOnDiskPBEKeys creates a new Keys structure with the specified key types
// store under PBE on disk. If keys are generated and name is not nil, then a
// self-signed x509 certificate will be generated and saved as well.
func NewOnDiskPBEKeys(keyTypes KeyType, password []byte, path string, name *pkix.Name) (*Keys, error) {
	if keyTypes == 0 || (keyTypes & ^Signing & ^Crypting & ^Deriving != 0) {
		return nil, newError("bad key type")
	}

	if path == "" {
		return nil, newError("bad init call: no path for keys")
	}

	if len(password) == 0 {
		// This means there's no secret information: just load a public verifying key.
		if keyTypes & ^Signing != 0 {
			return nil, newError("without a password, only a verifying key can be loaded")
		}

		k := &Keys{
			keyTypes: keyTypes,
			dir:      path,
		}
		cert, err := loadCert(k.X509VerifierPath())
		if err != nil {
			return nil, errors.New("Couldn't load cert")
		}
		if cert == nil {
			return nil, errors.New("Empty cert")
		}
		k.Cert = cert
		v, err := VerifierFromX509(cert)
		if err != nil {
			return nil, err
		}
		k.VerifyingKey = v
		return k, nil
	} else {
		// Check to see if there are already keys.
		tk := &Keys{
			keyTypes: keyTypes,
			dir:      path,
		}
		f, err := os.Open(tk.PBEKeysetPath())
		if err == nil {
			defer f.Close()
			ks, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			}

			data, err := PBEDecrypt(ks, password)
			if err != nil {
				return nil, err
			}
			defer ZeroBytes(data)

			k, err := UnmarshalKeys(data)
			if err != nil {
				return nil, err
			}
			k.dir = path

			// Note that this loads the certificate if it's
			// present, and it returns nil otherwise.
			if k.Cert == nil {
				cert, err := loadCert(k.X509VerifierPath())
				if err != nil {
					return nil, err
				}
				k.Cert = cert
				if k.Cert == nil {
					k.VerifyingKey = k.SigningKey.GetVerifierFromSigner()
				} else {
					k.VerifyingKey, err = VerifierFromX509(cert)
					if err != nil {
						return nil, errors.New("Can't get verifying key (1)")
					}
				}
			} else {
				k.VerifyingKey, err = VerifierFromX509(k.Cert)
				if err != nil {
					return nil, errors.New("Can't get verifying key (2)")
				}
			}
			return k, nil
		} else {
			// Create and store a new set of keys.
			k, err := NewTemporaryKeys(keyTypes)
			if err != nil {
				return nil, err
			}
			k.dir = path

			// Use correct cert name
			signerAlg := SignerTypeFromSuiteName(TaoCryptoSuite)
			if signerAlg == nil {
				return nil, errors.New("SignerTypeFromSuiteName failed")
			}
			pkInt := PublicKeyAlgFromSignerAlg(*signerAlg)
			skInt := SignatureAlgFromSignerAlg(*signerAlg)
			if pkInt < 0 || skInt < 0 {
				return nil, errors.New("Can't get PublicKeyAlgFromSignerAlg")
			}

			// reset cert and verifying keys
			if name == nil {
				us := "US"
				textName := "Some Tao service"
				name = &pkix.Name{
					Organization: []string{textName},
					CommonName:   textName,
					Country:      []string{us},
				}
			}
			cert, err := k.SigningKey.CreateSelfSignedX509(pkInt, skInt, int64(1), name)
			if err != nil {
				return nil, errors.New("Can't create self signing cert")
			}
			k.Cert = cert
			k.VerifyingKey, err = VerifierFromX509(cert)

			m, err := MarshalKeys(k)
			if err != nil {
				return nil, err
			}
			defer ZeroBytes(m)

			enc, err := PBEEncrypt(m, password)
			if err != nil {
				return nil, err
			}

			if err = util.WritePath(k.PBEKeysetPath(), enc, 0777, 0600); err != nil {
				return nil, err
			}

			if k.SigningKey != nil && name != nil {
				err = k.newCert(name)
				if err != nil {
					return nil, err
				}
			}
			return k, nil
		}
	}
	return nil, errors.New("Shouldnt happen")

}

func (k *Keys) newCert(name *pkix.Name) (err error) {
	pkInt := PublicKeyAlgFromSignerAlg(*k.SigningKey.Header.KeyType)
	sigInt := SignatureAlgFromSignerAlg(*k.SigningKey.Header.KeyType)
	if pkInt < 0 || sigInt < 0 {
		return errors.New("No signing algorithm")
	}
	k.Cert, err = k.SigningKey.CreateSelfSignedX509(pkInt, sigInt, int64(1), name)
	if err != nil {
		return err
	}
	if err = util.WritePath(k.X509VerifierPath(), k.Cert.Raw, 0777, 0666); err != nil {
		return err
	}
	return nil
}

func loadCert(path string) (*x509.Certificate, error) {
	f, err := os.Open(path)
	// Allow this operation to fail silently, since there isn't always a
	// certificate available.
	if err != nil {
		return nil, nil
	}
	defer f.Close()

	der, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(der)
}

// NewTemporaryTaoDelegatedKeys initializes a set of temporary keys under a host
// Tao, using the Tao to generate a delegation for the signing key. Since these
// keys are never stored on disk, they are not sealed to the Tao.
func NewTemporaryTaoDelegatedKeys(keyTypes KeyType, t Tao) (*Keys, error) {
	k, err := NewTemporaryKeys(keyTypes)
	if err != nil {
		return nil, err
	}

	if t != nil && k.SigningKey != nil {

		self, err := t.GetTaoName()
		if err != nil {
			return nil, err
		}

		s := &auth.Speaksfor{
			Delegate:  k.SigningKey.ToPrincipal(),
			Delegator: self,
		}
		if k.Delegation, err = t.Attest(&self, nil, nil, s); err != nil {
			return nil, err
		}
	}

	return k, nil
}

// PBEEncrypt encrypts plaintext using a password to generate a key. Note that
// since this is for private program data, we don't try for compatibility with
// the C++ Tao version of the code.
func PBEEncrypt(plaintext, password []byte) ([]byte, error) {
	if password == nil || len(password) == 0 {
		return nil, newError("null or empty password")
	}
	cipherName := CipherTypeFromSuiteName(TaoCryptoSuite)
	hashName := HashTypeFromSuiteName(TaoCryptoSuite)
	if cipherName == nil || hashName == nil {
		return nil, errors.New("Bad cipher in PBEEncrypt")
	}
	pbed := &PBEData{
		Version: CryptoVersion_CRYPTO_VERSION_2.Enum(),
		Cipher:  proto.String(*cipherName),
		Hmac:    proto.String(*hashName),
		// The IV is required, so we include it, but this algorithm doesn't use it.
		Iv:         make([]byte, aes.BlockSize),
		Iterations: proto.Int32(4096),
		Salt:       make([]byte, aes.BlockSize),
	}

	// We use the first half of the salt for the AES key and the second
	// half for the HMAC key, since the standard recommends at least 8
	// bytes of salt.
	if _, err := rand.Read(pbed.Salt); err != nil {
		return nil, err
	}

	var aesKey []byte
	var hmacKey []byte
	switch TaoCryptoSuite  {
	default:
		return nil, errors.New("Unsupported cipher suite")
	case Basic128BitCipherSuite:
		// 128-bit AES key.
		aesKey = pbkdf2.Key(password, pbed.Salt[:8], int(*pbed.Iterations), 16, sha256.New)
		defer ZeroBytes(aesKey)
		// 64-byte HMAC-SHA256 key.
		hmacKey = pbkdf2.Key(password, pbed.Salt[8:], int(*pbed.Iterations), 32, sha256.New)
		defer ZeroBytes(hmacKey)
	case Basic192BitCipherSuite:
		// 256-bit AES key.
		aesKey = pbkdf2.Key(password, pbed.Salt[:8], int(*pbed.Iterations), 32, sha512.New384)
		defer ZeroBytes(aesKey)
		// 48-byte HMAC-SHA384 key.
		hmacKey = pbkdf2.Key(password, pbed.Salt[8:], int(*pbed.Iterations), 48, sha512.New384)
		defer ZeroBytes(hmacKey)
	case Basic256BitCipherSuite:
		// 256-bit AES key.
		aesKey = pbkdf2.Key(password, pbed.Salt[:8], int(*pbed.Iterations), 32, sha512.New)
		defer ZeroBytes(aesKey)
		// 64-byte HMAC-SHA512 key.
		hmacKey = pbkdf2.Key(password, pbed.Salt[8:], int(*pbed.Iterations), 64, sha512.New)
		defer ZeroBytes(hmacKey)
	}

	ver := CryptoVersion_CRYPTO_VERSION_2
	keyName := "PBEKey"
	keyEpoch := int32(1)
	keyType := CrypterTypeFromSuiteName(TaoCryptoSuite)
	keyPurpose := "crypting"
	keyStatus := "active"
	ch := &CryptoHeader{
		Version:    &ver,
		KeyName:    &keyName,
		KeyEpoch:   &keyEpoch,
		KeyType:    keyType,
		KeyPurpose: &keyPurpose,
		KeyStatus:  &keyStatus,
	}
	ck := &CryptoKey{
		KeyHeader: ch,
	}
	ck.KeyComponents = append(ck.KeyComponents, aesKey)
	ck.KeyComponents = append(ck.KeyComponents, hmacKey)
	c := CrypterFromCryptoKey(*ck)
	if c == nil {
		return nil, errors.New("Empty crypter")
	}
	// Note that we're abusing the PBEData format here, since the IV and
	// the MAC are actually contained in the ciphertext from Encrypt().
	var err error
	if pbed.Ciphertext, err = c.Encrypt(plaintext); err != nil {
		return nil, err
	}
	return proto.Marshal(pbed)
}

// PBEDecrypt decrypts ciphertext using a password to generate a key. Note that
// since this is for private program data, we don't try for compatibility with
// the C++ Tao version of the code.
func PBEDecrypt(ciphertext, password []byte) ([]byte, error) {
	if password == nil || len(password) == 0 {
		return nil, newError("null or empty password")
	}

	var pbed PBEData
	if err := proto.Unmarshal(ciphertext, &pbed); err != nil {
		return nil, err
	}

	// Recover the keys from the password and the PBE header.
	if *pbed.Version != CryptoVersion_CRYPTO_VERSION_2 {
		return nil, newError("bad version")
	}

	var aesKey []byte
	var hmacKey []byte
	switch TaoCryptoSuite  {
	default:
		return nil, errors.New("Unsupported cipher suite")
	case Basic128BitCipherSuite:
		// 128-bit AES key.
		aesKey = pbkdf2.Key(password, pbed.Salt[:8], int(*pbed.Iterations), 16, sha256.New)
		defer ZeroBytes(aesKey)
		// 64-byte HMAC-SHA256 key.
		hmacKey = pbkdf2.Key(password, pbed.Salt[8:], int(*pbed.Iterations), 32, sha256.New)
	case Basic192BitCipherSuite:
		// 256-bit AES key.
		aesKey = pbkdf2.Key(password, pbed.Salt[:8], int(*pbed.Iterations), 32, sha512.New384)
		defer ZeroBytes(aesKey)
		// 48-byte HMAC-SHA384 key.
		hmacKey = pbkdf2.Key(password, pbed.Salt[8:], int(*pbed.Iterations), 48, sha512.New384)
		defer ZeroBytes(hmacKey)
	case Basic256BitCipherSuite:
		// 256-bit AES key.
		aesKey = pbkdf2.Key(password, pbed.Salt[:8], int(*pbed.Iterations), 32, sha512.New)
		defer ZeroBytes(aesKey)
		// 64-byte HMAC-SHA512 key.
		hmacKey = pbkdf2.Key(password, pbed.Salt[8:], int(*pbed.Iterations), 64, sha512.New)
		defer ZeroBytes(hmacKey)
	}

	ck := new(CryptoKey)
	ver := CryptoVersion_CRYPTO_VERSION_2
	keyName := "PBEKey"
	keyEpoch := int32(1)
	keyType := CrypterTypeFromSuiteName(TaoCryptoSuite)
	if keyType == nil {
		return nil, newError("bad CrypterTypeFromSuiteName")
	}
	keyPurpose := "crypting"
	keyStatus := "active"
	ch := &CryptoHeader{
		Version:    &ver,
		KeyName:    &keyName,
		KeyEpoch:   &keyEpoch,
		KeyType:    keyType,
		KeyPurpose: &keyPurpose,
		KeyStatus:  &keyStatus,
	}
	ck.KeyHeader = ch
	ck.KeyComponents = append(ck.KeyComponents, aesKey)
	ck.KeyComponents = append(ck.KeyComponents, hmacKey)
	c := CrypterFromCryptoKey(*ck)

	defer ZeroBytes(hmacKey)

	// Note that we're abusing the PBEData format here, since the IV and
	// the MAC are actually contained in the ciphertext from Encrypt().
	data, err := c.Decrypt(pbed.Ciphertext)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// NewOnDiskTaoSealedKeys sets up the keys sealed under a host Tao or reads sealed keys.
func NewOnDiskTaoSealedKeys(keyTypes KeyType, t Tao, path, policy string) (*Keys, error) {

	// Fail if no parent Tao exists (otherwise t.Seal() would not be called).
	if t == nil {
		return nil, errors.New("parent tao is nil")
	}

	k := &Keys{
		keyTypes: keyTypes,
		dir:      path,
		policy:   policy,
	}

	// Check if keys exist: if not, generate and save a new set.
	f, err := os.Open(k.SealedKeysetPath())
	if err != nil {
		k, err = NewTemporaryTaoDelegatedKeys(keyTypes, t)
		if err != nil {
			return nil, err
		}
		k.dir = path
		k.policy = policy

		if err = k.Save(t); err != nil {
			return k, err
		}
		return k, nil
	}
	f.Close()

	// Otherwise, load from file.
	return LoadKeys(keyTypes, t, path, policy)
}

// Save serializes, seals, and writes a key set to disk. It calls t.Seal().
func (k *Keys) Save(t Tao) error {
	// Marshal key set.
	m, err := MarshalKeys(k)
	if err != nil {
		return err
	}
	// cks.Delegation = k.Delegation

	// TODO(tmroeder): defer zeroKeyset(cks)
	defer ZeroBytes(m)

	data, err := t.Seal(m, k.policy)
	if err != nil {
		return err
	}

	if err = util.WritePath(k.SealedKeysetPath(), data, 0700, 0600); err != nil {
		return err
	}

	return nil
}

// LoadKeys reads a key set from file. If there is a parent tao (t!=nil), then
// expect the keys are sealed and call t.Unseal(); otherwise, expect the key
// set to be plaintext.
func LoadKeys(keyTypes KeyType, t Tao, path, policy string) (*Keys, error) {
	k := &Keys{
		keyTypes: keyTypes,
		dir:      path,
		policy:   policy,
	}

	// Check to see if there are already keys.
	var keysetPath string
	if t == nil {
		keysetPath = k.PlaintextKeysetPath()
	} else {
		keysetPath = k.SealedKeysetPath()
	}
	f, err := os.Open(keysetPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ks, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if t != nil {
		data, p, err := t.Unseal(ks)
		if err != nil {
			return nil, err
		}
		defer ZeroBytes(data)

		if p != policy {
			return nil, errors.New("invalid policy from Unseal")
		}
		return UnmarshalKeys(data)
	} else {
		return UnmarshalKeys(ks)
	}
	return nil, errors.New("Shouldnt happen")
}

// NewSecret creates and encrypts a new secret value of the given length, or it
// reads and decrypts the value and checks that it's the right length. It
// creates the file and its parent directories if these directories do not
// exist.
func (k *Keys) NewSecret(file string, length int) ([]byte, error) {
	if _, err := os.Stat(file); err != nil {
		// Create the parent directories and the file.
		if err := os.MkdirAll(path.Dir(file), 0700); err != nil {
			return nil, err
		}

		secret := make([]byte, length)
		if _, err := rand.Read(secret); err != nil {
			return nil, err
		}

		enc, err := k.CryptingKey.Encrypt(secret)
		if err != nil {
			return nil, err
		}

		if err := ioutil.WriteFile(file, enc, 0700); err != nil {
			return nil, err
		}

		return secret, nil
	}

	enc, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	dec, err := k.CryptingKey.Decrypt(enc)
	if err != nil {
		return nil, err
	}

	if len(dec) != length {
		ZeroBytes(dec)
		return nil, newError("The decrypted value had length %d, but it should have had length %d", len(dec), length)
	}

	return dec, nil
}

// SaveKeyset serializes and saves a Keys object to disk in plaintext.
func SaveKeyset(k *Keys, dir string) error {
	k.dir = dir
	m, err := MarshalKeys(k)
	if err != nil {
		return err
	}

	if err = util.WritePath(k.PlaintextKeysetPath(), m, 0700, 0600); err != nil {
		return err
	}

	return nil
}
