package sensu

import (
	"fmt"
	"testing"

	"github.com/sharenowTech/virity/internal/config"
	"github.com/sharenowTech/virity/internal/pluginregistry"
)

func TestSendToTCP(t *testing.T) {
	cfg := config.GetMonitorConfig()
	tcp := sendTCP(cfg.Endpoint)

	err := tcp([]byte("{\"name\":\"localhost-/test\",\"status\":0,\"container_id\":\"\",\"hostname\":\"localhost\",\"vulnerability_cve\":[{\"Fix\":\"None\",\"Package\":\"tar-1.29b-1.1\",\"Severity\":3,\"URL\":\"https://security-tracker.debian.org/tracker/CVE-2005-2541\",\"Vuln\":\"CVE-2005-2541\",\"Description\":\"\"},{\"Fix\":\"0.0.9+deb9u1\",\"Package\":\"sensible-utils-0.0.9\",\"Severity\":1,\"URL\":\"https://security-tracker.debian.org/tracker/CVE-2017-17512\",\"Vuln\":\"CVE-2017-17512\",\"Description\":\"\"},{\"Fix\":\"None\",\"Package\":\"util-linux-2.29.2-1\",\"Severity\":0,\"URL\":\"https://security-tracker.debian.org/tracker/CVE-2016-2779\",\"Vuln\":\"CVE-2016-2779\",\"Description\":\"\"},{\"Fix\":\"None\",\"Package\":\"apt-1.4.8\",\"Severity\":2,\"URL\":\"https://security-tracker.debian.org/tracker/CVE-2011-3374\",\"Vuln\":\"CVE-2011-3374\",\"Description\":\"\"}],\"image_name\":\"debian:latest\",\"image_id\":\"da653cee0545dfbe3c1864ab3ce782805603356a9cc712acc7b3100d9932fa5e\",\"output\":\"anchore found 4 vulnerabilities.\"}"))
	if err != nil {
		t.Error(err)
		return
	}
}

func TestToJSON(t *testing.T) {
	vuln := pluginregistry.Vulnerabilities{
		Scanner: "anchore",
		CVE: []pluginregistry.CVE{
			pluginregistry.CVE{
				Fix:      "None",
				Package:  "tar-1.29b-1.1",
				Severity: pluginregistry.SeverityNegligible,
				URL:      "https://security-tracker.debian.org/tracker/CVE-2005-2541",
				Vuln:     "CVE-2005-2541",
			},
			pluginregistry.CVE{
				Fix:      "0.0.9+deb9u1",
				Package:  "sensible-utils-0.0.9",
				Severity: pluginregistry.SeverityMedium,
				URL:      "https://security-tracker.debian.org/tracker/CVE-2017-17512",
				Vuln:     "CVE-2017-17512",
			}, pluginregistry.CVE{
				Fix:      "None",
				Package:  "util-linux-2.29.2-1",
				Severity: pluginregistry.SeverityHigh,
				URL:      "https://security-tracker.debian.org/tracker/CVE-2016-2779",
				Vuln:     "CVE-2016-2779",
			}, pluginregistry.CVE{
				Fix:      "None",
				Package:  "apt-1.4.8",
				Severity: pluginregistry.SeverityLow,
				URL:      "https://security-tracker.debian.org/tracker/CVE-2011-3374",
				Vuln:     "CVE-2011-3374",
			},
		},
	}

	check := sensuCheck{
		Name:        fmt.Sprintf("Testname"),
		Output:      fmt.Sprintf("%v found %v vulnerabilities.", vuln.Scanner, len(vuln.CVE)),
		CVEs:        vuln.CVE,
		ImageID:     fmt.Sprintf("ImageID"),
		ImageDigest: fmt.Sprintf("ImageDigest"),
		ImageTag:    fmt.Sprintf("ImageTag"),
		Status:      0,
	}

	flat, err := toFlatMap(check)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = toJSON(flat)
	if err != nil {
		t.Error(err)
		return
	}
}
