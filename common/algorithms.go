package common

import (
	"crypto/sha256"
	"fmt"
	"math"
	"time"

	"github.com/dchest/siphash"
)

func Entropy(input []byte) float64 {
	size := len(input)
	if size == 0 {
		return 0.0
	}
	frequencies := make([]int, 256)
	for _, value := range input {
		frequencies[int(value)] += 1
	}
	total := float64(size)
	entropy := 0.0
	for _, count := range frequencies {
		if count == 0 {
			continue
		}
		weight := float64(count) / total
		entropy += (-weight * math.Log2(weight))
	}
	return entropy / 8.0
}

func Hexdigest(raw []byte) string {
	return fmt.Sprintf("%02x", raw)
}

func ShortDigest(content string) string {
	digester := sha256.New()
	digester.Write([]byte(content))
	result := Hexdigest(digester.Sum(nil))
	return result[:16]
}

func Digest(content string) string {
	digester := sha256.New()
	digester.Write([]byte(content))
	return Hexdigest(digester.Sum(nil))
}

func Siphash(left, right uint64, body []byte) uint64 {
	return siphash.Hash(left, right, body)
}

func DayCountSince(timestamp time.Time) int {
	duration := time.Since(timestamp)
	days := math.Floor(duration.Hours() / 24.0)
	return int(days)
}
