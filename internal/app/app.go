package app

import (
	"log"

	"github.com/lpmourato/c9s/internal/cli"
	"github.com/lpmourato/c9s/internal/config"
	"github.com/lpmourato/c9s/internal/datasource"
	"github.com/lpmourato/c9s/internal/mock"
	"github.com/lpmourato/c9s/internal/ui"
	"github.com/lpmourato/c9s/internal/views"
)

type App struct {
	cli *cli.CLI
}

func New() *App {
	return &App{
		cli: &cli.CLI{},
	}
}

func (a *App) Run() error {
	ctx, err := a.cli.Parse()
	if err != nil {
		return err
	}

	dsFlag := ctx.Command()

	app := ui.NewApp()
	cfg := &config.CloudRunConfig{
		ProjectID: a.cli.Project,
		Region:    a.cli.Region,
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

	return nil
}
