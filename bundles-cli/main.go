package main

import (
	"flag"
	"fmt"
	"github.com/srohatgi/bundles"
)

func main() {
	var yamlFile string
	flag.StringVar(&yamlFile, "file", "bundle-compose.yaml", "docker-compose like file for reading")
	flag.Parse()

	b, err := bundles.NewBundleFile(yamlFile)
	if err != nil {
		panic(fmt.Errorf("unable to read yamlFile: %s, error: %v", yamlFile, err))
	}

	//fmt.Printf("parsed service tree: %v\n", b.Services)

	fmt.Printf("key services: %v\n", b.BaseServices)

	fmt.Printf("scaling kafka to 3\n")
	sb, err := b.Scale("kafka", 3)
	if err != nil {
		panic(err)
	}

	fmt.Printf("scaled bundle: \n%s\n", sb.Contents)
}
