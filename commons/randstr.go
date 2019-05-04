package commons

import (
	"crypto/rand"
	"encoding/base64"
)

// Base64RandomString returns a random, base64 encoded string, representing
// nBytes of random data.
func Base64RandomString(nBytes uint) string {
	data := make([]byte, nBytes)
	rand.Read(data)
	return base64.RawURLEncoding.EncodeToString(data)
}
