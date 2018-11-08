package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monzo/kryp/pkg/diff"
	"github.com/monzo/kryp/pkg/k8s"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

type DiffManager struct {
	LastRun *DiffRun
	LastErr error
	*k8s.ResourceHelper
}

var (
	currentDiffsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "kryp_current_diffs",
		Help: "Number of diffs between manifests and cluster",
	})
)

func (dm *DiffManager) DiffRun(path string) (*DiffRun, error) {
	d := &DiffRun{
		Time: time.Now(),
		Path: path,
	}

	numDiffs := 0
	err := filepath.Walk(path, func(fp string, fi os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil // not a file.  ignore.
		}

		if !strings.HasSuffix(fi.Name(), ".yaml") {
			fmt.Printf("Ignoring %s as it doesn't end in .yaml\n", fp)
			return nil
		}
		f := dm.processFile(fp)
		numDiffs += f.DiffResult.NumDiffs
		d.Files = append(d.Files, f)
		return nil
	})
	currentDiffsGauge.Set(float64(numDiffs))
	d.DiffResult = DiffFromNumber(numDiffs)
	dm.LastRun = d
	dm.LastErr = err
	return d, err
}

func (dm *DiffManager) processFile(path string) File {

	k8sResources, err := dm.ResourceHelper.NewResourcesFromFilename(path)

	if err != nil {
		log.Error("Error getting resources: %v\n", err)
		return File{
			Name:       path,
			DiffResult: ErrorDiffStatus(err.Error()),
		}
	}

	resources := []Resource{}
	numDiffs := 0
	for _, k8sr := range k8sResources {
		r := dm.processResource(k8sr)
		numDiffs += r.DiffResult.NumDiffs
		resources = append(resources, r)
	}

	return File{
		Name:       path,
		DiffResult: DiffFromNumber(numDiffs),
		Resources:  resources,
	}
}

func (dm *DiffManager) processResource(k8sr *k8s.Resource) Resource {
	gvk := k8sr.Object.GetObjectKind().GroupVersionKind()
	r := Resource{
		Name:             k8sr.Name,
		Namespace:        k8sr.Namespace,
		GroupVersionKind: fmt.Sprintf("%s.%s", gvk.Version, gvk.Kind),
	}

	d, err := diff.GetDiffsForResource(k8sr, dm.ResourceHelper)

	if err != nil {
		log.Errorf("Error getting resource: %v\n", err)
		r.DiffResult = ErrorDiffStatus(err.Error())
		return r
	}

	switch d.(type) {
	case diff.NotPresentOnServerDiff:
		r.IsNewResource = true
		r.DiffResult.Status = New
		r.DiffResult.NumDiffs = 1
	case diff.ChangesPresentDiff:
		r.DiffResult.NumDiffs = len(d.Deltas())
		if len(d.Deltas()) > 0 {
			r.DiffResult.Status = DiffPresent
		} else {
			r.DiffResult.Status = Clean
		}
	}

	for _, delta := range d.Deltas() {
		r.Diffs = append(r.Diffs, DiffFromDelta(delta))
	}

	return r
}

func DiffFromDelta(delta diff.Delta) Diff {
	return Diff{
		Key:   delta.SourceItem.Key,
		Left:  strOrRepr(delta.SourceItem.Value),
		Right: strOrRepr(delta.ServerItem.Value),
	}
}

func NewDiffManager(config *rest.Config) (*DiffManager, error) {
	helper, err := k8s.NewResourceHelperWithDefaults(config)
	prometheus.MustRegister(currentDiffsGauge)
	if err != nil {
		return &DiffManager{}, err
	}
	return &DiffManager{
		ResourceHelper: helper,
	}, nil
}

func strOrRepr(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}
