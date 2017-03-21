package main

import (
	"flag"
	"fmt"
	"github.com/srohatgi/bundles"
	"io/ioutil"
	"os"
	logger "github.com/Sirupsen/logrus"
)

func init() {
	logger.SetFormatter(&logger.TextFormatter{})
	logger.SetLevel(logger.DebugLevel)
}

func main() {
	var yamlFile string
	var service string
	var count int64
	flag.StringVar(&yamlFile, "file", "bundle-compose.yaml", "docker-compose like file that defines a reusable bundle")
	flag.StringVar(&service, "service", "zookeeper", "specify bundle service for scaling")
	flag.Int64Var(&count, "count", 3, "specify service instance count")
	flag.Parse()

	bt, err := readFile(yamlFile)
	if err != nil {
		panic(fmt.Errorf("unable to read yamlFile: %s, error: %v", yamlFile, err))
	}

	bundles.SetOutput(logger.New().Writer())

	b, err := bundles.NewBundleFile(bt)

	logger.Debugf("key services: %v", b.BaseServices)

	logger.Infof("scaling %s to %d", service, count)

	sb, err := b.Scale(service, count)
	if err != nil {
		panic(err)
	}

	logger.Debugf("scaled bundle: %s", sb.Contents)
}

func readFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
