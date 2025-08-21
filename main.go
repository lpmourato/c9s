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
	flag.Parse()

	app := ui.NewApp()

	// Create config with default settings
	cfg := &config.CloudRunConfig{
		ProjectID: "", // Will be set via command
		Region:    "", // Will be set via command
	}

	// Create data source config based on flags
	var dsType datasource.Type
	dsConfig := &datasource.Config{
		ProjectID: cfg.ProjectID,
		Region:    cfg.Region,
	}

	switch {
	case *testMode:
		dsType = datasource.Mock
		dsConfig.MockedData = model.GetDefaultMockData()
	case *jsonFile != "":
		dsType = datasource.JSON
		dsConfig.JSONPath = *jsonFile
	case *mysqlConn != "":
		dsType = datasource.MySQL
		dsConfig.MySQLConn = *mysqlConn
	default:
		dsType = datasource.GCP
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
