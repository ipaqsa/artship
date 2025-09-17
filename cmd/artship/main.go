package main

import (
	"log"

	"github.com/ipaqsa/artship/internal/command"
)

func main() {
	if err := command.Run(); err != nil {
		log.Fatal(err)
	}
}
