package anchore

import (
	"fmt"
	"time"

	"github.com/sharenowTech/virity/internal/log"
	"github.com/sharenowTech/virity/internal/pluginregistry"
)

var severityMap = map[string]pluginregistry.VulnSeverity{
	"High":       pluginregistry.SeverityHigh,
	"Medium":     pluginregistry.SeverityMedium,
	"Low":        pluginregistry.SeverityLow,
	"Negligible": pluginregistry.SeverityNegligible,
}

type anchore struct {
	api      api
	interval time.Duration //to calculate anchore cve timeout
}

// TODO: replace by interval provided by anchore
const cveWaitTimeout = 2 * time.Minute

func init() {
	// register New function at pluginregistry
	_, err := pluginregistry.RegisterScanner("anchore", New)
	if err != nil {
		log.Info(log.Fields{
			"function": "init",
			"package":  "anchore",
			"error":    err.Error(),
		}, "An error occoured while register a monitor")
	}
}

// New initializes the plugin
func New(config pluginregistry.Config) pluginregistry.Scan {
	anchore := anchore{
		api: api{
			username: config.User,
			password: config.Password,
			endpoint: config.Endpoint,
		},
	}

	// set default values
	if anchore.api.endpoint == "" {
		anchore.api.endpoint = "http://localhost:8228"
	}
	log.Info(log.Fields{
		"function": "New",
		"package":  "anchore",
	}, "Scanner initialized")
	return anchore
}

func (anchore anchore) Scan(scanImage pluginregistry.Image) (*pluginregistry.Vulnerabilities, error) {
	//TODO check for registry
	err := anchore.api.Healthcheck()
	if err != nil {
		return nil, fmt.Errorf("Anchore Healthcheck: %v", err.Error())
	}

	img := &image{
		ImageID: scanImage.ImageID,
		Fulltag: scanImage.Tag,
	}
	log.Debug(log.Fields{
		"function": "Scan",
		"package":  "anchore",
		"image":    img.Fulltag,
		"imageID":  img.ImageID,
	}, "Pushing Image")
	img, err = anchore.api.PushImage(*img)
	if err != nil {
		return nil, err
	}

	log.Debug(log.Fields{
		"function": "Scan",
		"package":  "anchore",
		"image":    img.Fulltag,
		"imageID":  img.ImageID,
	}, "Fetching CVEs for image")
	cves, err := fetchCVE(anchore.api, img.ImageID, cveWaitTimeout)
	if err != nil {
		return nil, err
	}

	return toVulnarabilities(cves), nil
}

func fetchCVE(api api, imageID string, timeout time.Duration) (*imageCVE, error) {
	const retries = 5
	cves, err := api.GetCVEs(imageID)
	if err != nil {
		ticker := time.NewTicker(timeout / retries)
		chanCVE := make(chan *imageCVE)
		go func() {
			for t := range ticker.C {
				cves, err = api.GetCVEs(imageID)
				if err == nil {
					chanCVE <- cves
					return
				}
				log.Debug(log.Fields{
					"function": "Scan",
					"package":  "anchore",
					"error":    err.Error(),
					"imageID":  imageID,
					"retry_at": t.Add(timeout / retries).Format("2006-01-02 15:04:05"),
				}, "No CVE received - I will retry")
			}
		}()
		select {
		case cves := <-chanCVE:
			ticker.Stop()
			cves.ImageID = imageID
			return cves, nil
		case <-time.After(timeout):
			ticker.Stop()
			time.Sleep(timeout / retries) // wait for last tick
			return nil, fmt.Errorf("Analysis failed for Image: %s (Timeout while getting CVEs)", imageID)
		}
	}
	cves.ImageID = imageID
	return cves, nil
}

func toVulnarabilities(cves *imageCVE) *pluginregistry.Vulnerabilities {
	data := pluginregistry.Vulnerabilities{
		Digest:  cves.ImageDigest,
		Scanner: "anchore",
	}
	for _, cve := range cves.Vulnerabilities {
		if cve.Fix == "None" {
			cve.Fix = ""
		}
		severity := severityMap[cve.Severity]
		data.CVE = append(data.CVE, pluginregistry.CVE{
			Fix:      cve.Fix,
			URL:      cve.URL,
			Vuln:     cve.Vuln,
			Severity: severity,
			Package:  cve.Package,
		})
	}

	return &data
}
