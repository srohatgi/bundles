package bundles

import (
	"fmt"
	"github.com/mohae/deepcopy"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type BundleFile struct {
	Contents     []byte
	Services     map[string]Service
	BaseServices map[string]int
}

type composer struct {
	Version struct {
		Services yaml.MapSlice
	}
}

type Service struct {
	Image       string        `yaml:""`
	Environment yaml.MapSlice `yaml:""`
	Ports       []int         `yaml:""`
	Labels      yaml.MapSlice `yaml:""`
	DependsOn   []string      `yaml:"depends_on,omitempty"`
}

func NewBundleFile(fileName string) (*BundleFile, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	b := BundleFile{Contents: bytes, Services: map[string]Service{}, BaseServices: map[string]int{}}

	c := composer{}
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return nil, err
	}

	svcNamePattern := regexp.MustCompile(`^([a-z]+)-[0-9]+$`)

	for _, v := range c.Version.Services {

		svcName, ok := v.Key.(string)
		if !ok {
			return nil, fmt.Errorf("unable to convert from serviceName(%v) to string", svcName)
		}

		matches := svcNamePattern.FindSubmatch([]byte(svcName))
		if len(matches) == 2 {
			b.BaseServices[string(matches[1])]++
		}

		svcYaml, err := yaml.Marshal(v.Value)
		if err != nil {
			panic(err)
		}

		svc := Service{}
		err = yaml.Unmarshal(svcYaml, &svc)
		if err != nil {
			panic(err)
		}

		b.Services[svcName] = svc
	}

	return &b, nil
}

func (b *BundleFile) Scale(serviceName string, count int) (*BundleFile, error) {

	if _, found := b.BaseServices[serviceName]; !found {
		return nil, fmt.Errorf("unknown service: %s", serviceName)
	}

	if b.BaseServices[serviceName] > count {
		return nil, fmt.Errorf("desired count:%d cannot be less than:%d", count, b.BaseServices[serviceName])
	}

	sb := &BundleFile{Contents: []byte{}, Services: map[string]Service{}, BaseServices: b.BaseServices}

	c := composer{}
	//fmt.Printf("composer: %v\n", c)

	// supported patterns:
	// <svc>-<digits>
	// <svc>-<digits>:<port>
	// <svc>-<digits>:<port>,...
	// <svc>-<digits>,...

	created := false
	for name, service := range b.Services {

		if strings.HasPrefix(name, serviceName) && created == false {
			node := deepcopy.Copy(service).(Service)

			nodeName := fmt.Sprintf("%s-%d", serviceName, sb.BaseServices[serviceName])
			sb.BaseServices[serviceName]++
			// check & fix for environment

			// check & fix for depends_on

			c.Version.Services = append(c.Version.Services, yaml.MapItem{Key: nodeName, Value: node})
		} else {
			copy := deepcopy.Copy(service).(Service)

			// check & fix for environment

			// check & fix for depends_on
			c.Version.Services = append(c.Version.Services, yaml.MapItem{Key: name, Value: copy})
		}

	}

	contents, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	sb.Contents = contents

	return sb, nil
}
