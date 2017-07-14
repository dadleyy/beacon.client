package security

import "io"
import "fmt"
import "io/ioutil"
import "crypto"
import "crypto/rsa"
import "crypto/x509"
import "crypto/sha256"
import "encoding/pem"
import "encoding/hex"

// DeviceKey objects contain the rsa private key used to secure communications w/ the api
type DeviceKey struct {
	*rsa.PrivateKey
}

// SharedSecret returns the string version of the rsa public key
func (key *DeviceKey) SharedSecret() (string, error) {
	publicKeyData, e := x509.MarshalPKIXPublicKey(key.Public())

	if e != nil {
		return "", e
	}

	return hex.EncodeToString(publicKeyData), nil
}

// Decrypt implements crypto.Decrypter
func (key *DeviceKey) Decrypt(rand io.Reader, encodedMessage []byte, opts crypto.DecrypterOpts) ([]byte, error) {
	decoded, err := rsa.DecryptOAEP(sha256.New(), rand, key.PrivateKey, encodedMessage, []byte("beacon"))
	return decoded, err
}

// ReadDeviceKeyFromFile returns a new device key from a filename
func ReadDeviceKeyFromFile(filename string) (*DeviceKey, error) {
	privateKeyData, e := ioutil.ReadFile(filename)

	if e != nil {
		return nil, e
	}

	privateBlock, _ := pem.Decode(privateKeyData)

	if privateBlock == nil {
		return nil, fmt.Errorf("invalid-pem")
	}

	privateKey, e := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)

	if e != nil {
		return nil, e
	}

	return &DeviceKey{privateKey}, nil
}
