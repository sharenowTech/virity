package anchore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	// Main Routes
	images     = "/v1/images"
	imageByID  = "/v1/images/by_id/"
	registries = "/v1/registries"
	status     = "/v1/system/services"

	// JSON Name Mappings for API
	//Error
	errCode    = "code"
	errType    = "error_type"
	errMessage = "message"
	errDetail  = "detail"
)

type api struct {
	username string
	password string
	endpoint string
}

type imageMeta struct {
	AnalysisStatus string `json:"analysis_status"`
	CreatedAt      string `json:"created_at"`
	ImageDigest    string `json:"imageDigest"`
	ImageContent   struct {
		Metadata struct {
			Arch           string `json:"arch"`
			Distro         string `json:"distro"`
			DistroVersion  string `json:"distro_version"`
			DockerfileMode string `json:"dockerfile_mode"`
			ImageSize      int    `json:"image_size"`
			LayerCount     int    `json:"layer_count"`
		} `json:"metadata"`
	} `json:"image_content"`
	ImageDetail []image `json:"image_detail"`
	ImageStatus string  `json:"image_status"`
	ImageType   string  `json:"image_type"`
	LastUpdated string  `json:"last_updated"`
	UserID      string  `json:"userId"`
}

type image struct {
	CreatedAt   string `json:"created_at"`
	Digest      string `json:"digest"`
	Dockerfile  string `json:"dockerfile"`
	Fulldigest  string `json:"fulldigest"`
	Fulltag     string `json:"fulltag"`
	ImageDigest string `json:"imageDigest"`
	ImageID     string `json:"imageId"`
	LastUpdated string `json:"last_updated"`
	Registry    string `json:"registry"`
	Repo        string `json:"repo"`
	Tag         string `json:"tag"`
	UserID      string `json:"userId"`
}

type registry struct {
	Username string `json:"registry_user"`
	Password string `json:"registry_pass"`
	Registry string `json:"registry"`
	RegType  string `json:"registry_type"`
	Insecure bool   `json:"registry_verify"`
}

type imageCVE struct {
	ImageID         string
	ImageDigest     string `json:"imageDigest"`
	Vulnerabilities []struct {
		Fix      string `json:"fix"`
		Package  string `json:"package"`
		Severity string `json:"severity"`
		URL      string `json:"url"`
		Vuln     string `json:"vuln"`
	} `json:"vulnerabilities"`
	VulnerabilityType string `json:"vulnerability_type"`
}

type services []struct {
	BaseURL string `json:"base_url"`
	HostID  string `json:"hostid"`
	// On error service detail returns a string instead a struct. Detail is not necessary
	/*ServiceDetail struct {
		Busy    bool   `json:"busy"`
		Message string `json:"message"`
		Up      bool   `json:"up"`
	} `json:"service_detail"`*/
	Servicename   string `json:"servicename"`
	Status        bool   `json:"status"`
	StatusMessage string `json:"status_message"`
	Version       string `json:"version"`
}

func (api api) newGetRequest(url string) (*http.Request, error) {

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(api.username, api.password)

	return req, nil
}

func (api api) newPostRequest(url string, body interface{}) (*http.Request, error) {
	data := new(bytes.Buffer)
	err := json.NewEncoder(data).Encode(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, data)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(api.username, api.password)

	return req, nil
}

func (api api) Healthcheck() error {
	req, err := api.newGetRequest(api.endpoint + status)
	if err != nil {
		return err
	}
	resp, err := request(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	services := make(services, 6)
	err = json.NewDecoder(resp.Body).Decode(&services)
	if err != nil {
		return fmt.Errorf("Decoding Error: %v - JSON: %v", err.Error(), resp.Body)
	}

	for _, service := range services {
		if service.Status == false {
			return fmt.Errorf("Service %s is %s", service.Servicename, service.StatusMessage)
		}
	}
	return nil
}

func (api api) PushImage(image image) (*image, error) {

	body := make(map[string]interface{})

	if image.Dockerfile != "" {
		body["dockerfile"] = image.Dockerfile
	}

	switch {
	case image.Fulltag != "":
		body["tag"] = image.Fulltag
	case image.Digest != "":
		body["digest"] = image.Digest
	default:
		return nil, fmt.Errorf("No valid image data provided")
	}

	req, err := api.newPostRequest(api.endpoint+images, &body)
	if err != nil {
		return nil, err
	}

	resp, err := request(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	images := make([]imageMeta, 0, 1)
	err = json.NewDecoder(resp.Body).Decode(&images)
	if err != nil {
		return nil, err
	}

	if len(images[0].ImageDetail) < 1 {
		return nil, fmt.Errorf("No vaild detail information found")
	}

	return &images[0].ImageDetail[0], nil
}

func (api api) AddRegistry(reg registry) error {
	reg.Insecure = !reg.Insecure //stored in json variable "verify"

	req, err := api.newPostRequest(api.endpoint+registries, &reg)
	if err != nil {
		return err
	}

	_, err = request(req)
	if err != nil {
		return err
	}

	return nil
}

func (api api) GetCVEs(imageID string) (*imageCVE, error) {
	var cves imageCVE
	req, err := api.newGetRequest(api.endpoint + imageByID + imageID + "/vuln/os")
	if err != nil {
		return nil, err
	}
	resp, err := request(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&cves)
	if err != nil {
		return nil, err
	}

	return &cves, nil
}

func (api api) GetImage(imageID string) (*image, error) {
	images := make([]imageMeta, 0, 1)
	req, err := api.newGetRequest(api.endpoint + imageByID + imageID)
	if err != nil {
		return nil, err
	}
	resp, err := request(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&images)
	if err != nil {
		return nil, err
	}
	if len(images[0].ImageDetail) < 1 {
		return nil, fmt.Errorf("No vaild detail information found")
	}

	return &images[0].ImageDetail[0], nil
}

func request(req *http.Request) (*http.Response, error) {
	client := http.Client{
		Timeout: time.Second * 60,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		dat := make(map[string]interface{})
		jsonErr := json.NewDecoder(resp.Body).Decode(&dat)
		if jsonErr != nil {
			return nil, fmt.Errorf("Failed decoding response - %v - HTTP Code: %v - HTTP Status: %v", jsonErr.Error(), resp.StatusCode, resp.Status)
		}
		return nil, fmt.Errorf("%v - Detail: %v - HTTP Code: %v", dat[errMessage], dat[errDetail], resp.StatusCode)
	}
	return resp, nil
}

func fetch(url string, ch chan<- string) {
	panic("not implemented")
}
