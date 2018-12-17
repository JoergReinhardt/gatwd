package main

import (
	"fmt"
)

func main() {

	//path := "./dump"
	path := "/home/j/.config/kubextract/cluster-backup/"
	//path := "/home/j/.config/kubextract/cluster-backup/resources/statefulsets.apps/namespaces/monitoring/prometheus-k8s.json"
	//path := "/home/j/.config/kubextract/idx-manifests.libsonnet"
	//path := "/home/j/.config/kubextract/tokens-unsorted.libsonnet"

	//////////////////////////////
	//rd := NewRecursiveReader(path)
	/////////////////////////////////////////////////
	//fmt.Println(rd.Paths())
	fmt.Println(path)
}
