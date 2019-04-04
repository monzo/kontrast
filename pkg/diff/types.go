package diff

import "github.com/monzo/kontrast/pkg/k8s"

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
	Pretty(colorEnabled bool) string
}

type DiffMeta struct {
	Resource *k8s.Resource
}

type ChangesPresentDiff struct {
	DiffMeta
	deltas []Delta
}

type NotPresentOnServerDiff struct {
	DiffMeta
}
