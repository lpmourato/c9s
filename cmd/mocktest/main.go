package main

import (
	"log"

	"github.com/lpmourato/c9s/internal/ui"
)

func main() {
	if err := ui.StartMockLogView(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
