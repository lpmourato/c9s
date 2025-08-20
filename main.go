package main

import (
	"log"

	"github.com/lpmourato/c9s/internal/ui"
	"github.com/lpmourato/c9s/internal/views"
)

func main() {
	app := ui.NewApp()
	views.NewCloudRunView(app)

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
