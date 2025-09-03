package cli

import (
	"slices"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Datasource string `kong:"help='Data source to use',default='gcp'"`
	Project    string `kong:"help='GCP project ID',env='GOOGLE_CLOUD_PROJECT'"`
	Region     string `kong:"help='Cloud Run region (e.g., us-central1)'"`

	Mock MockCmd `kong:"cmd,help='Run in mock mode'"`
	Gcp  GcpCmd  `kong:"cmd,help='Run normally',default='1'"`
}

func (c *CLI) ValidCommands() []string {
	return []string{"mock", "gcp"}
}

func (c *CLI) Parse() (*kong.Context, error) {
	ctx := kong.Parse(c,
		kong.Name("c9s"),
		kong.Description("Cloud Run status UI"),
	)

	dsFlag := ctx.Command()

	if !slices.Contains(c.ValidCommands(), dsFlag) {
		ctx.Fatalf("unsupported datasource %q (allowed: mock,gcp)", dsFlag)
	}

	// Validate project required for GCP
	if dsFlag == "gcp" && c.Project == "" {
		ctx.Fatalf("project is required for datasource=gcp; set --project or GOOGLE_CLOUD_PROJECT")
	}

	return ctx, nil
}

type MockCmd struct{}
type GcpCmd struct{}
