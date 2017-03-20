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

type Service struct {
	Image       string        `yaml:""`
	Environment yaml.MapSlice `yaml:""`
	Ports       []int         `yaml:""`
	Labels      yaml.MapSlice `yaml:""`
	DependsOn   []string      `yaml:"depends_on,omitempty"`
}

type composer struct {
	Version struct {
		Services yaml.MapSlice
	}
}

type replacer struct {
	pat   *regexp.Regexp
	value string
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

	return ParseBundleFile(bytes)
}

func ParseBundleFile(bytes []byte) (*BundleFile, error) {
	
	b := BundleFile{Contents: bytes, Services: map[string]Service{}, BaseServices: map[string]int{}}

	c := composer{}
	err := yaml.Unmarshal(bytes, &c)
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

	// a suitable cloner for the new service node(s)
	clonerName := fmt.Sprintf("%s-%d", serviceName, sb.BaseServices[serviceName])
	replacers := buildReplacerPatterns(serviceName, clonerName, b.Services[clonerName], b.BaseServices[serviceName])
	created := false
	sb.BaseServices[serviceName]++
	nodeName := fmt.Sprintf("%s-%d", serviceName, sb.BaseServices[serviceName])

	for name, service := range b.Services {

		if name == clonerName && !created {
			created = true
			node := deepcopy.Copy(service).(Service)

			// check & fix for environment
			environmentFix(node, replacers)

			c.Version.Services = append(c.Version.Services, yaml.MapItem{Key: nodeName, Value: node})
		}

		copy := deepcopy.Copy(service).(Service)

		// check & fix for environment
		environmentFix(copy, replacers)

		// check & fix for depends_on
		for _, d := range copy.DependsOn {
			if d == clonerName {
				copy.DependsOn = append(copy.DependsOn, nodeName)
				break
			}
		}

		c.Version.Services = append(c.Version.Services, yaml.MapItem{Key: name, Value: copy})
	}

	contents, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	sb.Contents = contents

	return sb, nil
}

func buildReplacerPatterns(base, name string, service Service, count int) []replacer {
	m := []replacer{}
	newName := fmt.Sprintf("%s-%d", base, count+1)
	nodesNow := []string{}
	nodesPorts := map[int][]string{}

	for i := 1; i <= count+1; i++ {
		nodesNow = append(nodesNow, fmt.Sprintf("%s-%d", base, i))
		for p := range service.Ports {
			nodesPorts[p] = append(nodesPorts[p], fmt.Sprintf("%s-%d:%d", base, i, p))
		}
	}

	var pat *regexp.Regexp

	// kafka-1
	pat = regexp.MustCompile(name)
	m = append(m, replacer{pat, newName})

	// kafka-1,kafka-2
	pat = regexp.MustCompile(fmt.Sprintf("%s-[0-9]+,?", base))
	m = append(m, replacer{pat, strings.Join(nodesNow, ",")})

	for p := range service.Ports {
		// kafka-1:9092,kafka-2:9092
		pat = regexp.MustCompile(fmt.Sprintf("%s-[0-9]+:(%d),?", base, p))
		m = append(m, replacer{pat, strings.Join(nodesPorts[p], ",")})
	}

	return m
}

func environmentFix(copy Service, replacers []replacer) {
	for i, v := range copy.Environment {
		if s, ok := v.Value.(string); ok {

			for _, r := range replacers {
				fmt.Printf("******s: %s r.pat: %v r.value: %s find: %s\n", s, r.pat, r.value, r.pat.Find([]byte(s)))

				if string(r.pat.Find([]byte(s))) == s {
					fmt.Printf("--------s: %s r.pat: %v r.value: %s\n", s, r.pat, r.value)
					copy.Environment[i].Value = r.value
					break
				}
			}

		}
	}
}
