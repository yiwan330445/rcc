package peercc

type (
	Partquery struct {
		Catalog string
		Reply   chan string
	}
	Partqueries chan *Partquery
)
