package main

import (
	"log"

	"artship/internal/command"
)

func main() {
	if err := command.Run(); err != nil {
		log.Fatal(err)
	}
}
