package sensu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sharenowTech/virity/internal/log"
	"github.com/sharenowTech/virity/internal/pluginregistry"
)

const (
	ok       = 0
	warning  = 1
	critical = 2
)

const prefix = "VIRITY"

type sensu struct {
	endpoint      string
	httpUser      string
	httpPassword  string
	defaultOwner  string
	createTickets bool
}

const name = "name"

type sensuCheck struct {
	Name        string               `json:"name"`
	Status      int                  `json:"status"`
	Containers  string               `json:"in_containers"`
	CVEs        []pluginregistry.CVE `json:"vulnerability_cve"`
	ImageID     string               `json:"image_id"`
	ImageTag    string               `json:"image_tag"`
	ImageDigest string               `json:"image_digest"`
	Output      string               `json:"output"`
}

func init() {
	// register New function at pluginregistry
	_, err := pluginregistry.RegisterMonitor("sensu", New)
	if err != nil {
		log.Info(log.Fields{
			"function": "init",
			"package":  "sensu",
			"error":    err.Error(),
		}, "An error occoured while register a monitor")
	}
}

// New initializes the plugin
func New(config pluginregistry.Config) pluginregistry.Monitor {
	sensu := sensu{
		endpoint:      config.Endpoint,
		httpUser:      config.User,
		httpPassword:  config.Password,
		defaultOwner:  config.DefaultAssignee,
		createTickets: config.CreateTickets,
	}
	if sensu.endpoint == "" {
		sensu.endpoint = "localhost:3030"
	}
	return sensu
}

func (sensu sensu) Push(image pluginregistry.ImageStack, status pluginregistry.MonitorStatus) error {
	//TODO what if no owner available
	if len(image.MetaData.OwnerID) < 1 {
		if sensu.createTickets == true {
			return fmt.Errorf("No owner/assignee was specified for image %v", image.MetaData.Tag)
		}
		image.MetaData.OwnerID = append(image.MetaData.OwnerID, "dummy")
	}

	check, err := createCheck(image, int(status))
	if err != nil {
		return err
	}

	flatMap, err := toFlatMap(check)
	if err != nil {
		return err
	}
	for _, owner := range image.MetaData.OwnerID {
		if sensu.createTickets == true {
			flatMap = addTicketInfo(flatMap, owner, "") // Projectname not needed (owners default project is used)
		}
		json, err := toJSON(flatMap)
		if err != nil {
			return err
		}
		log.Debug(log.Fields{
			"function": "Push",
			"package":  "sensu",
			"data":     string(json)[:100],
		}, "Sending JSON to sensu")
		tcp := sendTCP(sensu.endpoint)
		err = tcp(json)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sensu sensu) Resolve(image pluginregistry.ImageStack) error {
	log.Info(log.Fields{
		"function": "Resolve",
		"package":  "sensu",
		"image":    image.MetaData.Tag,
		"imageID":  image.MetaData.ImageID,
		"owner":    image.MetaData.OwnerID,
	}, "Image is not running anymore and will be resolved")
	return sensu.Push(image, pluginregistry.StatusOK)
}

func sendTCP(endpoint string) func(data []byte) error {
	return func(data []byte) error {
		d := net.Dialer{Timeout: 60 * time.Second}

		conn, err := d.Dial("tcp", endpoint)
		if err != nil {
			return err
		}
		defer conn.Close()

		ret, err := conn.Write(data)
		if err != nil {
			return err
		}
		log.Debug(log.Fields{
			"function": "sendTCP",
			"package":  "sensu",
			"bytes":    ret,
		}, "Sending data to sensu")
		return nil
	}
}

func toJSON(obj interface{}) ([]byte, error) {

	json, err := json.Marshal(obj)

	if err != nil {
		return nil, err
	}
	return json, nil
}

func toFlatMap(check interface{}) (map[string]interface{}, error) {
	//flatMap := make(map[string]interface{})

	reflValue := reflect.ValueOf(check)
	if reflValue.Kind() == reflect.Ptr {
		reflValue = reflValue.Elem()
	}

	if reflValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("an error occoured during reflecting the struct")
	}

	flatMap := make(map[string]interface{})

	//reflType := reflValue.Type()
	for i := 0; i < reflValue.NumField(); i++ {
		field := reflValue.Field(i)
		jsonTag := reflValue.Type().Field(i).Tag.Get("json")
		switch field.Type().Kind() {
		case reflect.Slice, reflect.Array:

			listSlice, ok := field.Interface().([]pluginregistry.CVE)
			if ok {
				log.Debug(log.Fields{
					"function": "toFlatMap",
					"package":  "sensu",
				}, "Slice is a CVE array")
				for _, value := range listSlice {
					counter := 0
					if _, ok := flatMap[value.Vuln]; ok {
						counter++
						flatMap[value.Vuln+"_"+strconv.Itoa(counter)] = fmt.Sprintf("%v  -  Package: %v  -  Available Fix: %v  -  More Info: %v", value.Severity, value.Package, value.Fix, value.URL)
					} else {
						flatMap[value.Vuln] = fmt.Sprintf("%v  -  Package: %v  -  Available Fix: %v  -  More Info: %v", value.Severity.String(), value.Package, value.Fix, value.URL)
					}
				}
			} else {
				for index, value := range listSlice {
					flatMap[strconv.Itoa(index)] = fmt.Sprintf("%+v", value)
				}
			}

		default:
			flatMap[jsonTag] = field.Interface()
		}
	}
	return flatMap, nil
}

func addTicketInfo(check map[string]interface{}, ownerID, projectName string) map[string]interface{} {
	const owner = "team"
	const ticket = "ticket"
	const handlers = "handlers"
	const project = "project"

	check[owner] = ownerID
	if projectName != "" {
		check[project] = projectName
	}
	check[handlers] = []string{"jira"}
	check[ticket] = true
	check[name] = fmt.Sprintf("%v_%v", check[name], ownerID)

	return check
}

func createCheck(img pluginregistry.ImageStack, status int) (sensuCheck, error) {
	//img.container.Name = strings.Replace(img.container.Name, "/", "", -1)
	var strBuffer bytes.Buffer

	for _, container := range img.Containers {
		strBuffer.WriteString(fmt.Sprintf("%v-%v; ", container.Hostname, strings.TrimPrefix(container.Name, "/")))
	}

	replacer := strings.NewReplacer("/", "_", ":", ".")
	checkName := fmt.Sprintf("%v_%v", prefix, replacer.Replace(img.MetaData.Tag))

	// TODO: Manage Output better
	var output string
	switch status {
	case 0:
		if len(img.Vuln.CVE) == 0 {
			output = fmt.Sprintf("%v did not find vulnerabilities for the specified severity.", img.Vuln.Scanner)
		} else {
			output = fmt.Sprintf("image is not running anymore and got resolved.")
		}
	//case 1:
	//	output = fmt.Sprintf("%v found vulnerabilities for a severity lower than specified.", img.Vuln.Scanner)
	case 2:
		output = fmt.Sprintf("%v found %v vulnerabilities.", img.Vuln.Scanner, len(img.Vuln.CVE))
	default:
		return sensuCheck{}, fmt.Errorf("invalid sensu status id")
	}

	return sensuCheck{
		Name:        checkName,
		Containers:  strBuffer.String(),
		Output:      output,
		CVEs:        img.Vuln.CVE,
		ImageTag:    img.MetaData.Tag,
		ImageID:     img.MetaData.ImageID,
		ImageDigest: img.Vuln.Digest,
		Status:      status,
	}, nil
}
