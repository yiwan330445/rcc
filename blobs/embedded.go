package blobs

import (
	"embed"
)

//go:embed assets/*.yaml docs/*.md
//go:embed assets/*.zip assets/man/*.txt
var content embed.FS

func Asset(name string) ([]byte, error) {
	return content.ReadFile(name)
}

func MustAsset(name string) []byte {
	body, err := Asset(name)
	if err != nil {
		panic(err)
	}
	return body
}
