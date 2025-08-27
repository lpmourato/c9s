package views

import (
	"context"
	"fmt"
	"strings"
	"time"

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
			v.ScrollToBeginning()
		})
	}()
}

func (v *DeploymentView) displayDetails(details *model.ServiceDetails) {
	// Service Information
	v.writeSectionHeader("Service Information")
	v.writeKeyValue("Name", details.Name)
	v.writeKeyValue("Region", details.Region)
	v.writeKeyValue("URL", details.URL)
	v.writeKeyValue("Status", getStatusText(details.Ready))
	v.writeKeyValue("Creation Time", formatTime(details.CreationTime))
	v.writeKeyValue("Last Updated", formatTime(details.LastUpdated))
	if details.Creator != "" {
		v.writeKeyValue("Creator", details.Creator)
	}
	if details.LastModifier != "" {
		v.writeKeyValue("Last Modifier", details.LastModifier)
	}
	v.writeLine("")

	// Service Metadata
	if details.UID != "" || details.Generation > 0 || len(details.Labels) > 0 {
		v.writeSectionHeader("Service Metadata")
		if details.UID != "" {
			v.writeKeyValue("UID", details.UID)
		}
		if details.Generation > 0 {
			v.writeKeyValue("Generation", fmt.Sprintf("%d", details.Generation))
		}
		if details.LaunchStage != "" {
			v.writeKeyValue("Launch Stage", details.LaunchStage)
		}

		// Display important labels
		if len(details.Labels) > 0 {
			v.writeKeyValue("Labels", "")
			for k, val := range details.Labels {
				v.writeIndentedLine(1, "[dim::]%s: [silver::]%s", k, val)
			}
		}
		v.writeLine("")
	}

	// Container Configuration
	v.writeSectionHeader("Container Configuration")
	v.writeKeyValue("Image", details.ContainerImage)
	if details.ImageDigest != "" {
		v.writeKeyValue("Image Digest", truncateDigest(details.ImageDigest))
	}
	if details.ContainerName != "" {
		v.writeKeyValue("Container Name", details.ContainerName)
	}
	v.writeKeyValue("CPU", details.CPU)
	v.writeKeyValue("Memory", details.Memory)
	v.writeKeyValue("Port", fmt.Sprintf("%d", details.Port))
	if details.ContainerConcurrency > 0 {
		v.writeKeyValue("Concurrency", fmt.Sprintf("%d", details.ContainerConcurrency))
	}
	if details.TimeoutSeconds > 0 {
		v.writeKeyValue("Timeout", fmt.Sprintf("%ds", details.TimeoutSeconds))
	}
	v.writeLine("")

	// Scaling Configuration
	v.writeSectionHeader("Scaling Configuration")
	v.writeKeyValue("Minimum Instances", fmt.Sprintf("%d", details.MinInstances))
	v.writeKeyValue("Maximum Instances", fmt.Sprintf("%d", details.MaxInstances))
	v.writeLine("")

	// Network & Security
	if details.ServiceAccount != "" || details.VPCConnector != "" || details.IngressSettings != "" {
		v.writeSectionHeader("Network & Security")
		if details.ServiceAccount != "" {
			v.writeKeyValue("Service Account", details.ServiceAccount)
		}
		if details.VPCConnector != "" {
			v.writeKeyValue("VPC Connector", details.VPCConnector)
		}
		if details.VPCEgress != "" {
			v.writeKeyValue("VPC Egress", details.VPCEgress)
		}
		if details.IngressSettings != "" {
			v.writeKeyValue("Ingress", details.IngressSettings)
		}
		if details.ExecutionEnv != "" {
			v.writeKeyValue("Execution Environment", details.ExecutionEnv)
		}
		v.writeKeyValue("CPU Throttling", getBoolText(details.CPUThrottling))
		v.writeLine("")
	}

	// Health Checks
	if details.LivenessProbe != nil || details.ReadinessProbe != nil || details.StartupProbe != nil {
		v.writeSectionHeader("Health Checks")
		if details.LivenessProbe != nil {
			v.displayHealthProbe("Liveness Probe", details.LivenessProbe)
		}
		if details.ReadinessProbe != nil {
			v.displayHealthProbe("Readiness Probe", details.ReadinessProbe)
		}
		if details.StartupProbe != nil {
			v.displayHealthProbe("Startup Probe", details.StartupProbe)
		}
		v.writeLine("")
	}

	// Traffic Configuration
	if len(details.Traffic) > 0 {
		v.writeSectionHeader("Traffic Configuration")
		for _, t := range details.Traffic {
			status := ""
			if t.Latest {
				status = " [green::](latest)"
			}
			if t.Tag != "" {
				status += fmt.Sprintf(" [blue::](%s)", t.Tag)
			}
			v.writeKeyValue("Revision "+t.RevisionName, fmt.Sprintf("%d%%%s", t.Percent, status))
		}
		v.writeLine("")
	}

	// Revision Information
	if details.LatestRevision != "" || details.LatestReadyRevision != "" {
		v.writeSectionHeader("Revision Information")
		if details.LatestRevision != "" {
			v.writeKeyValue("Latest Revision", details.LatestRevision)
		}
		if details.LatestReadyRevision != "" {
			v.writeKeyValue("Latest Ready Revision", details.LatestReadyRevision)
		}
		if !details.RevisionCreationTime.IsZero() {
			v.writeKeyValue("Revision Created", formatTime(details.RevisionCreationTime))
		}

		// Container statuses
		if len(details.ContainerStatuses) > 0 {
			v.writeKeyValue("Container Statuses", "")
			for _, cs := range details.ContainerStatuses {
				status := "[red]Not Ready"
				if cs.Ready {
					status = "[green::]Ready"
				}
				v.writeIndentedLine(1, "[dim::]%s: %s (restarts: %d)", cs.Name, status, cs.RestartCount)
			}
		}
		v.writeLine("")
	}

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

	// Secrets & Volumes
	if len(details.Secrets) > 0 || len(details.Volumes) > 0 {
		v.writeSectionHeader("Secrets & Volumes")

		if len(details.Secrets) > 0 {
			v.writeKeyValue("Secrets", "")
			for _, secret := range details.Secrets {
				v.writeIndentedLine(1, "	[dim::]%s[silver::]%s", secret.Name, secret.MountPath)
				// TODO: no need for extra loop for instance
				// for _, item := range secret.Items {
				// 	fmt.Fprintf(v, "%s as [silver::]%s", item.Key, item.Path)
				// 	v.writeIndentedLine(1, "	[dim::]%s:[silver::]%s", item.Key, item.Path)
				// }
			}
		}

		if len(details.Volumes) > 0 {
			v.writeKeyValue("Volumes", "")
			for _, vol := range details.Volumes {
				readOnlyText := ""
				if vol.ReadOnly {
					readOnlyText = " [dim::](read-only)"
				}
				// TODO: check missing info
				v.writeIndentedLine(1, "	[dim::]%s (%s):[silver::]%s%s", vol.Name, vol.VolumeType, vol.MountPath, readOnlyText)
			}
		}
		v.writeLine("")
	}

	// Service Conditions
	if len(details.RevisionConditions) > 0 {
		v.writeSectionHeader("Service Conditions")
		for _, cond := range details.RevisionConditions {
			status := "[red::]" + cond.Status + ""
			if cond.Status == "True" {
				status = "[green::]" + cond.Status + ""
			}
			v.writeKeyValue(cond.Type, fmt.Sprintf("%s (%s)", status, cond.Reason))
			if cond.Message != "" {
				v.writeIndentedLine(2, "[dim::]%s", cond.Message)
			}
		}
		v.writeLine("")
	}

	// Additional Information
	if details.LogURL != "" || details.SelfLink != "" {
		v.writeSectionHeader("Additional Information")
		if details.LogURL != "" {
			v.writeKeyValue("Logs", details.LogURL)
		}
		if details.SelfLink != "" {
			v.writeKeyValue("Self Link", details.SelfLink)
		}
		if details.OperationID != "" {
			v.writeKeyValue("Operation ID", details.OperationID)
		}
		v.writeLine("")
	}
}

func (v *DeploymentView) writeSectionHeader(title string) {
	fmt.Fprintf(v, "[orange::b]%s[-:-:-]\n", title)
}

func (v *DeploymentView) writeKeyValue(key, value string) {
	fmt.Fprintf(v, "	[teal::]%s: [silver::]%s\n", key, value)
}

func (v *DeploymentView) writeLine(text string) {
	fmt.Fprintf(v, "%s\n", text)
}

func (v *DeploymentView) writeIndentedLine(indent int, format string, args ...interface{}) {
	indentStr := strings.Repeat("  ", indent)
	fmt.Fprintf(v, indentStr+format+"\n", args...)
}

func (v *DeploymentView) displayHealthProbe(name string, probe *model.HealthProbe) {
	if probe.HTTPGet != nil {
		path := probe.HTTPGet.Path
		if path == "" {
			path = "/"
		}
		v.writeKeyValue(name, fmt.Sprintf("HTTP %s:%d%s (delay: %ds, period: %ds)",
			probe.HTTPGet.Scheme, probe.HTTPGet.Port, path,
			probe.InitialDelaySeconds, probe.PeriodSeconds))
	}
}

func getStatusText(ready bool) string {
	if ready {
		return "[green::]Ready"
	}
	return "[red::]Not Ready"
}

func getBoolText(value bool) string {
	if value {
		return "[green::]Enabled"
	}
	return "[red::]Disabled"
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04:05")
}

func truncateDigest(digest string) string {
	if len(digest) > 64 {
		return digest[:64] + "..."
	}
	return digest
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
