package diff

import (
	"github.com/milesbxf/petrel/k8s"

	_ "k8s.io/kubernetes/pkg/master"
)

func GetFileDiff(filename string, helper *k8s.ResourceHelper) (Diff, error) {
	resource, err := helper.NewResourceFromFilename(filename)
	if err != nil {
		return EmptyDiff{}, err
	}
	defaultedObj := k8s.GetWithDefaults(resource.Object)
	meta := DiffMeta{Resource: resource}

	serverObj, err := resource.Get()
	if k8s.IsNotFoundError(err) {
		return NotPresentOnServerDiff{DiffMeta: meta}, nil
	}
	deltas, err := calculateDiff(defaultedObj, serverObj)
	if err != nil {
		return EmptyDiff{}, err
	}
	if len(deltas) == 0 {
		return EmptyDiff{DiffMeta: meta}, nil
	} else {
		return ChangesPresentDiff{DiffMeta: meta, deltas: deltas}, nil
	}
}

var empty = struct{}{}

type Diff interface {
	Deltas() []Delta
}

type DiffMeta struct {
	Resource *k8s.Resource
}

type EmptyDiff struct {
	DiffMeta
}

func (ed EmptyDiff) Deltas() []Delta {
	return []Delta{}
}

type ChangesPresentDiff struct {
	DiffMeta
	deltas []Delta
}

func (d ChangesPresentDiff) Deltas() []Delta {
	return d.deltas
}

type NotPresentOnServerDiff struct {
	DiffMeta
}

func (d NotPresentOnServerDiff) Deltas() []Delta {
	return []Delta{}
}
