package beacon

import "crypto/rsa"

// RegistrationInfo defines the structure that holds information returned from the server about our device.
type RegistrationInfo struct {
	serverKey *rsa.PublicKey
	deviceID  string
}
