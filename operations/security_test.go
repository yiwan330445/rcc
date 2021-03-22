package operations_test

import (
	"crypto/rsa"
	"strings"
	"testing"

	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/operations"
)

func TestCanCreatePrivateKey(t *testing.T) {
	must, wont := hamlet.Specifications(t)

	key, err := operations.GenerateEphemeralKey()
	must.Nil(err)
	wont.Nil(key)
	wont.Nil(key.Public())
	publicKey, ok := key.Public().(*rsa.PublicKey)
	must.True(ok)
	must.Equal(256, publicKey.Size())
	must.Equal(360, len(key.PublicDER()))
	must.Equal(426, len(key.PublicPEM()))
	body, err := key.RequestObject(nil)
	must.Nil(err)
	must.Equal(493, len(body))
	textual := string(body)
	must.True(strings.Contains(textual, "encryption"))
	must.True(strings.Contains(textual, "scheme"))
	must.True(strings.Contains(textual, "publicKey"))
	reader, err := key.RequestBody("hello, world!")
	must.Nil(err)
	wont.Nil(reader)
}
