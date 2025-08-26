package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/lpmourato/c9s/internal/interfaces"
	"github.com/lpmourato/c9s/internal/model"
)

// DeploymentView shows Cloud Run service deployment details
type DeploymentView struct {
	*tview.TextView
	app         interfaces.UIController
	serviceName string
	region      string
}

// NewDeploymentView creates a new deployment details view
func NewDeploymentView(app interfaces.UIController, serviceName, region string) *DeploymentView {
	v := &DeploymentView{
		TextView:    tview.NewTextView().SetDynamicColors(true).SetScrollable(true),
		app:         app,
		serviceName: serviceName,
		region:      region,
	}

	v.SetBorder(true)
	v.SetTitle(fmt.Sprintf(" %s - Deployment Details ", serviceName))
	v.SetTitleAlign(tview.AlignLeft)

	// Set up key bindings
	v.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			app.ReturnToMain()
			return nil
		case tcell.KeyUp:
			row, _ := v.GetScrollOffset()
			v.ScrollTo(row-1, 0)
			return nil
		case tcell.KeyDown:
			row, _ := v.GetScrollOffset()
			v.ScrollTo(row+1, 0)
			return nil
		}
		return event
	})

	// Show loading message
	fmt.Fprintf(v, "[yellow::b]Loading deployment details for %s...\n", serviceName)

	return v
}

// LoadDetails loads and displays the service details
func (v *DeploymentView) LoadDetails(provider model.CloudProvider) {
	go func() {
		details, err := provider.GetServiceDetails(context.Background(), v.serviceName, v.region)
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.Clear()
				fmt.Fprintf(v, "[red]Error loading deployment details: %v\n", err)
			})
			return
		}

		v.app.QueueUpdateDraw(func() {
			v.Clear()
			v.displayDetails(details)
		})
	}()
}

func (v *DeploymentView) displayDetails(details *model.ServiceDetails) {
	// Service Information
	v.writeSectionHeader("Service Information")
	v.writeKeyValue("Name", details.Name)
	v.writeKeyValue("Region", details.Region)
	v.writeKeyValue("URL", details.URL)
	v.writeKeyValue("Last Updated", details.LastUpdated.Format("2006-01-02 15:04:05"))
	v.writeKeyValue("Status", getStatusText(details.Ready))
	v.writeLine("")

	// Container Configuration
	v.writeSectionHeader("Container Configuration")
	v.writeKeyValue("Image", details.ContainerImage)
	v.writeKeyValue("CPU", details.CPU)
	v.writeKeyValue("Memory", details.Memory)
	v.writeKeyValue("Port", fmt.Sprintf("%d", details.Port))
	v.writeLine("")

	// Traffic Information
	v.writeSectionHeader("Traffic Configuration")
	for _, t := range details.Traffic {
		status := ""
		if t.Latest {
			status = " [green::](latest)[-]"
		}
		if t.Tag != "" {
			status += fmt.Sprintf(" [blue::](%s)[-]", t.Tag)
		}
		v.writeKeyValue("Revision "+t.RevisionName, fmt.Sprintf("%d%%%s", t.Percent, status))
	}
	v.writeLine("")

	// Environment Variables
	if len(details.EnvVars) > 0 {
		v.writeSectionHeader("Environment Variables")
		for k, val := range details.EnvVars {
			// Mask sensitive values
			if isSensitive(k) {
				val = "********"
			}
			v.writeKeyValue(k, val)
		}
		v.writeLine("")
	}

	// Scaling Configuration
	v.writeSectionHeader("Scaling Configuration")
	v.writeKeyValue("Minimum Instances", fmt.Sprintf("%d", details.MinInstances))
	v.writeKeyValue("Maximum Instances", fmt.Sprintf("%d", details.MaxInstances))
}

func (v *DeploymentView) writeSectionHeader(title string) {
	fmt.Fprintf(v, "[orange::b]%s[-:-:-]\n", title)
}

func (v *DeploymentView) writeKeyValue(key, value string) {
	fmt.Fprintf(v, "[teal::]%s[-]: [silver::]%s\n", key, value)
}

func (v *DeploymentView) writeLine(text string) {
	fmt.Fprintf(v, "%s\n", text)
}

func getStatusText(ready bool) string {
	if ready {
		return "[green::]Ready[-]"
	}
	return "[red::]Not Ready[-]"
}

func isSensitive(key string) bool {
	key = strings.ToUpper(key)
	sensitiveKeys := []string{
		"PASSWORD", "SECRET", "KEY", "TOKEN", "CREDENTIAL",
		"AUTH", "PRIVATE", "CERT", "SIGNATURE",
	}

	for _, k := range sensitiveKeys {
		if strings.Contains(key, k) {
			return true
		}
	}
	return false
}
