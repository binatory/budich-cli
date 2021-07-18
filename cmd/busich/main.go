package main

import (
	"os"

	"io.github.binatory/nhac-cli/internal/domain"
)

func main() {
	b := domain.NewBusich(domain.NewConnectorZingMp3(), os.Stdout)

	switch os.Args[1] {
	case "search":
		b.Search(os.Args[2])
	case "play":
		b.Play(os.Args[2])
	default:
		panic("invalid command")
	}
}
