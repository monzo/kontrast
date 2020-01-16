package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	objectLabel = "object"
	nsLabel     = "object_ns"
)

var (
	currentDiffsGauge = prometheus.NewDesc(
		"kontrast_current_diffs",
		"Number of diffs between manifests and cluster",
		[]string{objectLabel, nsLabel}, nil)
)

type labelSet struct {
	Kind      string
	Name      string
	Namespace string
}

// KontrastCollector is here to satisfy the Prometheus Collector interface
type KontrastCollector struct {
	manager *DiffManager
}

func NewKontrastCollector(manager *DiffManager) *KontrastCollector {
	return &KontrastCollector{
		manager: manager,
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (c *KontrastCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- currentDiffsGauge
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (c *KontrastCollector) Collect(ch chan<- prometheus.Metric) {
	resources := map[labelSet]float64{}
	for _, file := range c.manager.GetDiffFiles() {
		for _, resource := range file.Resources {
			if resource.DiffResult.Status == DiffPresent {
				ls := labelSet{resource.Kind, resource.Name, resource.Namespace}
				resources[ls] = resources[ls] + 1
			}
		}
	}

	for resource, _ := range resources {
		objectLabel := fmt.Sprintf("%s/%s", resource.Kind, resource.Name)
		ch <- prometheus.MustNewConstMetric(currentDiffsGauge,
			prometheus.GaugeValue, 1.0,
			objectLabel, resource.Namespace)
	}
}
