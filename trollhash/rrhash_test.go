package trollhash_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/trollhash"
)

func TestUsingRollingHashWorks(t *testing.T) {
	must, wont := hamlet.Specifications(t)

	must.Equal(12345, 12345^0)
	must.Equal(9876543210, 9876543210^0)

	must.Equal(256, len(trollhash.Seedlings()))

	must.Equal(uint64(0x2f203fec99682f8d), trollhash.Hash([]byte("A")))
	must.Equal(uint64(0x1443e52adddcf08e), trollhash.Hash([]byte("B")))
	must.Equal(uint64(0xef09ff3b737a55fc), trollhash.Hash([]byte("C")))
	must.Equal(uint64(0x6d421a4e169d8ce7), trollhash.Hash([]byte("AB")))
	must.Equal(uint64(0x85192d4bc79632c7), trollhash.Hash([]byte("ABC")))

	rolling := trollhash.Find("loha")
	wont.Nil(rolling)
	result := make([]int64, 0)
	for _, step := range []byte("O aloha! Aloha, Holoham!") {
		ok, at := rolling(step)
		if ok {
			result = append(result, at)
		}
	}
	must.Equal([]int64{3, 10, 18}, result)
	must.Equal(uint64(0xbe16aca9b15d96fa), trollhash.Hash([]byte("loha")))
	limit := 256
	for key := 0; key < 256; key++ {
		result := make(map[uint64]bool)
		flow := make([]byte, 0, limit)
		for size := 0; size < limit; size++ {
			flow = append(flow, byte(key))
			result[trollhash.Hash(flow)] = true
		}
		must.Equal(256, len(flow))
		must.True(63 < len(result))
		must.True(len(result) < 129)
	}
	limit = 10240
	uniques := make(map[uint64]bool)
	flow := make([]byte, 0, limit)
	source := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(source)
	for count := 0; count < limit; count++ {
		flow = append(flow, byte(rnd.Uint32()))
		uniques[trollhash.Hash(flow)] = true
	}
	must.Equal(10240, len(flow))
	must.Equal(10240, len(uniques))
}
