package views

import (
	"context"
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/lpmourato/c9s/internal/gcp"
	"github.com/lpmourato/c9s/internal/ui"
)

// CloudRunView represents the Cloud Run services view
type CloudRunView struct {
	*ui.Table
	client *gcp.Client
}

// NewCloudRunView returns a new Cloud Run view
func NewCloudRunView(ctx context.Context, client *gcp.Client) *CloudRunView {
	view := &CloudRunView{
		Table:  ui.NewTable(),
		client: client,
	}

	view.SetSelectable(true, false)
	view.SetTitle(" Cloud Run Services ")
	view.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})

	// Setup key bindings
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlL {
			view.showLogs()
			return nil
		}
		return event
	})

	return view
}

// Refresh updates the view with latest Cloud Run services
func (v *CloudRunView) Refresh(ctx context.Context) error {
	services, err := v.client.ListServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list services: %v", err)
	}

	v.Clear()
	v.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})

	// Add data
	for row, svc := range services {
		summary := gcp.ToSummary(svc)
		v.SetCell(row+1, 0, ui.NewTableCell(summary.Name))
		v.SetCell(row+1, 1, ui.NewTableCell(summary.Region))
		v.SetCell(row+1, 2, ui.NewTableCell(summary.URL))
		v.SetCell(row+1, 3, ui.NewTableCell(summary.Status))
		v.SetCell(row+1, 4, ui.NewTableCell(summary.LastDeployTime))

		traffic := ""
		for i, t := range summary.Traffic {
			if i > 0 {
				traffic += ", "
			}
			traffic += fmt.Sprintf("%s (%d%%)", t.RevisionName, t.Percent)
		}
		v.SetCell(row+1, 5, ui.NewTableCell(traffic))
	}

	return nil
}

func (v *CloudRunView) showLogs() {
	row, _ := v.GetSelection()
	if row == 0 {
		return // Header row
	}

	serviceName := v.GetCell(row, 0).Text
	logView := NewLogView(v.client)
	if err := logView.ShowLogs(context.Background(), serviceName); err != nil {
		// Handle error
		return
	}
}
