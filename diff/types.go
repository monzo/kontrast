package diff

import "github.com/milesbxf/petrel/k8s"

type Item struct {
	Key   string
	Value interface{}
}

type Delta struct {
	SourceItem Item
	ServerItem Item
}

type Diff interface {
	Deltas() []Delta
	Pretty() string
}

type DiffMeta struct {
	Resource *k8s.Resource
}

type EmptyDiff struct {
	DiffMeta
}

type ChangesPresentDiff struct {
	DiffMeta
	deltas []Delta
}

type NotPresentOnServerDiff struct {
	DiffMeta
}
