package diff

import (
	"log"

	"github.com/milesbxf/kryp/pkg/k8s"

	_ "k8s.io/kubernetes/pkg/master"
)

func GetDiffsForResource(resource *k8s.Resource, helper *k8s.ResourceHelper) (Diff, error) {
	log.Printf("Setting defaults for object %v", resource.Object)
	defaultedObj := k8s.GetWithDefaults(resource.Object)
	meta := DiffMeta{Resource: resource}

	serverObj, err := resource.Get()
	if k8s.IsNotFoundError(err) {
		return NotPresentOnServerDiff{DiffMeta: meta}, nil
	}
	if err != nil {
		log.Printf("Error getting object: %v", err)
		return ChangesPresentDiff{}, err
	}

	deltas, err := calculateDiff(defaultedObj, serverObj)
	if err != nil {
		log.Printf("Error calculating deltas: %v", err)
		return ChangesPresentDiff{}, err
	}

	filtered := deltas
	for _, f := range []DeltaFilter{MetadataFilter} {
		filtered = f(filtered)
	}

	log.Printf("Found %d deltas", len(deltas))

	return ChangesPresentDiff{DiffMeta: meta, deltas: filtered}, nil
}

var empty = struct{}{}

func (d ChangesPresentDiff) Deltas() []Delta { return d.deltas }

func (d NotPresentOnServerDiff) Pretty(colorEnabled bool) string { return "" }
func (d NotPresentOnServerDiff) Deltas() []Delta                 { return []Delta{} }
