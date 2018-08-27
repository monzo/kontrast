package diff

import (
	"log"

	"github.com/milesbxf/petrel/k8s"

	_ "k8s.io/kubernetes/pkg/master"
)

func GetFileDiff(filename string, helper *k8s.ResourceHelper) (Diff, error) {
	resource, err := helper.NewResourceFromFilename(filename)
	if err != nil {
		log.Printf("Error getting resource: %v", err)
		return EmptyDiff{}, err
	}
	log.Printf("Setting defaults for object %v", resource.Object)
	defaultedObj := k8s.GetWithDefaults(resource.Object)
	meta := DiffMeta{Resource: resource}

	serverObj, err := resource.Get()
	if k8s.IsNotFoundError(err) {
		return NotPresentOnServerDiff{DiffMeta: meta}, nil
	}
	deltas, err := calculateDiff(defaultedObj, serverObj)
	if err != nil {
		log.Printf("Error calculating deltas: %v", err)
		return EmptyDiff{}, err
	}
	log.Printf("Found %d deltas", len(deltas))
	if len(deltas) == 0 {
		return EmptyDiff{DiffMeta: meta}, nil
	}
	return ChangesPresentDiff{DiffMeta: meta, deltas: deltas}, nil
}

var empty = struct{}{}

func (d EmptyDiff) Pretty() string   { return "" }
func (ed EmptyDiff) Deltas() []Delta { return []Delta{} }

func (d ChangesPresentDiff) Deltas() []Delta { return d.deltas }

func (d NotPresentOnServerDiff) Pretty() string  { return "" }
func (d NotPresentOnServerDiff) Deltas() []Delta { return []Delta{} }
