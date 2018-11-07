package diff

import (
	"log"

	"github.com/monzo/kryp/pkg/k8s"
	//_ "k8s.io/kubernetes/pkg/master"
)

// GetDiffsForResource takes a resource, and uses to generate a local Kubernetes object
// which it compares to the equivalent object fetched from the cluster
func GetDiffsForResource(resource *k8s.Resource, helper *k8s.ResourceHelper) (Diff, error) {

	// Create a Kubernetes object from the file
	defaultedObj := k8s.GetWithDefaults(resource.Object)
	meta := DiffMeta{Resource: resource}

	// Get the Kubernetes object from the server
	serverObj, err := resource.Get()
	if err != nil {
		if k8s.IsNotFoundError(err) {
			return NotPresentOnServerDiff{DiffMeta: meta}, nil
		}

		log.Printf("Error getting object: %v", err)
		return ChangesPresentDiff{}, err
	}

	// Compare the File and Server Objects
	deltas, err := calculateDiff(defaultedObj, serverObj)
	if err != nil {
		log.Printf("Error calculating deltas: %v", err)
		return ChangesPresentDiff{}, err
	}

	// Some deltas are to be expected, so we filter them
	filtered := deltas
	for _, f := range []DeltaFilter{MetadataFilter} {
		filtered = f(filtered)
	}

	log.Printf("Found %d deltas", len(deltas))
	return ChangesPresentDiff{DiffMeta: meta, deltas: filtered}, nil
}

var empty = struct{}{}

func (d ChangesPresentDiff) Deltas() []Delta                     { return d.deltas }
func (d NotPresentOnServerDiff) Pretty(colorEnabled bool) string { return "" }
func (d NotPresentOnServerDiff) Deltas() []Delta                 { return []Delta{} }
