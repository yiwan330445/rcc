package cloud_test

import (
	"strings"
	"testing"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/hamlet"
)

func TestCannotCreateClientForBadEndpoint(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := cloud.NewClient("http://some.server.com/endpoint")
	must_be.Nil(sut)
	wont_be.Nil(err)
	must_be.True(strings.HasPrefix(err.Error(), "Endpoint '"))

	sut, err = cloud.NewClient("some.server.com/endpoint")
	must_be.Nil(sut)
	wont_be.Nil(err)
	must_be.True(strings.HasPrefix(err.Error(), "Endpoint '"))
}

func TestCanCreateClient(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := cloud.NewClient("https://some.server.com/endpoint")
	wont_be.Nil(sut)
	must_be.Nil(err)
}
