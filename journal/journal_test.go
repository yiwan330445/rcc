package journal_test

import (
	"testing"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/journal"
)

func TestJounalCanBeCalled(t *testing.T) {
	must, wont := hamlet.Specifications(t)

	wont.Nil(must)

	must.Equal("foo bar", journal.Unify("  foo  \t  \r\n   bar  "))

	common.ControllerType = "unittest"

	must.Nil(journal.Post("unittest", "journal", "from journal/journal_test.go"))
}
