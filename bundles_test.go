package bundles

import "testing"

func TestNewBundleFile(t *testing.T) {
	bundle := `
version: '2'
services:
  zookeeper-1:
    image: 801351377084.dkr.ecr.us-west-1.amazonaws.com/zookeeper-exhibitor
    environment:
      S3_BUCKET: hello-world
      S3_PREFIX: dummy
    ports:
      - 2181
    labels:
      avanti.service.count: "1"
`

	b, err := NewBundleFile([]byte(bundle))
	if err != nil {
		t.Fatalf("unable to create bundle, error: %v", err)
	}

	if len(b.BaseServices) != 1 {
		t.Fatal("length of BaseServices is not 1")
	}

	if _, found := b.BaseServices["zookeeper"]; !found {
		t.Fatal("unable to parse BaseService")
	}

	if b.BaseServices["zookeeper"] != 1 {
		t.Fatal("there should be only one instance of zooKeeper service")
	}

	if _, found := b.Services["zookeeper-1"]; !found {
		t.Fatal("unable to parse service zookeeper-1")
	}
}

func TestBundleFile_ScaleSimple(t *testing.T) {
	bundle := `
version: '2'
services:
  zookeeper-1:
    image: 801351377084.dkr.ecr.us-west-1.amazonaws.com/zookeeper-exhibitor
    environment:
      S3_BUCKET: hello-world
      S3_PREFIX: dummy
    ports:
      - 2181
    labels:
      avanti.service.count: "1"
`

	b, err := NewBundleFile([]byte(bundle))
	if err != nil {
		t.Fatalf("unable to create bundle file object: %v", err)
	}

	nb, err := b.Scale("zookeeper", 1)
	if err == nil {
		t.Fatalf("should not be able to scale bundle file object to same size: %v", err)
	}

	nb, err = b.Scale("zookeeper", 2)
	if err != nil {
		t.Fatal("unable to scale to zookeeper size = 2")
	}

	t.Logf("b.Services: %v nb.Services: %v", b.Services, nb.Services)

//	if len(nb.Services) != 2 {
//		t.Fatalf("after scaling, nb.services: %v", nb.Services)
//	}
}
