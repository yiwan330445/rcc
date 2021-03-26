package settings_test

import (
	"testing"

	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/settings"
)

func TestCanCallEntropyFunction(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut, err := settings.SummonSettings()
	must_be.Nil(err)
	wont_be.Nil(sut)

	wont_be.Nil(settings.Global)
	must_be.True(len(settings.Global.Hostnames()) > 1)

	must_be.Equal("https://robocorp.com/docs/hello.html", settings.Global.DocsLink("hello.html"))
	must_be.Equal("https://robocorp.com/docs/products/manual.html", settings.Global.DocsLink("products/manual.html"))
}
