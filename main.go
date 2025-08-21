package main

import (
	"flag"
	"log"

	"github.com/lpmourato/c9s/internal/config"
	"github.com/lpmourato/c9s/internal/domain/cloudrun"
	"github.com/lpmourato/c9s/internal/gcp"
	"github.com/lpmourato/c9s/internal/mocks"
	"github.com/lpmourato/c9s/internal/ui"
	"github.com/lpmourato/c9s/internal/views"
)

func main() {
	// Parse command line flags
	testMode := flag.Bool("test", false, "Run in test mode with mock data")
	flag.Parse()

	app := ui.NewApp()

	// Create config with default settings
	cfg := &config.CloudRunConfig{
		ProjectID: "", // Will be set via command
		Region:    "", // Will be set via command
	}

	// Create service provider based on mode
	var provider cloudrun.ServiceProvider
	var err error

	if *testMode {
		provider = mocks.NewMockProvider()
	} else {
		provider, err = gcp.NewServiceProvider()
		if err != nil {
			log.Fatalf("Error creating GCP service provider: %v", err)
		}
	}

	views.NewCloudRunView(app, cfg, provider)

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
