package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/lpmourato/c9s/internal/gcp"
	"github.com/lpmourato/c9s/internal/ui"
	"github.com/lpmourato/c9s/internal/views"
)

var (
	project = flag.String("project", "", "GCP project ID")
	region  = flag.String("region", "", "GCP region")
)

func main() {
	flag.Parse()

	// Use environment variables as fallback
	if *project == "" {
		*project = os.Getenv("GCP_PROJECT")
	}
	if *region == "" {
		*region = os.Getenv("GCP_REGION")
	}

	if *project == "" || *region == "" {
		log.Fatal("Project ID and region are required. Set via flags or GCP_PROJECT/GCP_REGION environment variables")
	}

	ctx := context.Background()
	client, err := gcp.NewClient(ctx, *project, *region)
	if err != nil {
		log.Fatalf("Failed to create GCP client: %v", err)
	}

	app := ui.NewApp()
	mainView := views.NewCloudRunView(ctx, client)
	app.SetRoot(mainView, true)

	// Refresh the view periodically
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
				app.QueueUpdateDraw(func() {
					if err := mainView.Refresh(ctx); err != nil {
						log.Printf("Error refreshing view: %v", err)
					}
				})
			}
		}
	}()

	if err := mainView.Refresh(ctx); err != nil {
		log.Printf("Initial refresh failed: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
