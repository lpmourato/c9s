package main

import (
	"log"
	"slices"

	"github.com/alecthomas/kong"

	"github.com/lpmourato/c9s/internal/config"
	"github.com/lpmourato/c9s/internal/datasource"
	"github.com/lpmourato/c9s/internal/mock"
	"github.com/lpmourato/c9s/internal/ui"
	"github.com/lpmourato/c9s/internal/views"
)

type CLIStruct struct {
	Datasource string `kong:"help='Data source to use',default='gcp'"`
	Project    string `kong:"help='GCP project ID',env='GOOGLE_CLOUD_PROJECT'"`
	Region     string `kong:"help='Cloud Run region (e.g., us-central1)'"`

	Mock MockCmd `kong:"cmd,help='Run in mock mode'"`
	Gcp  GcpCmd  `kong:"cmd,help='Run normally',default='1'"`
}

func (c *CLIStruct) ValidCommands() []string {
	return []string{"mock", "gcp"}
}

var CLI CLIStruct

type MockCmd struct{}
type GcpCmd struct{}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name("c9s"),
		kong.Description("Cloud Run status UI"),
	)

	dsFlag := ctx.Command()

	if !slices.Contains(CLI.ValidCommands(), dsFlag) {
		ctx.Fatalf("unsupported datasource %q (allowed: mock,gcp)", dsFlag)
	}

	// Validate project required for GCP
	if dsFlag == "gcp" && CLI.Project == "" {
		ctx.Fatalf("project is required for datasource=gcp; set --project or GOOGLE_CLOUD_PROJECT")
	}

	app := ui.NewApp()
	cfg := &config.CloudRunConfig{
		ProjectID: CLI.Project,
		Region:    CLI.Region,
	}
	dsConfig := &datasource.Config{
		ProjectID:  cfg.ProjectID,
		Region:     cfg.Region,
		MockedData: mock.GetDefaultServices(),
	}

	// Map flag to datasource.Type
	switch dsFlag {
	case "mock":
		dsConfig.Type = datasource.Mock
	default:
		dsConfig.Type = datasource.GCP
	}

	ds, err := datasource.Factory(dsConfig)
	if err != nil {
		log.Fatalf("Error creating data source: %v", err)
	}

	views.NewCloudRunView(app, cfg, ds)

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
