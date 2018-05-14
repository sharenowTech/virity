package image

import "github.com/car2go/virity/internal/pluginregistry"

func evalStatus(vuln []pluginregistry.CVE, severityLevel pluginregistry.VulnSeverity) pluginregistry.MonitorStatus {
	status := pluginregistry.StatusOK
	for _, cve := range vuln {
		if cve.Severity <= severityLevel {
			status = pluginregistry.StatusError
		} /*else if cve.Severity == severityLevel+1 && status != pluginregistry.StatusError {
			status = pluginregistry.StatusWarning
		}*/
	}
	return status
}

func filterCVEs(severityLevel pluginregistry.VulnSeverity, vuln []pluginregistry.CVE) []pluginregistry.CVE {
	filtered := make([]pluginregistry.CVE, 0)
	for _, cve := range vuln {
		if cve.Severity <= severityLevel {
			filtered = append(filtered, cve)
		}
	}
	return filtered
}
