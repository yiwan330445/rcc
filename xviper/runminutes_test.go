package xviper_test

import (
	"testing"
	"time"

	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/xviper"
)

func TestCanCreateRunMinutes(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut := xviper.RunMinutes()
	wont_be.Nil(sut)
	time.Sleep(100 * time.Millisecond)
	first := sut.Done()
	must_be.True(first > 0)
	second := xviper.RunMinutes().Done()
	must_be.True(second > first)
	must_be.Equal(first+1, second)
}
