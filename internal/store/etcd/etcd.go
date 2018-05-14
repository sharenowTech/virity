package etcd

import (
	"context"
	"encoding/base64"
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/car2go/virity/internal/log"
	"github.com/car2go/virity/internal/pluginregistry"
	"github.com/coreos/etcd/client"
)

type etcd struct {
	endpoint string
}

/*
Structure:
	Backup
		Monitored
			<imageID>
			...
	Agents
		<agentID>
			lifetime
			stacks
				<stackID>
					name
					hostname
					...
*/

const (
	stacksPath = "Stacks"
	agentsPath = "Agents"
	lifetime   = "Lifetime"

	name      = "Name"
	hostname  = "Hostname"
	image     = "Image"
	imageID   = "ImageID"
	timestamp = "Timestamp"
	ownerID   = "OwnerID"
)

func init() {
	// register New function at pluginregistry
	_, err := pluginregistry.RegisterStore("etcd", New)
	if err != nil {
		log.Warn(log.Fields{
			"package":  "etcd",
			"function": "init",
		}, err.Error())
	}
}

// New initializes the plugin
func New(config pluginregistry.Config) pluginregistry.Store {
	etcd := etcd{
		endpoint: config.Endpoint,
	}
	// set default values
	if etcd.endpoint == "" {
		etcd.endpoint = "localhost:2379"
	}
	return &etcd
}

// Store pushes data based on a key to the etcd using a etcd endpoint
// it will be stored at a specified agent (agentID/prefix/)
func (etcd etcd) StoreContainer(container pluginregistry.Container, agent pluginregistry.Agent) error {
	agentID := encodeAgentID(agent, agentsPath)

	//store agent data
	if err := etcd.store(agent.Lifetime.String(), path.Join(agentID, lifetime)); err != nil {
		return err
	}

	stackPath := path.Join(agentID, stacksPath, strconv.FormatInt(container.Timestamp.Unix(), 10))
	return etcd.store(container, path.Join(stackPath, container.ID))
}

func (etcd etcd) StoreImageStack(stack pluginregistry.ImageStack, prefix string) error {
	return etcd.store(stack, path.Join(prefix, stack.MetaData.ImageID))
}

func (etcd etcd) store(obj interface{}, prefix string) error {
	api, err := etcd.connect()
	if err != nil {
		return err
	}

	// Push data
	node, err := createNode(reflect.ValueOf(obj), prefix)
	err = pushNode(node, api)
	if err != nil {
		return err
	}
	return nil
}

func (etcd etcd) LoadImageStacks(prefix string) ([]pluginregistry.ImageStack, error) {
	api, err := etcd.connect()
	if err != nil {
		return nil, err
	}

	images := make([]pluginregistry.ImageStack, 0)

	get := getData(api)

	resp, err := get(prefix)
	if err != nil {
		return nil, err
	}

	for _, image := range resp.Node.Nodes {
		img := pluginregistry.ImageStack{}
		for _, value := range image.Nodes {
			switch path.Base(value.Key) {
			case "MetaData":
				image, err := parseImage(value)
				if err != nil {
					return nil, err
				}
				img.MetaData = image
			case "Vuln":
				vuln, err := parseVuln(value)
				if err != nil {
					return nil, err
				}
				img.Vuln = vuln
			case "Containers":
				//img.Containers = value.Value
				container := make([]pluginregistry.Container, len(value.Nodes))
				for index, cont := range value.Nodes {
					parsed, err := parseContainer(cont)
					if err != nil {
						return nil, err
					}
					container[index] = parsed
				}
				img.Containers = container

			default:
				log.Debug(log.Fields{
					"package":  "etcd",
					"function": "LoadImageStack",
					"NodeKey":  value.Key,
				}, "key can not be handled and will be dropped")
			}
		}
		images = append(images, img)
	}

	return images, nil
}

func parseImage(base *client.Node) (pluginregistry.Image, error) {
	img := pluginregistry.Image{}
	for _, child := range base.Nodes {
		switch path.Base(child.Key) {
		case "Tag":
			img.Tag = child.Value
		case "ImageID":
			img.ImageID = child.Value
		case "StackID":
			id, err := strconv.ParseInt(child.Value, 10, 64)
			if err != nil {
				return img, err
			}
			img.StackID = id

		case "OwnerID":
			owners := make([]string, len(child.Nodes))
			for index, key := range child.Nodes {
				owners[index] = key.Value
			}
			img.OwnerID = owners
		default:
			log.Debug(log.Fields{
				"package":  "etcd",
				"function": "parseImage",
				"NodeKey":  child.Key,
			}, "key can not be handled and will be dropped")
		}
	}
	return img, nil
}

func parseVuln(base *client.Node) (pluginregistry.Vulnerabilities, error) {
	vuln := pluginregistry.Vulnerabilities{}
	for _, child := range base.Nodes {
		switch path.Base(child.Key) {
		case "Digest":
			vuln.Digest = child.Value
		case "Scanner":
			vuln.Scanner = child.Value
		case "CVE":
			cves := make([]pluginregistry.CVE, len(child.Nodes))
			for index, cve := range child.Nodes {
				parsed, err := parseCVE(cve)
				if err != nil {
					return vuln, err
				}
				cves[index] = parsed
			}
			vuln.CVE = cves
		default:
			log.Debug(log.Fields{
				"package":  "etcd",
				"function": "parseVuln",
				"NodeKey":  child.Key,
			}, "key can not be handled and will be dropped")
		}
	}
	return vuln, nil
}

func parseCVE(base *client.Node) (pluginregistry.CVE, error) {
	cve := pluginregistry.CVE{}
	for _, val := range base.Nodes {
		switch path.Base(val.Key) {
		case "Fix":
			cve.Fix = val.Value
		case "Package":
			cve.Package = val.Value
		case "Severity":
			sev, err := strconv.Atoi(val.Value)
			if err != nil {
				return cve, err
			}
			cve.Severity = pluginregistry.VulnSeverity(sev)
		case "URL":
			cve.URL = val.Value
		case "Vuln":
			cve.Vuln = val.Value
		case "Description":
			cve.Description = val.Value
		default:
			log.Debug(log.Fields{
				"package":  "etcd",
				"function": "parseCVE",
				"NodeKey":  val.Key,
			}, "key can not be handled and will be dropped")
		}
	}
	return cve, nil
}

func parseContainer(base *client.Node) (pluginregistry.Container, error) {
	cont := pluginregistry.Container{}
	for _, container := range base.Nodes {
		switch path.Base(container.Key) {
		case "ImageID":
			cont.ImageID = container.Value
		case "OwnerID":
			cont.OwnerID = container.Value
		case "Timestamp":
			t, err := strconv.ParseInt(container.Value, 10, 64)
			if err != nil {
				return cont, err
			}
			unixT := time.Unix(t, 0).UTC()
			cont.Timestamp = unixT
		case "Name":
			cont.Name = container.Value
		case "Hostname":
			cont.Hostname = container.Value
		case "Image":
			cont.Image = container.Value
		case "ID":
			cont.ID = container.Value

		default:
			log.Debug(log.Fields{
				"package":  "etcd",
				"function": "parseContainer",
				"NodeKey":  container.Key,
			}, "key can not be handled and will be dropped")
		}
	}

	return cont, nil
}

// LoadContainerGroup loads a complete complete container group from etcd server
func (etcd etcd) LoadContainerGroup(agent pluginregistry.Agent) (*pluginregistry.ContainerGroup, error) {
	api, err := etcd.connect()
	if err != nil {
		return nil, err
	}

	agentID := encodeAgentID(agent, agentsPath)

	latestStack, err := latestStackKey(api, agentID)
	if err != nil {
		return nil, err
	}

	get := getData(api)
	stackPath := path.Join(agentID, stacksPath, strconv.FormatInt(latestStack, 10))
	resp, err := get(stackPath)
	if err != nil {
		return nil, err
	}

	unixTime, err := strconv.ParseInt(path.Base(resp.Node.Key), 10, 64)
	if err != nil {
		return nil, err
	}

	stack := &pluginregistry.ContainerGroup{
		ID:   unixTime,
		Date: time.Unix(unixTime, 0).UTC(), // only for debugging
	}

	for _, values := range resp.Node.Nodes {
		container, err := parseContainer(values)
		if err != nil {
			return nil, err
		}
		stack.Container = append(stack.Container, container)
	}

	return stack, nil
}

func (etcd etcd) CheckID(agent pluginregistry.Agent) (bool, error) {
	agentID := encodeAgentID(agent, agentsPath)
	api, err := etcd.connect()
	if err != nil {
		return false, err
	}
	keys, err := listKeys(api, "/")
	if err != nil {
		return false, err
	}
	for _, key := range keys {
		if key == agentID {
			return false, nil
		}
	}
	return true, nil
}

func (etcd etcd) Maintain(agent pluginregistry.Agent, latestKey int64) error {
	agentID := encodeAgentID(agent, agentsPath)
	api, err := etcd.connect()
	if err != nil {
		return err
	}
	keys, err := listKeys(api, path.Join(agentID, stacksPath))
	if err != nil {
		return err
	}

	del := delData(api)
	for _, key := range keys {
		keyID, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return err
		}
		if keyID != latestKey {
			err := del(path.Join(agentID, stacksPath, key))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (etcd etcd) Delete(key string) error {
	api, err := etcd.connect()
	if err != nil {
		return err
	}

	del := delData(api)
	err = del(key)
	if err != nil {
		return err
	}
	return nil
}

func (etcd etcd) DeleteAgent(agent pluginregistry.Agent) error {
	agentID := encodeAgentID(agent, agentsPath)
	return etcd.Delete(agentID)
}

func (etcd etcd) LoadAgents() ([]pluginregistry.Agent, error) {
	api, err := etcd.connect()
	if err != nil {
		return nil, err
	}

	keys, err := listKeys(api, agentsPath)
	if err != nil {
		return nil, err
	}

	agentList := make([]pluginregistry.Agent, 0, len(keys))
	get := getData(api)
	for _, key := range keys {
		agent, err := decodeAgentID(key)
		if err != nil {
			log.Info(log.Fields{
				"package":  "etcd",
				"function": "LoadAgents",
				"ID":       key,
			}, "AgentID could not be decoded and will be ommitted")
			continue
		}

		// not yet implemented
		/*if agent.Version != version {
			log.Debugf("Agent %v has the wrong version and will be omitted", key)
			continue
		}*/
		resp, err := get(path.Join(agentsPath, key, lifetime))
		if err != nil {
			return nil, err
		}
		life, err := time.ParseDuration(resp.Node.Value)
		if err != nil {
			return nil, err
		}
		agent.Lifetime = life
		agentList = append(agentList, agent)
	}
	return agentList, nil
}

func listKeys(api client.KeysAPI, location string) ([]string, error) {
	get := getData(api)

	resp, err := get(location)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0)
	for _, val := range resp.Node.Nodes {
		trimmed := path.Base(val.Key)
		keys = append(keys, trimmed)
	}
	return keys, nil
}

func latestStackKey(api client.KeysAPI, agentID string) (int64, error) {
	keys, err := listKeys(api, path.Join("/", agentID, stacksPath))
	if err != nil {
		return 0, err
	}
	maxKey := int64(0)

	for _, key := range keys {
		intKey, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return 0, err
		}
		if maxKey < intKey {
			maxKey = intKey
		}
	}

	return maxKey, nil
}

// Connect creates a connection to the KeysAPI of
func (etcd etcd) connect() (client.KeysAPI, error) {
	cfg := client.Config{
		Endpoints: []string{etcd.endpoint},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: 10 * time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	kapi := client.NewKeysAPI(c)
	return kapi, nil
}

func pushNode(node *client.Node, api client.KeysAPI) error {
	set := setData(api)

	if len(node.Nodes) == 0 {
		err := set(path.Join(node.Key), node.Value)
		if err != nil {
			return err
		}
		return nil
	}

	for _, value := range node.Nodes {
		err := pushNode(value, api)
		if err != nil {
			return err
		}
	}
	return nil
}

func setData(kapi client.KeysAPI) func(string, string) error {
	return func(nodeName, value string) error {
		_, err := kapi.Set(context.Background(), nodeName, value, nil)
		return err
	}
}

func delData(kapi client.KeysAPI) func(string) error {
	return func(nodeName string) error {
		_, err := kapi.Delete(context.Background(), nodeName, &client.DeleteOptions{Recursive: true})
		return err
	}
}

func getData(kapi client.KeysAPI) func(string) (*client.Response, error) {
	return func(nodeName string) (*client.Response, error) {
		resp, err := kapi.Get(context.Background(), nodeName, &client.GetOptions{Recursive: true})
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

// BucketBaseKey calculates a key used to store data based on the time
func bucketBaseKey(date time.Time, duration time.Duration) string {
	unixDate := date.Unix()
	key := unixDate - (unixDate % int64(duration.Seconds()))
	return strconv.FormatInt(key, 10)
}

func encodeAgentID(agent pluginregistry.Agent, prefix string) string {
	id := fmt.Sprintf("%v&%v&%v&%v", agent.Hostname, agent.DockerHostID, agent.Version, agent.RNG)
	return path.Join(prefix, base64.URLEncoding.EncodeToString([]byte(id)))
}

func decodeAgentID(agentID string) (pluginregistry.Agent, error) {
	dec, err := base64.URLEncoding.DecodeString(agentID)
	if err != nil {
		return pluginregistry.Agent{}, err
	}
	splitted := strings.Split(string(dec), "&")
	if len(splitted) != 4 {
		return pluginregistry.Agent{}, fmt.Errorf("ID does not match an agent ID")
	}
	return pluginregistry.Agent{
		Hostname:     splitted[0],
		DockerHostID: splitted[1],
		Version:      splitted[2],
		RNG:          splitted[3],
	}, nil
}

// CreateNode from a reflectable object
// Prefix defines the root key
func createNode(field reflect.Value, prefix string) (*client.Node, error) {
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	node := client.Node{}

	switch field.Kind() {
	case reflect.Struct:

		if val, ok := field.Interface().(time.Time); ok {
			node.Key = prefix
			node.Value = strconv.FormatInt(val.Unix(), 10)
			return &node, nil
		}

		node.Key = prefix
		node.Dir = true

		for i := 0; i < field.NumField(); i++ {
			subfield := field.Field(i)
			subfieldType := field.Type().Field(i)

			path := normalizeTag(subfieldType.Name)
			if len(path) == 0 {
				continue
			}
			path = prefix + "/" + path

			childNode, err := createNode(subfield, path)
			if err != nil {
				return nil, err
			}
			node.Nodes = append(node.Nodes, childNode)

		}

	case reflect.Map:
		node.Key = prefix
		node.Dir = true

		for _, key := range field.MapKeys() {
			value := field.MapIndex(key)
			path := prefix + "/" + key.String()

			switch value.Kind() {
			case reflect.Struct:
				childNode, err := createNode(value, path)
				if err != nil {
					return nil, err
				}
				node.Nodes = append(node.Nodes, childNode)
			case reflect.String:
				node.Key = path
				node.Value = value.String()
			}
		}

	case reflect.Slice:
		node.Key = prefix
		node.Dir = true

		for i := 0; i < field.Len(); i++ {
			item := field.Index(i)

			if item.Kind() == reflect.Struct {
				path := fmt.Sprintf("%s/%d", prefix, i)

				childNode, err := createNode(item, path)
				if err != nil {
					return nil, err
				}
				node.Nodes = append(node.Nodes, childNode)

			} else {
				path := fmt.Sprintf("%s/%d", prefix, i)
				node.Key = path
				node.Value = item.String()
			}
		}

	case reflect.String:
		value := field.Interface().(string)
		node.Key = prefix
		node.Value = value

	case reflect.Int:
		value, ok := field.Interface().(int)
		if !ok {
			forceCast := int(value)
			node.Key = prefix
			node.Value = strconv.Itoa(forceCast)
			break
		}
		node.Key = prefix
		node.Value = strconv.Itoa(value)

	case reflect.Int64:
		value := field.Interface().(int64)
		node.Key = prefix
		node.Value = strconv.FormatInt(value, 10)

	case reflect.Bool:
		value := field.Interface().(bool)

		var valueStr string
		if value {
			valueStr = "true"
		} else {
			valueStr = "false"
		}

		node.Key = prefix
		node.Value = valueStr
	}

	return &node, nil
}

// normalizeTag removes the slash from the beggining or end of the tag name and replace the other
// slashs with hyphens. The idea is to limit the hierarchy to the configuration structure
func normalizeTag(tag string) string {
	for strings.HasPrefix(tag, "/") {
		tag = strings.TrimPrefix(tag, "/")
	}

	for strings.HasSuffix(tag, "/") {
		tag = strings.TrimSuffix(tag, "/")
	}

	return strings.Replace(tag, "/", "-", -1)
}
