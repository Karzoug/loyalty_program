package main

import (
	"log"

	"github.com/Karzoug/loyalty_program/internal/config"
)

func main() {
	_, err := config.Read()
	if err != nil {
		log.Fatalf("Read config error: %s", err)
	}
}
