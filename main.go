package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/lpmourato/c9s/internal/gcp"
	"github.com/lpmourato/c9s/internal/ui"
	"github.com/lpmourato/c9s/internal/views"
)

var (
	project  = flag.String("project", "", "GCP project ID")
	region   = flag.String("region", "", "GCP region")
	testMode = flag.Bool("test", false, "Run in test mode with mock data")
)

func main() {
	flag.Parse()

	// In test mode, use mock values
	if *testMode {
		*project = "mock-project"
		*region = "us-central1"
	} else {
		// Use environment variables as fallback
		if *project == "" {
			*project = os.Getenv("GCP_PROJECT")
		}
		if *region == "" {
			*region = os.Getenv("GCP_REGION")
		}
	}

	ctx := context.Background()

	var client *gcp.Client
	var err error

	if *testMode {
		// Create test client with mock data
		client = gcp.NewTestClient(*project, *region)
	} else {
		// Create real client
		client, err = gcp.NewClient(ctx, *project, *region)
		if err != nil {
			log.Fatalf("Failed to create GCP client: %v", err)
		}
	}

	app := ui.NewApp()
	mainView := views.NewCloudRunView(ctx, client, app)

	// Set up pages
	app.GetPages().AddPage("main", mainView, true, true)

	// Start the refresh timer in the background
	go mainView.StartRefreshTimer(ctx)

	// Run the app (this blocks until the app exits)
	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
