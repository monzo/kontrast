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

func (d Delta) Key() string {
	if d.SourceItem.Key != "" {
		return d.SourceItem.Key
	}
	return d.ServerItem.Key
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
