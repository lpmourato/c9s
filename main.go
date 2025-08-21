package main

import (
	"flag"
	"log"

	"github.com/lpmourato/c9s/internal/config"
	"github.com/lpmourato/c9s/internal/datasource"
	"github.com/lpmourato/c9s/internal/model"
	"github.com/lpmourato/c9s/internal/ui"
	"github.com/lpmourato/c9s/internal/views"
)

func main() {
	// Parse command line flags
	testMode := flag.Bool("test", false, "Run in test mode with mock data")
	jsonFile := flag.String("json", "", "Path to JSON file with service data")
	mysqlConn := flag.String("mysql", "", "MySQL connection string")
	projectID := flag.String("project", "", "GCP project ID")
	region := flag.String("region", "", "Cloud Run region (e.g., us-central1)")
	flag.Parse()

	app := ui.NewApp()

	// Create config with default settings
	cfg := &config.CloudRunConfig{
		ProjectID: *projectID, // Set from command line
		Region:    *region,    // Set from command line
	}

	// Create data source config based on flags
	var dsType datasource.Type
	dsConfig := &datasource.Config{
		ProjectID:  cfg.ProjectID,
		Region:     cfg.Region,
		MockedData: model.GetDefaultMockData(),
	}

	switch {
	case *testMode:
		dsType = datasource.Mock
	case *jsonFile != "":
		dsType = datasource.JSON
		dsConfig.JSONPath = *jsonFile
	case *mysqlConn != "":
		dsType = datasource.MySQL
		dsConfig.MySQLConn = *mysqlConn
	default:
		dsType = datasource.GCP
		if cfg.ProjectID == "" {
			log.Fatal("Project ID is required. Use --project flag or set GOOGLE_CLOUD_PROJECT environment variable")
		}
	}
	dsConfig.Type = dsType

	// Create data source
	ds, err := datasource.Factory(dsConfig)
	if err != nil {
		log.Fatalf("Error creating data source: %v", err)
	}

	views.NewCloudRunView(app, cfg, ds)

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
